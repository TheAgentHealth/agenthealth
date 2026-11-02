# Phase 23 executable adapters

Proposal: [RFC #68](https://github.com/TheAgentHealth/agenthealth/issues/68).
The maintainer authorized implementation of Phase 23. This records the additive
adapter transport contract; AHS/AHP v1 and target configuration remain unchanged.

An operator supplies a bounded JSON manifest, protocol `v1`, absolute executable
path, lowercase SHA-256 and existing adapter metadata. Registration validates
compatibility `v1`, known target types and dimensions, mandatory configuration /
reachability gates, active functional checks, and duplicate target ownership.
The CLI registers built-ins first; executable adapters can claim `custom`,
`tool`, or `runtime`. Embedded Go callers can construct a registry with other
explicit adapter choices. There is no automatic replacement or discovery.

Each check starts the pinned executable with no arguments, empty environment,
a single JSON request on stdin and a 64 KiB maximum stdout stream. Request fields
are `protocol`, `dimension`, `name`, `type`, `endpoint` and optional `credential`.
Only the target's configured bearer credential is forwarded; database, OAuth,
model and other rich options are unsupported by this transport. No dependencies,
arbitrary configuration, environment snapshot or transport client is serialized.

The executable writes optional `{"kind":"response"}` immediately upon any target
response, followed by exactly one `{"kind":"result","status":"HEALTHY"}` (or any
other AHS health status), then exits successfully. Unknown fields, extra values,
invalid statuses, excessive output and nonzero exit fail closed to engine-owned
UNKNOWN diagnostics. Response markers preserve partial-response timeout semantics.
No response marker means no received response, including on cancellation.
The engine retains configuration/reachability prerequisites, active opt-in,
retry suppression after responses, timing, aggregation, redaction and outputs.

The executable must verify TLS, reject redirects, respect cancellation, avoid
side effects and avoid logging secrets. Stderr is discarded. On cancellation the
host terminates the child and closes its output pipe; Unix also kills its process
group. On other platforms child descendants are not guaranteed to be terminated.
No long-lived session or cleanup hook is exposed. Every dimension is independent.

This is lifecycle separation, not an OS sandbox. A program has the user's OS
permissions and network access. Manifests, executable directories, shared libraries
and interpreter dependencies must be operator trusted. The hash is checked before
each spawn; concurrent replacement of an executable in a writable directory is
outside this trust boundary. No executable download, registry certification,
filesystem isolation, egress policy or security audit is supplied. Phase 26 owns
stronger OS containment. Phase 24 will test this contract independently; Phase 25
will measure real community adapters separately from controlled protocol fixtures.

On Windows, Go may supply the OS-required `SystemRoot` variable; application
environment variables remain excluded. Unix group termination is used before
reaping starts. After stdout closes and reaping starts, cancellation uses the
lifetime-aware direct-child handle; descendants that close inherited pipes are
not guaranteed to be terminated. OS containment remains outside this protocol.
