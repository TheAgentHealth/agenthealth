import assert from 'node:assert/strict';
import express from 'express';
import { agentHealth } from './agenthealth-ref.js';

const app = express();
app.use(agentHealth(() => ({ name: 'reference', live: true, ready: true,
  capabilities: ['health'], dependencies: [] }),
  async task => ({ completed: true, success: task.text === 'health' && !task.downstream })));
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
  assert.equal((await post({ safe: true, text: 'x'.repeat(65537) })).status, 413);
} finally { await new Promise<void>((resolve, reject) => server.close(err => err ? reject(err) : resolve())); }
