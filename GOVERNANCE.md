# Governance

AgentHealth is intended to evolve as a community-driven, vendor-neutral, open-source project under the TheAgentHealth organization. This document describes how decisions are made today, and how that is expected to evolve.

## Current stage

The project is pre-1.0 (see [Project Status](README.md#project-status)). Governance is currently lightweight and maintainer-driven; it is expected to formalize as the contributor base and adoption grow.

## Roles

- **Maintainers** — review and merge pull requests, triage issues, and make final calls on implementation changes.
- **Contributors** — anyone who opens issues, pull requests, or participates in specification discussions.

There is currently no formal application process for becoming a maintainer; this will be defined as the project grows past its initial phases.

## Decision-making

- **Implementation changes** (core engine, CLI, adapters, SDKs, integrations) use lazy consensus: a pull request can be merged by a maintainer once reviewed, absent sustained objection.
- **Specification changes** (anything under [spec/](spec/README.md)) require an RFC-style process, because they affect every downstream adapter, SDK, and (eventually) AHP conformance test:
  1. Open an issue or discussion proposing the change and its rationale.
  2. Allow time for community feedback, especially from adapter/SDK maintainers.
  3. A maintainer merges the spec change once objections are resolved or a clear majority supports it.
- **Agent Health Protocol (AHP)** changes are held to an even higher bar once AHP moves out of "proposed/experimental" status (see [spec/protocol.md](spec/protocol.md)), since AHP is meant to be implementable by parties outside this repository.

## Who can participate

The project welcomes participation from:

- AI agent developers
- MCP developers
- A2A developers
- platform engineers
- SREs and DevOps engineers
- AI infrastructure engineers
- cloud-native developers
- framework maintainers
- model providers
- enterprises
- researchers
- standards communities (e.g. groups such as AAIF, should the project pursue broader interoperability contribution in the future)

## Long-term direction

As adoption grows, the project intends to move toward:

- a documented maintainer nomination/election process,
- a technical steering body for AHS/AHP specification changes,
- transparent public meeting notes or written RFC archives.

Any certification or trademark program (see [Phase 24 — AgentHealth Conformance](ROADMAP.md#phase-24--agenthealth-conformance) in the roadmap) would require separate governance and community approval before being introduced.

## Changes to this document

Changes to GOVERNANCE.md itself follow the same RFC-style process as specification changes.
