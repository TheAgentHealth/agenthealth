import assert from 'node:assert/strict';
import express from 'express';
import { agentHealth } from './agenthealth-ref.js';

let probeCalls = 0;
const app = express();
app.use(agentHealth(() => ({ name: 'reference', live: true, ready: true,
  capabilities: ['health'], dependencies: [] }),
  async task => { probeCalls++; return { completed: true, success: task.text === 'health' && !task.downstream }; }));
const server = app.listen(0, '127.0.0.1');
await new Promise<void>(resolve => server.once('listening', resolve));
const address = server.address();
assert(address && typeof address !== 'string');
const url = `http://127.0.0.1:${address.port}/health`;
try {
  assert.equal((await fetch(url, { method: 'HEAD' })).status, 200);
  const doc = await (await fetch(url)).json();
  assert.equal(doc.version, 'v1');
  assert.deepEqual(doc.dependencies, []);
  async function post(body: unknown) {
    return fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
  }
  assert.deepEqual(await (await post({ safe: true, text: 'health' })).json(), { completed: true, success: true });
  assert.equal((await post({ safe: false, text: 'health' })).status, 400);
  assert.equal((await post({ safe: true, text: 'health', command: 'sh' })).status, 400);
  for (const [text, downstream, status] of [
    ['x'.repeat(65536), '', 200],
    ['\u0000'.repeat(65536), '😀'.repeat(1024), 200],
    ['😀'.repeat(16384), '', 200],
    ['x'.repeat(65537), '', 400],
    ['😀'.repeat(16384) + 'x', '', 400],
    ['health', '😀'.repeat(1025), 400],
    ['health', ' ', 400],
    [' ', '', 400],
  ] as const) {
    const before = probeCalls;
    assert.equal((await post({ safe: true, text, downstream })).status, status);
    assert.equal(probeCalls - before, status === 200 ? 1 : 0);
  }
  const before = probeCalls;
  assert.equal((await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: ' '.repeat(512 * 1024 + 1) })).status, 413);
  assert.equal(probeCalls, before);
} finally { await new Promise<void>((resolve, reject) => server.close(err => err ? reject(err) : resolve())); }
