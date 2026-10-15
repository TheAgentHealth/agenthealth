/** Install the actual tarball outside the source tree, then use its ESM/types. */
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { spawnSync } from 'node:child_process';
const directory = await mkdtemp(join(tmpdir(), 'agenthealth-installed-js-'));
function run(command, args) {
  const result = spawnSync(command, args, { cwd: directory, stdio: 'inherit',
    shell: process.platform === 'win32' && command === 'npm' });
  if (result.status !== 0) throw new Error(`Installed SDK check failed: ${command}`);
}
try {
  await writeFile(join(directory, 'package.json'), JSON.stringify({ private: true, type: 'module' }));
  run('npm', ['install', '--ignore-scripts', resolve(process.argv[2])]);
  await writeFile(join(directory, 'smoke.mjs'), `
import assert from 'node:assert/strict';
import { check, version, Configuration, BinaryNotFoundError } from '@theagenthealth/agenthealth';
assert.match(version, /^\\d+\\.\\d+\\.\\d+/);
const config = new Configuration({ version: 'v1', targets: [{ name: 'api', type: 'http', endpoint: 'http://127.0.0.1:1', checks: ['reachability'] }] });
await assert.rejects(check(config, {binary: '/missing-cli'}), BinaryNotFoundError);
if (process.env.AGENTHEALTH_SDK_BINARY) {
  const report = await check(config, {binary: process.env.AGENTHEALTH_SDK_BINARY});
  assert.equal(report.status, 'UNREACHABLE');
}
`);
  run(process.execPath, ['smoke.mjs']);
} finally { await rm(directory, { recursive: true, force: true }); }
