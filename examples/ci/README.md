# CI deployment gates

Use a trusted deployment configuration with an exact-version, verified CLI.
Replace the placeholder HTTPS readiness endpoint in
[agenthealth.yaml](agenthealth.yaml) with your deployed service. It uses passive
checks; explicitly select safe functional checks only when required.

- [GitHub Actions](github.yml): verified install, reusable local Action and report artifact.
- [GitLab](gitlab.yml): JUnit report even on failure.
- [Jenkins](Jenkinsfile): JUnit plugin, timeout and failure retention.
- [Azure DevOps](azure-pipelines.yml): publish reports with `always()`.
- [Shell](gate.sh): `sh examples/ci/gate.sh configuration.yaml report.xml`.

The non-GitHub examples assume the CLI was installed and verified before the
job, using the [installer](../../scripts/install_ci_binary.py) or the
[installation guide](../../docs/installation.md). They are templates, not
claims of testing against hosted GitLab/Jenkins/Azure services.

Exit codes 1–5 stop promotion; 6 means the tool failed and a complete report
may not exist. Keep reports on failure, restrict artifact access and expiry,
and never use `|| true` on the gate. Configure runner/job deadlines above the
configuration budget. See the [guide](../../docs/ci-cd.md) for evidence mapping,
credential boundaries, Kubernetes validation and GitHub Action inputs.
