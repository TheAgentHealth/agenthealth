# Agent instructions for AgentHealth

These instructions apply throughout this repository. Read any more specific
`AGENTS.md` in a directory before changing its files. Follow the user's task
scope and preserve unrelated work in the working tree.

## Project and sources of truth

AgentHealth is a vendor-neutral health CLI and Go engine. Agent/multi-agent, HTTP/API, MCP, and A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility adapters are
implemented; other adapters follow the capability roadmap. The software is
pre-1.0. Do not describe planned adapters or distribution channels as available.

Read the relevant documents before making changes:

- [CONTRIBUTING.md](CONTRIBUTING.md): contribution and review policy.
- [ROADMAP.md](ROADMAP.md): capability phases and implementation status.
- [spec/README.md](spec/README.md): health, configuration, result, and adapter contracts.
- [core/README.md](core/README.md): engine behavior and safety guarantees.
- [RELEASING.md](RELEASING.md): versioning and release policy.
- [docs/cli.md](docs/cli.md), [docs/http-adapter.md](docs/http-adapter.md),
  [docs/mcp-adapter.md](docs/mcp-adapter.md), and [docs/a2a-adapter.md](docs/a2a-adapter.md): current CLI and adapter behavior.

The specifications define behavior; Go code, JSON schemas, fixtures, and
documentation must agree. Specification changes require the RFC-style
discussion described in CONTRIBUTING.md. Explicit user decisions can supply
task authorization; do not invent additional approval steps for routine work.

## Agent terminology and planned integrations

Use the agreed terminology consistently in code proposals, documentation,
examples, and reviews:

- **Direct agent:** the user's first, user-facing agent.
- **Composite agents:** downstream agents that the first agent communicates with, directly or through A2A.
- **Supporting dependencies:** models, tools, MCP servers, databases, vector stores, gateways, and routers. Their combined health is dependency health; do not call these dependencies composite agents.

These agent roles do not introduce target types or schema fields. `agent`
describes an individual agent, `a2a` a protocol endpoint, and `multi-agent`
the cooperating system as a whole. Keep evidence for each agent, the
communication path, and supporting dependencies distinguishable.

[Phase 7](ROADMAP.md#phase-7--agent-health) implements direct and composite agent
health; [Phase 8](ROADMAP.md#phase-8--agentgateway-integration) plans Agentgateway;
[Phase 9](ROADMAP.md#phase-9--agent-router-integration) plans Agent Router.
Dependency Graph and AHP follow in Phases 10 and 11 and remain planned; existing A2A endpoint checks do not establish that a first agent can
communicate with its downstream agents. See the
[target model](spec/target-model.md#direct-and-composite-agent-health).

## Contribute changes

For every new phase or substantive phase revision, follow the
[Phase Impact Review](ROADMAP.md#phase-impact-review). Assess earlier and later
phases, record affected layers or an explicit no-impact rationale, add required
extension tasks to the current phase, and update affected future deliverables.
Revisit the assessment during implementation; complete the required updates
and validation before marking the phase complete.

1. Inspect `git status`, the current branch, and relevant code before editing.
   Do not discard existing changes, reset branches, or commit unrelated files.
2. For substantial proposals, use an issue or discussion as described in
   CONTRIBUTING.md. Implement the work already authorized by the user.
3. Create a focused branch from current `main`. External contributors use a
   fork; maintainers may use a branch in the repository.
4. Keep changes within the requested scope. Put CLI behavior in
   `cmd/agenthealth`, shared execution/model/output in `core`, and protocol
   behavior in `adapters/<protocol>`.
5. Add meaningful tests for new behavior and regressions. Update documentation
   and examples when behavior changes. Schema changes need valid/invalid
   fixtures in `tests/spec/fixtures` and matching Go configuration validation.
6. Run the applicable checks below. Use descriptive commits, preferably with
   `feat:`, `fix:`, `docs:`, `spec:`, or `chore:` prefixes.
7. Open a PR against `main` with the problem, resulting behavior, validation,
   and material limitations. Address feedback; merge under the contribution
   policy after maintainer approval and green CI. Do not bypass protections.

Never commit compiled binaries, release archives, credentials, or environment
files. The local `/agenthealth` binary and `/dist/` outputs are ignored.

## Preserve behavior and safety

- Keep the shared health dimensions and exit-code contract consistent with
  `spec/`. Exit codes 0–5 represent health results; 6 represents tool failures.
- Preserve dependency aggregation and independent dependency policies.
- Keep active checks explicit, non-destructive, and non-retrying. Passive
  retries must respect the engine's response and cancellation rules.
- Verify TLS; do not add insecure verification or automatic redirect following.
- Resolve credentials through configured references. Do not print secrets,
  raw adapter errors, response bodies, or sensitive expected values.
- Use engine-owned diagnostics and allowlisted transport stages. Honor check
  deadlines, context cancellation, concurrency bounds, and body size limits.
- Keep machine output compatible across JSON/YAML, and terminal output free
  of control characters. Update result validation when extending the model.

## Local validation

Use Go 1.23 or newer and Python with `tests/requirements.txt` installed. Check
`go version` first; a toolchain under `/tmp` is temporary and must not be assumed
to exist or documented as the normal installation path.

Run from the repository root for code, schema, or release changes:

```bash
gofmt -w core adapters examples cmd
go vet ./...
go test -race ./...
python3 -m pip install -r tests/requirements.txt
python3 -m pytest tests -q
python3 scripts/check_markdown_links.py
python3 scripts/validate_doc_examples.py
git diff --check
```

Format only files affected by the task when unrelated edits are present.
For documentation-only changes, run the link checker, documentation example
validator, and `git diff --check`; do not add unrelated tests. Mark canonical
JSON/YAML examples with `<!-- spec-example: result -->` or
`<!-- spec-example: configuration -->`, and include newly covered documentation
in `scripts/validate_doc_examples.py` when needed.

## Prepare and publish a release

Publish only when the user requests a release. Preparing release tooling or
documentation does not by itself authorize pushing a tag. A tag push triggers
publication through [.github/workflows/release.yml](.github/workflows/release.yml).

1. Inspect existing remote tags/releases and choose the next version under
   RELEASING.md. Never reuse, move, or overwrite a published tag. Software
   versions such as `v0.1.1` are independent of the `spec_version: v1` contract.
2. Add `docs/releases/<tag>.md` before tagging. Include changes, installation,
   breaking changes/migrations, and actual limitations. Update installation
   examples that reference an older version where appropriate.
3. Validate locally and prepare archives with
   [scripts/build_release.py](scripts/build_release.py). The output directory
   must be empty. Example for a new patch release:

   ```bash
   python3 scripts/build_release.py v0.1.1 --output /tmp/agenthealth-v0.1.1-assets
   ```

   Check archive contents, hashes, and the native binary's `version` and CLI
   behavior. Do not claim runtime testing on platforms only cross-compiled.
4. Commit the release changes in a PR, obtain the required review, and wait for
   all CI checks, including Go tests. Merge to `main` under repository policy.
5. Check that the working tree is clean and local `main` matches the intended
   remote merge commit. Tag that exact commit, using the selected version:

   ```bash
   git switch main
   git pull --ff-only
   git status --short
   git log -1 --oneline
   git tag -a v0.1.1 -m "AgentHealth v0.1.1"
   git push origin v0.1.1
   ```

   Run these steps sequentially and stop on any failure. Push only the selected
   tag, not every local tag. Do not tag an unmerged feature branch.
6. Watch the Release workflow using GitHub CLI or GitHub Actions. It validates
   the tagged source, builds Linux AMD64/ARM64, macOS AMD64/ARM64, and Windows
   AMD64 archives, adds `checksums.txt`, smoke-tests Linux AMD64, uploads assets
   to a draft, and publishes a regular **Latest** release for plain version tags, or a
   **prerelease** for suffixed preview tags, after upload succeeds.
7. Verify that the release is public, targets the intended tag, and contains all
   five archives, five matching CycloneDX SBOMs, and checksums. Verify signed
   provenance and SBOM attestations before announcing the release. Download published assets, check hashes, and
   smoke-test the native binary. Report the release URL and validation results.

Pre-1.0 versions can be regular releases; use suffixed tags for intentional
previews as described in RELEASING.md. v0.4.0 publishes CycloneDX SBOMs and
keyless signed provenance/SBOM attestations. OS-native executable signing and
macOS notarization remain planned.

If a workflow fails, inspect its logs and any draft release before retrying.
The create-release step is not idempotent: an existing draft can cause a rerun
to fail. Preserve tag integrity, avoid publishing incomplete assets, and do not
delete or replace public releases without explicit authorization. A source fix
requires a new version; infrastructure-only recovery must use the same tagged
source and verified complete assets.
