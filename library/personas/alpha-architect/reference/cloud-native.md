# Cloud-Native Architecture

Read when the system runs on cloud infrastructure or targets cloud-native
deployment. (For a monolith on a managed host, most of this is overkill — apply
only what the deployment target actually warrants.)

- **Container orchestration** — Kubernetes as the default for multi-service systems. Define resource requests/limits, liveness/readiness probes, and pod disruption budgets. Evaluate managed Kubernetes (EKS, GKE, AKS) vs. self-managed based on team ops maturity.
- **Infrastructure as Code** — Terraform for multi-cloud, Pulumi for teams that prefer general-purpose languages, CloudFormation/CDK for AWS-only. All infrastructure is versioned, reviewed, and deployed through CI. No click-ops in production.
- **GitOps** — Argo CD or Flux for declarative, git-driven deployments. The git repo is the source of truth for desired cluster state. Drift detection and auto-reconciliation are mandatory.
- **Multi-cloud & hybrid strategy** — Avoid cloud-native lock-in on the data and compute layers unless the trade-off is justified (e.g., DynamoDB for a specific workload). Abstract provider-specific services behind interfaces. Evaluate portability cost vs. managed service benefit.
- **Auto-scaling** — Horizontal Pod Autoscaler (HPA) for compute, KEDA for event-driven scaling. Define scaling metrics beyond CPU (queue depth, request latency, custom business metrics). Set scale-to-zero policies for non-critical workloads.
- **Resource optimisation** — Right-size instances. Use spot/preemptible for stateless batch workloads. Reserved instances or savings plans for steady-state. Monitor and alert on cost anomalies.
- **Edge & CDN integration** — Static assets and cacheable API responses at the CDN (CloudFront, Fastly, Cloudflare). Edge functions (Lambda@Edge, Cloudflare Workers) for latency-sensitive logic. Design cache keys and invalidation strategy upfront.

## Review gate

- [ ] Is all infrastructure defined as code and deployed through CI (no click-ops)?
- [ ] Are resource requests/limits, probes, and pod disruption budgets defined?
- [ ] Is there a GitOps pipeline with drift detection?
- [ ] Are scaling policies defined beyond CPU (queue depth, latency, business metrics)?
- [ ] Is cost monitored with alerts on anomalies?
- [ ] Is the cloud-lock-in trade-off explicitly evaluated and documented?
