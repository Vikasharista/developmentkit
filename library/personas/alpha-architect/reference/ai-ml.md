# AI/ML System Architecture

Read when the system includes AI/ML capabilities (recommendations, search, NLP,
computer vision, generative AI).

- **Model serving** — Separate model inference from application logic. Options: dedicated inference service (TorchServe, Triton, vLLM), serverless inference (SageMaker, Vertex AI), or embedded model (ONNX Runtime) for edge/low-latency.
- **Feature stores** — Use a feature store (Feast, Tecton, or cloud-native) when features are shared across models or when online/offline feature consistency matters. Overkill for a single model — be pragmatic.
- **Vector databases** — For semantic search, RAG, or recommendation systems: pgvector (for simpler workloads), or a dedicated store (Pinecone, Weaviate, Qdrant) at scale. Design for index rebuild and embedding model versioning.
- **RAG pipelines** — Retrieval-Augmented Generation: document ingestion → chunking → embedding → vector store → retrieval → LLM generation. Design for: chunk size tuning, metadata filtering, re-ranking, citation/provenance tracking.
- **LLM integration patterns** — Route LLM calls through a gateway with provider fallback. Implement: token budget management, prompt versioning, response caching, content filtering, and cost tracking per request. (When building on Claude/Anthropic, consult the `claude-api` skill for current model IDs, pricing, and patterns — do not rely on memory.)
- **ML pipelines** — Training pipelines are separate from serving. Use versioned datasets, experiment tracking (MLflow, W&B), and model registries. Every model in production has a rollback path.
- **Feedback loops** — Design for continuous improvement: capture user interactions with AI outputs, build evaluation datasets, measure drift. A model without a feedback loop is a model that decays.

## Review gate

- [ ] Is model inference decoupled from application logic?
- [ ] Can models be rolled back independently of application deployments?
- [ ] Is there a strategy for embedding/model versioning and re-indexing?
- [ ] Are LLM API calls wrapped with fallback, retry, token budgeting, and cost tracking?
- [ ] Is there a feedback loop to capture quality signals from users?
- [ ] Are AI outputs logged for evaluation, debugging, and compliance?
- [ ] Is there a content safety / guardrails layer for generative AI outputs?
