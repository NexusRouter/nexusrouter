# database-design-standards Specification

## Purpose

Normative rules for relational schema design in NexusRouter and related services: default normalization, controlled denormalization and redundancy, partitioning and indexing alignment, and exceptions for logs, analytics, and archives.

## Requirements
### Requirement: Normalization baseline and explicit trade-off awareness

All new relational schemas for production workloads SHALL assume third normal form (3NF) as the default logical model unless a documented exception applies per other requirements in this spec. Schema authors MUST record the main normalization trade-offs they accept (fewer anomalies vs more JOINs and tuning surface) when deviating from strict normalization for a bounded surface.

#### Scenario: Core transactional domain

- **WHEN** designing tables for core transactional flows (payments, transfers, order creation, ledger-like writes)
- **THEN** the logical model SHALL conform to 3NF (or stronger where appropriate) and SHALL NOT rely on unstored redundant copies of mutable foreign attributes for correctness

#### Scenario: Team documents a denormalization exception

- **WHEN** a table intentionally stores redundant data for performance
- **THEN** the change MUST cite the applicable denormalization requirement in this spec and MUST document the redundant columns, owning source of truth, and sync or derivation mechanism in the schema change notes

### Requirement: Business-first priority over dogmatic purity

When normalization and latency or operational cost conflict, implementers SHALL prefer **data correctness and clear ownership of truth** on write paths, and MAY apply controlled denormalization on read paths or derived stores as specified in this spec. “Business-first” MUST NOT justify silent duplication of frequently updated attributes without a sync strategy.

#### Scenario: Read-mostly product surface

- **WHEN** a feature is predominantly read-heavy with rare writes (catalog listing, public profile snapshot, reporting slice)
- **THEN** denormalized projections or redundant columns MAY be used if they satisfy the redundancy documentation and synchronization requirements in this spec

### Requirement: Core normalized storage with query-layer denormalization

Persistent core entities (e.g., user, product, order heads and lines) SHALL remain normalized in the primary write store. Read-optimized shapes (replica tables, materialized views, or cache payloads) MAY denormalize for hot queries if they are clearly secondary to the normalized source and kept consistent.

#### Scenario: Single-table read on replica

- **WHEN** serving a high-traffic list that would otherwise require many JOINs from the write model
- **THEN** a read model MAY duplicate stable or slowly changing display fields from referenced entities provided updates propagate from the authoritative tables

#### Scenario: Application data access

- **WHEN** application code reads denormalized projections
- **THEN** data access boundaries SHOULD hide whether data came from a join, a projection table, or a cache so callers do not depend on accidental physical layout

### Requirement: Hot-field redundancy policy

Redundant columns SHALL be limited to **low update frequency** attributes relative to read frequency (display names, category labels) or **aggregates** maintained explicitly (follower counts, sales counters). Redundancy MUST be forbidden for attributes that change very frequently relative to business tolerance unless a proven low-latency sync path is specified and monitored.

#### Scenario: Allowed redundancy

- **WHEN** a column duplicates a rarely changing label from another entity to avoid a JOIN on every list row
- **THEN** the schema documentation MUST name the authoritative column and the replication path (e.g., transactional outbox, change data capture, batch job) and reviewers MUST verify the path exists or is planned before merge

#### Scenario: Forbidden redundancy

- **WHEN** a candidate redundant field updates often and inconsistency would confuse users or violate money-like invariants
- **THEN** the design MUST NOT duplicate that field in a secondary table for convenience; it SHALL be read from the authoritative row or joined

### Requirement: Consistency assurance for redundant data

Every redundant field or read model derived from normalized tables SHALL have a defined **update propagation** mechanism (triggers, message-driven consumers, stream processors, scheduled reconciliation, or database replication to a transformed replica) and a **periodic or continuous consistency verification** suitable to the business risk (e.g., nightly diff job for soft-reporting aggregates).

#### Scenario: Propagation on source update

- **WHEN** the source row for a redundant display field changes
- **THEN** dependent redundant copies SHALL be updated within the documented SLA or marked stale per an explicit policy

#### Scenario: Operational drift detection

- **WHEN** the system runs periodic consistency checks
- **THEN** discrepancies SHALL be logged and triaged and remediation playbooks SHALL exist before treating the redundant copy as authoritative for money-critical reads

### Requirement: Read/write topology alignment

Primary write databases SHALL hold the normalized canonical schema. Read replicas or analytical stores MAY hold additional denormalized columns or projections. Teams MUST understand that replica lag means denormalized read surfaces MAY be briefly stale unless queries are routed or engineered for strong consistency.

#### Scenario: Default read after write

- **WHEN** a user expects to immediately see their own write in the same session
- **THEN** reads that require freshness SHALL hit the primary or use a pattern that guarantees read-your-writes; denormalized replicas alone are insufficient without that guarantee

### Requirement: Partitioning combined with in-shard normalization

Vertical splits (separate databases or tables by domain or cold/hot columns) and horizontal sharding SHALL preserve **3NF within each shard/table** unless a redundancy rule in this spec is explicitly invoked per shard. Shard keys SHALL favor stable, high-cardinality dimensions common in queries (e.g., user id, order id) with even spread; range vs hash strategy MUST be justified for the dominant access pattern.

#### Scenario: Hot/cold vertical split

- **WHEN** splitting wide rows into a narrow hot table and a wide cold table
- **THEN** each resulting table SHALL remain normalized with clear foreign keys or shared keys and no hidden duplication across the split without documentation

#### Scenario: Hash sharded user data

- **WHEN** sharding by user identifier hash
- **THEN** co-location rules MUST avoid cross-shard joins for the hottest transactional paths or those paths MUST be redesigned

### Requirement: Index strategy coordinated with normalization model

Primary keys SHOULD use monotonic integer surrogates where feasible to limit index fragmentation; random UUID primary keys SHOULD be avoided unless paired with mitigations (e.g., ordered UUID variants) and documented rationale. Foreign key columns participating in JOINs SHALL be indexed. Index sets SHALL balance read acceleration against write amplification; over-indexing without query evidence is disallowed for new tables.

#### Scenario: Join-heavy normalized model

- **WHEN** two 3NF tables are joined on a foreign key for frequent queries
- **THEN** the foreign key column (or left side of the join predicate) SHALL be indexed unless a quantitative review shows full scans are acceptable

#### Scenario: Covering index on denormalized read model

- **WHEN** a denormalized list table serves a narrow query pattern
- **THEN** composite indexes MAY target that pattern to avoid repeated random lookups if write rate remains acceptable

### Requirement: Exceptional workload classes

Append-only or append-mostly logs and high-volume event streams MAY relax strict normalization when access patterns are append-only and immutability is guaranteed. Data warehouse and large-scale reporting models MAY be predominantly denormalized for batch efficiency. Historical archives MAY store denormalized snapshots when data is immutable. These exceptions MUST be labeled as such in schema and operational docs.

#### Scenario: Immutable audit log

- **WHEN** storing append-only audit events with some duplicated context fields per row
- **THEN** the table MUST be documented as a log store and MUST NOT be used as the mutable source of truth for those duplicated attributes

#### Scenario: Warehouse star schema

- **WHEN** building analytical tables for BI
- **THEN** denormalized star or snowflake designs are permitted with warehouse-specific freshness SLAs, separate from OLTP invariants

### Requirement: Anti-pattern guard for volatile redundancy

Teams MUST NOT introduce redundant copies of attributes that change frequently relative to acceptable inconsistency windows, except for immutable snapshots or explicitly engineered stream-derived counters with monitoring.

#### Scenario: Reject volatile copy

- **WHEN** a proposal duplicates a user's current account balance or inventory quantity in a second mutable table for convenience
- **THEN** the design SHALL be rejected or redesigned to a single authoritative balance/inventory row or a rigorously tested event-sourced projection with clear consistency semantics

