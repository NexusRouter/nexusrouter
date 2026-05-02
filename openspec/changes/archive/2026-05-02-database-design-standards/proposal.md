## Why

NexusRouter and related services will eventually need durable relational storage; without written standards, schema work drifts toward ad-hoc normalization, unsafe redundancy, and inconsistent review expectations. Codifying industry-proven rules (3NF-first, selective denormalization, sharding and indexing alignment) now reduces rework and review friction before persistence layers grow.

## What Changes

- Introduce a formal OpenSpec capability that defines **normative** relational database design rules: normalization trade-offs, production practices (core vs query layer, hot-field redundancy, read/write layering), sharding with normalization, indexing, and special-case handling.
- Add a **design** document that records architectural decisions (where the spec lives, how reviews apply it, what is out of scope for v1).
- No runtime gateway API or code behavior changes in this change; deliverable is specification and design artifacts only.

## Capabilities

### New Capabilities

- `database-design-standards`: Relational schema design rules covering normalization vs denormalization, redundancy and synchronization, partitioning, indexing, and exceptional workload types (logs, analytics, archives).

### Modified Capabilities

- (none)

## Impact

- **Specs**: After archive, `openspec/specs/database-design-standards/spec.md` becomes the long-lived requirement set for database-related work.
- **Process**: Pull requests that introduce or alter relational schemas SHOULD be checked against this spec during review (manual until tooling exists).
- **Code**: No mandatory code or migration changes in this proposal phase.
