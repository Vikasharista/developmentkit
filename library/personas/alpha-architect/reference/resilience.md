# Resilience & Distributed Systems Patterns

Read when the system is distributed across multiple services, regions, or
failure domains.

- **Circuit breakers** — Wrap calls to downstream services with circuit breakers (Resilience4j, Polly, or your platform/mesh's equivalent). Define open/half-open/closed thresholds based on error rate and latency percentiles, not just timeouts. A slow dependency should trip the breaker before it cascades.
- **Bulkheads** — Isolate failure domains. Separate thread pools, connection pools, or process boundaries for independent downstream dependencies so that one degraded service doesn't starve all others.
- **Timeouts & retries** — Every network call has an explicit timeout. Retries use exponential backoff with jitter. Set a retry budget (e.g., max 3 retries, max 20% of requests retried) to prevent retry storms.
- **Service discovery** — Services locate each other through platform-native discovery (Kubernetes Services, AWS Cloud Map) or a registry (Consul). Hardcoded URLs are a deployment-time landmine.
- **Service mesh** — For complex microservice topologies, evaluate a service mesh (Istio, Linkerd, Cilium) for consistent mTLS, traffic shaping, observability, and retry policies. Not every system needs one — the operational overhead must justify the benefits.
- **Load balancing** — Client-side (gRPC built-in, mesh sidecar) vs. server-side (ALB, Envoy). Evaluate health-check–based routing, weighted routing for canary, and locality-aware balancing for multi-region.
- **Distributed caching** — Redis Cluster or Memcached for shared state. Design for cache invalidation (TTL, event-driven invalidation, write-through). A stale cache is worse than no cache — define consistency requirements per cache.
- **Graceful degradation** — Every service must define its degraded mode. When a dependency is down, what does the user see? A fallback response, a cached result, or a clear error? Never an infinite spinner.

## Review gate

- [ ] Do all downstream calls have explicit timeouts and circuit breakers?
- [ ] Are retries bounded with exponential backoff, jitter, and a retry budget?
- [ ] Are failure domains isolated (bulkheads for thread pools, connection pools)?
- [ ] Does every service define a degraded mode (fallback, cached response, clear error)?
- [ ] Is service discovery dynamic (not hardcoded URLs)?
- [ ] Are cache invalidation strategies defined per cache layer?
- [ ] Is there a chaos engineering or fault injection practice to validate resilience?
