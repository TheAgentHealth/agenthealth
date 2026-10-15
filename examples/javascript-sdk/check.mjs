import { check } from '@theagenthealth/agenthealth';

const report = await check({
  version: 'v1',
  targets: [{ name: 'api', type: 'http',
    endpoint: process.argv[2] ?? 'http://localhost:8080/health',
    checks: ['reachability'] }]
}, { binary: process.env.AGENTHEALTH_BINARY ?? 'agenthealth', timeout: 30000 });
console.log(JSON.stringify(report, null, 2));
process.exitCode = report.exitCode;
