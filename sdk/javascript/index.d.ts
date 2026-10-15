export const version: string;
export type HealthStatus = 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY' | 'UNREACHABLE' | 'MISCONFIGURED' | 'UNKNOWN';
export const healthStates: readonly HealthStatus[];
export type Relationship = 'supporting' | 'downstream' | 'path' | 'gateway' | 'router';
export type TargetType = 'agent' | 'multi-agent' | 'a2a' | 'mcp' | 'tool' | 'model' | 'llm' | 'http' | 'api' | 'vector-store' | 'database' | 'runtime' | 'gateway' | 'router' | 'custom';
export interface CheckEntry {
  status: HealthStatus;
  message?: string;
  code?: string;
  steps?: Partial<Record<'dns' | 'tcp' | 'tls' | 'http', HealthStatus>>;
  [evidence: string]: unknown;
}
export interface HealthResult {
  target: { name: string; type: TargetType; id?: string; [evidence: string]: unknown };
  status: HealthStatus;
  checks: Record<string, CheckEntry>;
  dependencies: HealthResult[];
  relationship?: Relationship;
  critical?: boolean;
  latency_ms?: number | null;
  [evidence: string]: unknown;
}
export interface CheckReport {
  results: HealthResult[];
  exitCode: number;
  status: HealthStatus;
  spec_version: 'v1';
}
/** Schema validated; protocol-specific options use the canonical configuration schema. */
export interface TargetConfiguration {
  name: string;
  type: TargetType;
  endpoint?: string;
  checks?: string[];
  id?: string;
  [option: string]: unknown;
}
export interface ConfigurationDocument {
  version: 'v1';
  targets: TargetConfiguration[];
  [option: string]: unknown;
}
export interface ConfigurationOptions { baseDir?: string }
export class Configuration {
  constructor(data: ConfigurationDocument, options?: ConfigurationOptions);
  readonly baseDir: string;
  toJSON(): ConfigurationDocument;
}
export function loadConfig(path: string, options?: ConfigurationOptions): Promise<Configuration>;
export interface ClientOptions {
  binary?: string;
  /** Overall subprocess budget, in milliseconds; default 30000. */
  timeout?: number;
  environment?: Record<string, string | undefined>;
}
export interface CheckOptions { signal?: AbortSignal }
export type ConfigurationInput = Configuration | ConfigurationDocument | string;
export class Client {
  constructor(options?: ClientOptions);
  check(config: ConfigurationInput, options?: CheckOptions): Promise<CheckReport>;
}
export function check(config: ConfigurationInput, options?: ClientOptions & CheckOptions): Promise<CheckReport>;
export function parseReport(output: string, exitCode: number): CheckReport;
export class AgentHealthError extends Error {}
export class ConfigurationError extends AgentHealthError {}
export class BinaryNotFoundError extends AgentHealthError {}
export class CompatibilityError extends AgentHealthError {}
export class ExecutionError extends AgentHealthError {}
export class OutputError extends AgentHealthError {}
export class CheckTimeoutError extends AgentHealthError {}
export class CheckAbortedError extends AgentHealthError {}
