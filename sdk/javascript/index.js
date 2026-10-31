/** Node.js-only wrapper. The Go reference CLI owns health semantics. */
import { spawn } from 'node:child_process';
import { mkdtemp, readFile, writeFile, rm, open } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve, join } from 'node:path';
import { performance } from 'node:perf_hooks';
import Ajv2020 from 'ajv/dist/2020.js';
import { parseDocument } from 'yaml';

export const version = '0.18.0';
export const healthStates = Object.freeze(['HEALTHY', 'DEGRADED', 'UNHEALTHY', 'UNREACHABLE', 'MISCONFIGURED', 'UNKNOWN']);
export class AgentHealthError extends Error {
  constructor(message) { super(message); this.name = new.target.name; }
}
export class ConfigurationError extends AgentHealthError {}
export class BinaryNotFoundError extends AgentHealthError {}
export class CompatibilityError extends AgentHealthError {}
export class ExecutionError extends AgentHealthError {}
export class OutputError extends AgentHealthError {}
export class CheckTimeoutError extends AgentHealthError {}
export class CheckAbortedError extends AgentHealthError {}
const ajv = new Ajv2020({ strict: false, allErrors: false });
const validators = {};
for (const name of ['configuration', 'result']) {
  validators[name] = ajv.compile(JSON.parse(await readFile(new URL(`./schemas/${name}.schema.json`, import.meta.url), 'utf8')));
}
const MAX_CONFIG = 4 * 1024 * 1024;
const MAX_OUTPUT = 16 * 1024 * 1024;
const MAX_DIAGNOSTIC = 64 * 1024;
function serialize(data) {
  // Reject values JSON would silently discard or turn into null.
  return JSON.stringify(data, (_, value) => {
    if (value === undefined || typeof value === 'function' || typeof value === 'symbol' ||
        (typeof value === 'number' && !Number.isFinite(value))) throw new Error();
    return value;
  });
}
export class Configuration {
  #data;
  constructor(data, { baseDir = process.cwd() } = {}) {
    try {
      const text = serialize(data);
      if (Buffer.byteLength(text) > MAX_CONFIG) throw new Error();
      this.#data = JSON.parse(text);
      if (!validators.configuration(this.#data)) throw new Error();
      this.baseDir = resolve(baseDir);
      Object.freeze(this);
    } catch { throw new ConfigurationError('Invalid AgentHealth configuration'); }
  }
  toJSON() { return structuredClone(this.#data); }
}
export async function loadConfig(path, options) {
  try {
    const file = await open(path, 'r');
    let bytes;
    try {
      const buffer = Buffer.alloc(MAX_CONFIG + 1);
      let total = 0;
      while (total < buffer.length) {
        const { bytesRead } = await file.read(buffer, total, buffer.length - total, total);
        if (!bytesRead) break;
        total += bytesRead;
      }
      bytes = buffer.subarray(0, total);
    } finally { await file.close(); }
    if (bytes.length > MAX_CONFIG) throw new Error();
    const document = parseDocument(bytes.toString('utf8'), { uniqueKeys: true });
    if (document.errors.length) throw new Error();
    return new Configuration(document.toJS({ maxAliasCount: 50 }), options);
  } catch { throw new ConfigurationError('Unable to load AgentHealth configuration'); }
}
export function parseReport(output, exitCode) {
  try {
    if (Buffer.byteLength(output) > MAX_OUTPUT || !Number.isInteger(exitCode) || exitCode < 0 || exitCode > 5) throw new Error();
    const document = JSON.parse(output);
    if (!validators.result(document) || document.spec_version !== 'v1') throw new Error();
    const results = document.results ?? [document];
    if (Math.max(...results.map(result => healthStates.indexOf(result.status))) !== exitCode) throw new Error();
    return { results, exitCode, status: healthStates[exitCode], spec_version: 'v1' };
  } catch { throw new OutputError('Invalid AgentHealth result or exit-code contract'); }
}
async function run(binary, args, { cwd, environment, deadline, signal }) {
  if (signal?.aborted) throw new CheckAbortedError('AgentHealth check aborted');
  if (performance.now() >= deadline) throw new CheckTimeoutError('AgentHealth check deadline expired');
  return new Promise((accept, reject) => {
    let child;
    try { child = spawn(binary, args, { cwd, env: { ...process.env, ...environment }, shell: false,
      detached: process.platform !== 'win32', windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] });
    } catch { reject(new ExecutionError('Unable to start AgentHealth CLI')); return; }
    let chunks = [], size = 0, diagnosticSize = 0, failure, timer, killTimer;
    function kill(sig) {
      try {
        if (process.platform === 'win32') child.kill(sig);
        else process.kill(-child.pid, sig);
      } catch { /* The process/group has already exited. */ }
    }
    function stop(error) {
      failure ??= error;
      kill('SIGTERM');
      killTimer ??= setTimeout(() => kill('SIGKILL'), 1000);
    }
    const abort = () => stop(new CheckAbortedError('AgentHealth check aborted'));
    signal?.addEventListener('abort', abort, { once: true });
    // Close the abort race between the initial check and listener registration.
    if (signal?.aborted) abort();
    timer = setTimeout(() => stop(new CheckTimeoutError('AgentHealth check deadline expired')),
      Math.max(1, deadline - performance.now()));
    child.stdout.on('data', chunk => {
      size += chunk.length;
      if (size > MAX_OUTPUT) stop(new OutputError('AgentHealth output exceeded the SDK limit'));
      else if (!failure) chunks.push(chunk);
    });
    child.stderr.on('data', chunk => {
      diagnosticSize += chunk.length;
      if (diagnosticSize > MAX_DIAGNOSTIC) stop(new OutputError('AgentHealth output exceeded the SDK limit'));
    });
    child.on('error', () => { failure ??= new BinaryNotFoundError('Unable to start AgentHealth CLI'); });
    child.on('exit', () => {
      // Descendants may retain pipe descriptors after their parent exits.
      kill('SIGKILL');
    });
    child.on('close', code => {
      clearTimeout(timer); clearTimeout(killTimer);
      signal?.removeEventListener('abort', abort);
      kill('SIGKILL');
      if (failure) reject(failure);
      else accept({ code, output: Buffer.concat(chunks).toString('utf8') });
    });
  });
}
export class Client {
  constructor({ binary = 'agenthealth', timeout = 30000, environment = {} } = {}) {
    if (!Number.isFinite(timeout) || timeout <= 0 || timeout > 2147483647) throw new RangeError('timeout must be between 0 and 2147483647 milliseconds');
    if (typeof binary !== 'string' || !binary) throw new TypeError('binary must be an executable name or path');
    this.binary = /[/\\]/.test(binary) ? resolve(binary) : binary;
    this.timeout = timeout;
    this.environment = { ...environment };
  }
  async check(config, { signal } = {}) {
    const configuration = config instanceof Configuration ? config : typeof config === 'string' ? await loadConfig(config) : new Configuration(config);
    const options = { cwd: configuration.baseDir, environment: this.environment,
      deadline: performance.now() + this.timeout, signal };
    const cli = await run(this.binary, ['version'], { ...options, deadline: Math.min(options.deadline, performance.now() + 5000) });
    if (cli.code !== 0 || cli.output.trim() !== `agenthealth v${version}`) throw new CompatibilityError('AgentHealth CLI must match the SDK software version');
    let directory, report, failure;
    try {
      directory = await mkdtemp(join(tmpdir(), 'agenthealth-js-'));
      const path = join(directory, 'configuration.json');
      await writeFile(path, JSON.stringify(configuration.toJSON()), { mode: 0o600, flag: 'wx' });
      const { code, output } = await run(this.binary, ['check', '--format', 'json', '--', path], options);
      if (!Number.isInteger(code) || code < 0 || code > 5) throw new ExecutionError('AgentHealth CLI execution failed');
      report = parseReport(output, code);
    } catch (error) {
      failure = error instanceof AgentHealthError ? error : new ExecutionError('Unable to prepare AgentHealth CLI execution');
    } finally {
      if (directory) {
        try { await rm(directory, { recursive: true, force: true }); }
        catch { failure ??= new ExecutionError('Unable to clean up AgentHealth CLI execution'); }
      }
    }
    if (failure) throw failure;
    return report;
  }
}
export function check(config, options = {}) {
  return new Client(options).check(config, { signal: options.signal });
}
