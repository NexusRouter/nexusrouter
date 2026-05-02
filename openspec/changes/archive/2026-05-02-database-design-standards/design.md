## Context

NexusRouter is an OpenAI-compatible API gateway (Go, AGPL-3.0). Persistence requirements are evolving; the repository previously lacked a single, reviewable source for relational database design. Stakeholders are backend engineers and reviewers approving schema changes, migrations, and operational data layouts.

## Goals / Non-Goals

**Goals:**

- Capture **normative** design rules in OpenSpec (`database-design-standards`) so “what good looks like” is explicit and testable at review time.
- Encode the **paradigm** from the proposal: normalization as the default for write-critical paths, pragmatic denormalization for read-heavy surfaces, documented redundancy with sync and verification, alignment of indexes and sharding with that model.
- Provide **reference patterns** (e-commerce order, social user, content feed) as illustrative comparisons in this design doc—not as binding requirements—to ground abstract rules in familiar systems.

**Non-Goals:**

- Choosing a specific RDBMS vendor, ORM, or migration toolchain.
- Defining concrete table DDL for NexusRouter products (deferred until a feature needs persistence).
- Automating compliance (linters, CI schema gates) in this change; optional follow-up.

## Decisions

1. **Spec as single source of truth** — Requirements live in `specs/database-design-standards/spec.md` under this change, then under `openspec/specs/` after archive. Design.md explains intent and process; it does not duplicate every SHALL.

2. **“范式为底，业务为先” encoded as hierarchy** — The spec SHALL order priorities: transactional correctness and integrity for core domains first; performance-oriented denormalization only where justified and documented. This resolves tension between purity and pragmatism without leaving it implicit.

3. **Redundancy is permitted only with an explicit contract** — Any denormalized or replica-only column MUST be listed with sync mechanism and consistency check expectations in schema documentation or migration notes (the spec mandates documentation and operational discipline, not a specific tool like Canal or Flink).

4. **Sharding guidance stays logical** — The spec describes shard key selection and keeping shards internally normalized; physical topology (number of shards, middleware) remains a deployment decision.

5. **Reference implementations named without mandating stack** — Alibaba (Canal), Tencent (TencentDB triggers), ByteDance (Flink) appear here as **examples** of sync/analytics patterns; the spec refers generically to binlog, messaging, batch reconciliation, and stream processing.

**Alternatives considered:**

- **Wiki-only standards** — Rejected: harder to version with code and to gate in review; OpenSpec aligns with repo workflow.
- **One giant markdown in `docs/`** — Rejected for this workflow: OpenSpec gives structured ADDED requirements and scenarios for future archive.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Spec feels too “enterprise” for a small gateway codebase | Non-goals and “startup phase” guidance in requirements allow 3NF-first without premature denorm. |
| Reviewers treat examples (Canal/Flink) as mandatory | Design explicitly marks vendor patterns as illustrative; spec uses generic SHALL for sync and verification. |
| Requirements become untestable prose | Each requirement keeps at least one `#### Scenario` with WHEN/THEN for review checklists. |

## Migration Plan

- **Deploy**: Merge change; run `openspec archive` when implementation tasks are done (here: artifacts complete; “implementation” is adoption in review, not code migration).
- **Rollback**: Revert the change directory or archived spec; no data migration involved.

## Open Questions

- Whether to add a short `CONTRIBUTING` pointer to this spec after archive (optional follow-up; not part of this change’s tasks unless requested).
- When first real tables land: whether to add an ADR per major schema with explicit redundancy map.

## Appendix: Normalization trade-off summary

| Aspect | Advantage | Disadvantage |
|--------|-----------|----------------|
| Normalization | Minimal redundancy; easier consistency; fewer insert/update/delete anomalies; cheaper logical change (single place) | More JOINs; more design/maintenance complexity; read-heavy workloads can hit bottlenecks; more nuanced index tuning |
| (Implied denorm trade-off) | Fewer JOINs; simpler hot-path reads | Higher redundancy; sync and consistency cost; risky for frequently updated fields |

## Appendix: Reference patterns (non-normative)

- **Order systems (e-commerce)** — Core order header/line/payment/logistics in normalized form; list views on replicas with redundant product/shop labels; cache for hot reads; binlog-driven sync to read models.
- **Social user systems** — Core profile normalized; home feeds and notifications denormalize counters and display names with triggers and/or batch reconciliation.
- **High-QPS content feeds** — Core content normalized; feed/list replicas denormalize author display fields; tiered cache plus stream or batch updates for counters.
