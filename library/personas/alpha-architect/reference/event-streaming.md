# Event-Driven & Streaming Architecture

Read when the system involves asynchronous workflows, real-time data, or
cross-service coordination.

- **Event Sourcing** — Store state as an append-only sequence of events. Evaluate when auditability, temporal queries, or complex domain logic warrant it. Flag the operational cost (projection rebuilds, schema evolution of events).
- **CQRS** — Separate read and write models when read patterns diverge significantly from write patterns. Not every service needs CQRS — apply where it reduces complexity, not adds it.
- **Streaming platforms** — Kafka, Redpanda, or managed equivalents (AWS MSK, Confluent Cloud). Design topics with clear ownership, partitioning strategy, and retention policies.
- **Event schemas** — Use a schema registry with Avro or Protobuf. Enforce backward-compatible evolution. JSON Schema as a fallback for lighter workloads.
- **Saga / Choreography patterns** — For distributed transactions, prefer choreography (events) for loosely coupled flows; orchestration (saga coordinator) when ordering and compensation logic is complex. Always define compensating actions for each step.
- **Dead letter queues** — Every consumer must have a DLQ strategy. Unprocessable messages are not silently dropped.
- **Idempotency** — Every event consumer must be idempotent. Design idempotency keys into event schemas from day one.

## Review gate

- [ ] Does every consumer have a DLQ and an idempotency key?
- [ ] Are event schemas registered and evolved backward-compatibly?
- [ ] Are compensating actions defined for each step of a distributed transaction?
- [ ] Is the operational cost of event sourcing / CQRS justified by a concrete need?
