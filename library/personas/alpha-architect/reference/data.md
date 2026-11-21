# Data Durability & Mobility

Read for any design that owns or moves data. Durability and Mobility are Core
Principles #3 and #4 — apply these gates to every design that persists state.

## Data Durability gate

- [ ] What is the write-ahead / journaling strategy?
- [ ] What happens if the primary datastore goes down mid-transaction?
- [ ] Are backups automated, tested, and versioned?
- [ ] Is there a point-in-time recovery capability?
- [ ] Are idempotency keys used for critical writes?
- [ ] Is there an event log / audit trail for state changes?

## Data Mobility gate

- [ ] Can data be exported in a standard format (JSON, Parquet, CSV)?
- [ ] Are schemas versioned with backward-compatible evolution (e.g., Avro, Protobuf)?
- [ ] Is there a CDC (Change Data Capture) or event stream for downstream consumers?
- [ ] Can the system support multi-region replication if needed later?
- [ ] Are PII fields tagged for compliance-aware data movement (GDPR, HIPAA)?

> For schema modelling, normalization, index design, and datastore selection,
> hand off to the `dbarch` persona — Alpha sets data strategy, Dbarch designs
> the data layer.
