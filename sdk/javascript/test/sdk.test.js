import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, rm, readFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { createServer } from 'node:http';
import { Configuration, loadConfig, parseReport, check, version, healthStates, ConfigurationError,
  CompatibilityError, BinaryNotFoundError, OutputError, ExecutionError, CheckTimeoutError, CheckAbortedError } from '../index.js';
const config = { version: 'v1', targets: [{ name: 'api', type: 'http', endpoint: 'http://127.0.0.1:1', checks: ['reachability'] }] };
const result = { spec_version: 'v1', target: { name: 'root', type: 'agent', id: 'root' }, status: 'HEALTHY',
  checks: { reachability: { status: 'HEALTHY', code: 'reachable', steps: { http: 'HEALTHY' } } }, dependencies: [] };
async function fixture(t, body) {
  if (process.platform === 'win32') { t.skip('Executable script fixtures use POSIX shebangs; real CLI tests cover Windows'); return; }
  const dir = await mkdtemp(join(tmpdir(), 'agenthealth-test-'));
  t.after(() => rm(dir, { recursive: true, force: true }));
  const binary = join(dir, 'fake-cli');
  await writeFile(binary, `#!${process.execPath}\nif(process.argv[2]==='version'){console.log('agenthealth v${version}')}else{${body}}`, { mode: 0o700 });
  return binary;
}
test('configuration validates without leaking configured secrets', () => {
  assert.throws(() => new Configuration({ secret: 'credential-value' }), ConfigurationError);
  assert.throws(() => new Configuration({ ...config, unexpected: 'credential-value' }), /Invalid AgentHealth configuration/);
  const c = new Configuration(config); c.toJSON().targets[0].name = 'changed';
  assert.equal(c.toJSON().targets[0].name, 'api');
  assert.throws(() => new Configuration({ ...config, defaults: { timeout: NaN } }), ConfigurationError);
});
test('bounded YAML loading rejects duplicate keys and preserves cwd', async t => {
  const dir = await mkdtemp(join(tmpdir(), 'agenthealth-yaml-')); t.after(() => rm(dir, { recursive: true, force: true }));
  const path = join(dir, 'config.yaml');
  await writeFile(path, JSON.stringify(config)); assert.equal((await loadConfig(path)).baseDir, process.cwd());
  await writeFile(path, 'version: v1\nversion: v1\ntargets: []'); await assert.rejects(loadConfig(path), ConfigurationError);
  await writeFile(path, 'x'.repeat(4 * 1024 * 1024 + 1)); await assert.rejects(loadConfig(path), ConfigurationError);
});
test('all health exit codes resolve and graph evidence survives', () => {
  for (let code = 0; code < 6; code++) {
    const document = { ...result, status: healthStates[code], dependencies: [{ ...result, relationship: 'path', critical: false }] };
    const report = parseReport(JSON.stringify(document), code);
    assert.equal(report.status, healthStates[code]); assert.equal(report.results[0].dependencies[0].relationship, 'path');
  }
  assert.equal(parseReport(JSON.stringify({ spec_version: 'v1', results: [result, { ...result, status: 'UNKNOWN' }] }), 5).results.length, 2);
  for (const [output, code] of [['secret-error', 0], [JSON.stringify(result), 1], [JSON.stringify({ ...result, spec_version: 'v2' }), 0]]) {
    assert.throws(() => parseReport(output, code), OutputError);
  }
});
test('schemas remain identical to the canonical contracts', async () => {
  for (const name of ['result', 'configuration']) assert.deepEqual(
    await readFile(new URL(`../schemas/${name}.schema.json`, import.meta.url)),
    await readFile(new URL(`../../../spec/schemas/${name}.schema.json`, import.meta.url)));
});
test('missing CLI rejects with sanitized error', async () => {
  await assert.rejects(check(config, { binary: '/missing-agenthealth-cli' }), BinaryNotFoundError);
});
test('health failure is a report, subprocess failure is an exception', async t => {
  const binary = await fixture(t, `console.log(${JSON.stringify(JSON.stringify({ ...result, status: 'UNREACHABLE' }))});process.exitCode=3;`);
  if (!binary) return;
  assert.equal((await check(config, { binary })).status, 'UNREACHABLE');
  await writeFile(binary, `#!${process.execPath}\nconsole.log('agenthealth v0.0.0')\n`);
  await assert.rejects(check(config, { binary }), CompatibilityError);
});
test('malformed output, stderr flood and execution failures are bounded', async t => {
  for (const [body, error] of [["console.log('secret');", OutputError], ["process.stderr.write('s'.repeat(70000));", OutputError], ['process.exitCode=6;', ExecutionError]]) {
    const binary = await fixture(t, body); if (!binary) return;
    await assert.rejects(check(config, { binary }), error);
  }
});
test('deadline and AbortSignal stop execution', async t => {
  const binary = await fixture(t, 'setInterval(()=>{},1000);'); if (!binary) return;
  await assert.rejects(check(config, { binary, timeout: 150 }), CheckTimeoutError);
  const controller = new AbortController();
  const pending = check(config, { binary, signal: controller.signal });
  setTimeout(() => controller.abort(), 150);
  await assert.rejects(pending, CheckAbortedError);
  await assert.rejects(check(config, { binary, signal: controller.signal }), CheckAbortedError);
});
test('real CLI handles healthy HTTP, batch, dependency and failed health', async t => {
  const binary = process.env.AGENTHEALTH_SDK_BINARY;
  if (!binary) { t.skip('Set AGENTHEALTH_SDK_BINARY to the matching compiled CLI'); return; }
  const server = createServer((req, res) => { res.writeHead(200); res.end('ok'); });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  t.after(() => new Promise(resolve => server.close(resolve)));
  const endpoint = `http://127.0.0.1:${server.address().port}`;
  const target = { name: 'api', type: 'http', endpoint, checks: ['reachability'] };
  assert.equal((await check({ version: 'v1', targets: [target] }, { binary })).exitCode, 0);
  const batch = await check({ version: 'v1', targets: [target, { ...target, name: 'other' }] }, { binary });
  assert.equal(batch.results.length, 2);
  const graph = await check({ version: 'v1', targets: [
    { ...target, id: 'root', dependencies: [{ ref: 'child', relationship: 'supporting', critical: true }] },
    { ...target, name: 'child', id: 'child' }
  ] }, { binary });
  assert.equal(graph.results[0].dependencies[0].target.id, 'child');
  assert.equal(graph.results[0].dependencies[0].relationship, 'supporting');
  const unhealthy = await check(config, { binary }); assert.equal(unhealthy.status, 'UNREACHABLE');
});
test('POSIX cancellation cleans descendants and private config', async t => {
  const binary = await fixture(t, `
const fs=require('node:fs');const cp=require('node:child_process');
const descendant=cp.spawn(process.execPath,['-e','setInterval(()=>{},1000)'],{stdio:'ignore'});
fs.writeFileSync(process.env.MARKER,JSON.stringify({pid:descendant.pid,path:process.argv.at(-1),mode:fs.statSync(process.argv.at(-1)).mode&511}));
setInterval(()=>{},1000);
`);
  if (!binary) return;
  const marker = join(binary, '..', 'marker.json');
  const controller = new AbortController();
  const pending = check(config, { binary, environment: { MARKER: marker }, signal: controller.signal });
  // Attach rejection immediately while waiting for the subprocess marker.
  const rejection = assert.rejects(pending, CheckAbortedError);
  let recorded;
  for (let n=0; n<100; n++) {
    try { recorded = JSON.parse(await readFile(marker, 'utf8')); break; } catch {}
    await new Promise(resolve => setTimeout(resolve, 20));
  }
  controller.abort(); await rejection;
  assert.ok(recorded); assert.equal(recorded.mode, 0o600);
  await assert.rejects(readFile(recorded.path), { code: 'ENOENT' });
  // Linux can retain a killed child briefly as a zombie until init reaps it.
  let alive = true;
  for (let n=0; n<100; n++) {
    try {
      process.kill(recorded.pid, 0);
      if (process.platform === 'linux' && (await readFile(`/proc/${recorded.pid}/stat`, 'utf8')).split(' ')[2] === 'Z') { alive=false; break; }
    } catch { alive=false; break; }
    await new Promise(resolve => setTimeout(resolve, 20));
  }
  assert.equal(alive, false);
});
