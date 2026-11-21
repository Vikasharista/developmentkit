# Security Architecture

Read when evaluating or designing security posture — a cross-cutting concern for
every system.

- **Zero Trust model** — Never trust, always verify. Every service-to-service call is authenticated and authorised, regardless of network boundary. No implicit trust based on being "inside the VPC."
- **Identity & access management** — OAuth2 + OpenID Connect for user-facing auth. Short-lived tokens with refresh rotation. Service-to-service auth via mTLS or signed JWTs. Centralise identity (Keycloak, Auth0, Cognito) — do not roll your own.
- **API security** — Rate limiting, throttling, and request size limits at the gateway. Input validation at system boundaries. CORS policies locked to known origins. API keys for external consumers, OAuth scopes for internal.
- **Encryption** — Data at rest: AES-256 or cloud-native KMS (AWS KMS, GCP Cloud KMS, Azure Key Vault). Data in transit: TLS 1.2+ everywhere, mTLS for service mesh. No secrets in code, config files, or environment variables without a secrets manager.
- **Secret management** — HashiCorp Vault, AWS Secrets Manager, or GCP Secret Manager. Secrets are rotated on a schedule. Applications fetch secrets at runtime, never at build time. Audit access to secrets.
- **Defence in depth** — Layer security: WAF at the edge, network segmentation (security groups, network policies), application-level auth, row-level data access controls, audit logging. No single layer is the only line of defence.
- **Container & orchestration security** — Non-root containers, read-only filesystems, minimal base images, vulnerability scanning in CI (Trivy, Snyk). Kubernetes: Pod Security Standards, RBAC, network policies, no privileged pods.
- **Compliance-aware design** — Tag PII fields at the schema level. Design for GDPR right-to-erasure, data residency, and audit trail requirements from day one. Retrofitting compliance is 10x more expensive.

> For a deep security audit (threat modelling, exploit-level review), hand off to
> the `zeta` security persona — Alpha sets posture, Zeta hardens it.

## Review gate

- [ ] Is the system designed on Zero Trust principles (no implicit trust from network location)?
- [ ] Is authentication centralised (OAuth2/OIDC) with short-lived tokens?
- [ ] Are secrets managed via a dedicated secrets manager (not env vars or config files)?
- [ ] Is data encrypted at rest and in transit (TLS 1.2+, AES-256/KMS)?
- [ ] Are API endpoints protected with rate limiting, input validation, and size limits?
- [ ] Is there defence in depth (WAF + network segmentation + app-level auth + audit logging)?
- [ ] Are containers running as non-root with minimal base images and vulnerability scanning?
- [ ] Are PII and compliance requirements (GDPR, HIPAA) designed in, not bolted on?
