#!/usr/bin/env python3
"""Run the reference behavioral conformance suite with mandatory case evidence."""
import argparse
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
PREFIX = 'github.com/TheAgentHealth/agenthealth/'
# Exact case names make renamed, deleted or skipped coverage fail closed.
CASES = {
    'core': ['TestConformanceDocuments', 'TestConformanceDependencies',
             'TestSeverityAndDependencies', 'TestPrerequisiteGates',
             'TestPassiveDefaultsAndEmptyChecks', 'TestDependenciesRunAfterParentFailure',
             'TestRetriesAndActiveNeverRetried', 'TestTimeoutPartialAndIgnoringAdapter',
             'TestAdapterPanicInvalidResultAndSecrets', 'TestRegistryAndBoundedCalls',
             'TestGraphValidation', 'TestGraphSharedEvidenceAndPolicies',
             'TestGraphParallelBoundAndBudget', 'TestGraphSharedActiveCheckOnceAndFreshRuns',
             'TestGraphParentBudgetDoesNotCancelSharedDependency',
             'TestGraphRedactsSecretsRegisteredByLaterNodes',
             'TestAHPHTTPContract', 'TestAHPStatusMappingAndGraphEvidence',
             'TestObservabilityGraphAndRedaction', 'TestObservabilityValidationAndFailures',
             'TestCIReports'],
    'cmd/agenthealth': ['TestConformanceCLI'],
    'adapters/http': ['TestPassiveReadinessStatus', 'TestHeadersAndBody',
                      'TestInvalidCredentialControlsFailBeforeNetworking'],
    'adapters/mcp': ['TestModernHTTP', 'TestModernErrorsDoNotFallBack',
                     'TestOAuthAcquisitionAndRedaction'],
    'adapters/a2a': ['TestV1DefaultAndExplicitLegacy', 'TestPassiveAndFunctional',
                     'TestRedirectDoesNotForwardCredential'],
    'adapters/agent': ['TestAgentHealth', 'TestFunctionalAndPath',
                       'TestFirstRuntimeContactsSelectedPeer'],
    'adapters/gateway': ['TestGatewayAndBackendEvidence'],
    'adapters/router': ['TestRouterAndBackendEvidence', 'TestIndependentRouteEvidence'],
    'adapters/model': ['TestTransportSafety', 'TestInferenceDeadlineNeverRetries'],
    'adapters/database': ['TestEndpointSafety', 'TestReachabilityAndCancellation'],
    'adapters/vectorstore': ['TestProviderMetadataAndBoundedQuery',
                             'TestFunctionalFailureNeverRetries'],
    'adapters/process': ['TestProtocol', 'TestManifestAndPin'],
}


def evaluate(output, returncode):
    outcomes = {}
    skipped = set()
    malformed = False
    for line in output.splitlines():
        try:
            event = json.loads(line)
            if not isinstance(event, dict):
                raise ValueError('event')
        except (ValueError, TypeError):
            malformed = True
            continue
        key = (event.get('Package'), event.get('Test'))
        if event.get('Action') in ('pass', 'fail', 'skip'):
            outcomes[key] = event['Action']
            if event['Action'] == 'skip':
                skipped.add(key)
    cases = []
    for package, names in CASES.items():
        for name in names:
            key = (PREFIX + package, name)
            # A skipped subcase must not be hidden by its passing parent.
            child_skip = any(p == key[0] and (n or '').startswith(name + '/')
                             for p, n in skipped)
            cases.append({'package': package, 'test': name,
                          'passed': outcomes.get(key) == 'pass' and not child_skip})
    passed = returncode == 0 and not malformed and all(case['passed'] for case in cases)
    passed = passed and not any(value == 'fail' for value in outcomes.values())
    return {'conformance_version': 'v1', 'profile': 'reference-behavior',
            'passed': passed, 'cases': cases,
            'skipped_tests': len(skipped)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--report', type=Path, required=True)
    args = parser.parse_args()
    try:
        run = subprocess.run(['go', 'test', '-race', '-count=1', '-json',
                              './core', './cmd/agenthealth', './adapters/...'],
                             cwd=ROOT, capture_output=True, text=True,
                             encoding='utf-8', errors='replace', timeout=900)
        report = evaluate(run.stdout, run.returncode)
    except (OSError, subprocess.TimeoutExpired):
        report = evaluate('', 6)
    args.report.write_text(json.dumps(report, indent=2) + '\n', encoding='utf-8')
    print('Reference conformance passed' if report['passed'] else 'Reference conformance failed; inspect report and run Go tests for diagnostics')
    return 0 if report['passed'] else 6


if __name__ == '__main__':
    raise SystemExit(main())
