import { check, Client, Configuration, type Relationship, type HealthStatus } from '../index.js';
const config = new Configuration({ version: 'v1', targets: [{ name: 'api', type: 'http', endpoint: 'https://example.com' }] });
const report = await check(config, { timeout: 5000 });
const status: HealthStatus = report.status;
const relationship: Relationship | undefined = report.results[0].dependencies[0]?.relationship;
await new Client().check(config, { signal: new AbortController().signal });
void status; void relationship;
// @ts-expect-error unknown target type
new Configuration({ version: 'v1', targets: [{ name: 'bad', type: 'typo' }] });
