# SOLID Principles & Design Patterns

Read when evaluating code-level architecture — service internals, module
boundaries, and dependency graphs.

- **Single Responsibility** — Every class, module, and service should have one reason to change. If a service owns both business logic and infrastructure concerns, it's a refactoring target.
- **Open/Closed** — Design for extension without modification. Plugin architectures, strategy patterns, and event hooks are preferred over switch statements and conditional chains.
- **Liskov Substitution** — Subtypes must be substitutable for their base types without breaking callers. Violations here signal leaky abstractions — fix the contract, not the caller.
- **Interface Segregation** — Clients should not depend on methods they don't use. Fat interfaces create coupling. Prefer role-based interfaces (e.g., `Readable`, `Writable`) over god interfaces.
- **Dependency Inversion** — High-level modules depend on abstractions, not concrete implementations. Use DI containers or manual constructor injection. No service should instantiate its own dependencies.
- **Key patterns to evaluate for** — Repository (data access abstraction), Unit of Work (transactional consistency), Strategy (swappable algorithms), Factory (encapsulated construction), Command (deferred execution), Decorator (layered behaviour), Adapter (interface translation), Anti-corruption Layer (isolating legacy or third-party systems).
- **When to flag** — Over-abstraction is as harmful as under-abstraction. If a pattern adds indirection without enabling a concrete extension or testability need, it's premature. Three concrete usages before extracting an abstraction.
