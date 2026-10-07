# Specification changelog

## Foundation stabilization — software v0.4.0

- Clarify that server protocol incompatibility and missing explicitly required
  functionality are UNHEALTHY; invalid local version pins stay MISCONFIGURED.
- Add optional stable check diagnostic codes and their identifier pattern.
- Extend A2A configuration with protocol 1.0 and extendedAgentCard expectations;
  retain explicit protocol 0.3.0. The reference adapter’s unpinned default changes
  to 1.0; migrate legacy configurations by explicitly pinning 0.3.0.
- Extend the Go adapter request with per-target resource context and optional
  bounded TargetCloser cleanup. The two required adapter methods remain intact.

The machine-readable envelope remains spec_version v1: required fields, health
states and exit-code mapping are unchanged, and schema additions are optional.
The pre-1.0 software minor version and migration notes identify the changed
reference protocol default and corrected classifications.
