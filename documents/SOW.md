# SOW Pilot | Digital Clinical Protocols with AI Support | Freedom Cloud | Ministry of Health

## Author
- Beket Nurzhanov

## 1. Project Overview and Context
- This Scope of Work (SOW) defines the technical implementation scope for the pilot `Digital Clinical Protocols with AI Support` on Freedom Cloud for the Ministry of Health.
- The SOW is based on the target architecture and service boundaries documented in `AVD.md` and the project baseline in `project_overview.md`.
- The pilot uses a two-service architecture:
  - `frontend` (React/Next.js, TypeScript)
  - `backend` (single deployable application with internal modules: auth, content, ai, indexing)
- For domain alignment, the core content entity is interpreted as:
  - `lecture` in current APIs = `clinical protocol document` in pilot business context.

## 2. Objectives and Measurable Outcomes
### 2.1 Objectives
- Deliver a secure and observable pilot for protocol ingestion, retrieval, and AI-assisted interactions.
- Validate AI workflows for protocol Q&A, summarization, and quiz/assessment generation.
- Prove technical readiness for production scaling using the same modular backend decomposition.

### 2.2 Measurable Outcomes (Pilot KPIs)
- Availability: >= 99.0% monthly for pilot environment runtime window.
- Performance:
  - P80 synchronous API response <= 3 seconds (non-AI endpoints).
  - P80 RAG response <= 10 seconds.
- Security:
  - 100% protected endpoints require JWT authentication and access control checks.
  - 100% data in transit via HTTPS/TLS.
- Quality:
  - RAG answers include source chunk references.
  - Summary and quiz generation complete successfully for >= 90% valid inputs.

## 3. In-Scope Technical Work
### 3.1 Service Implementation Scope
- Implement and integrate:
  - identity and token lifecycle APIs in backend `auth` module.
  - protocol document lifecycle APIs in backend `content` module.
  - AI indexing and generation APIs in backend `ai` module.
  - user flows in `frontend` for protocol upload, browsing, and AI actions.

### 3.2 API Scope (from AVD/API Catalog)
- Auth APIs:
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout`
  - `GET /api/v1/auth/me`
- Content APIs:
  - `POST /api/v1/lectures`
  - `DELETE /api/v1/lectures/{lecture_id}`
  - `PATCH /api/v1/lectures/{lecture_id}`
  - `GET /api/v1/lectures/{lecture_id}/content`
  - `GET /api/v1/lectures`
- AI APIs:
  - `POST /api/v1/ai/lectures/{lecture_id}/index`
  - `POST /api/v1/ai/chat/rag`
  - `POST /api/v1/ai/lectures/{lecture_id}/summaries`
  - `POST /api/v1/ai/lectures/{lecture_id}/quizzes`

### 3.3 Data and Storage Scope
- PostgreSQL: users, content metadata, quiz records.
- Qdrant: vector embeddings and semantic retrieval index.
- RabbitMQ: Celery broker for indexing tasks.
- Redis: request caching.
- S3-compatible storage: source documents and generated artifacts.

### 3.4 Observability and Operations Scope
- Structured service logs with request and trace correlation IDs.
- Core dashboards for latency, error rate, throughput, queue/job status.
- Alert routing to operational channels for critical incidents.

## 4. Out-of-Scope Boundaries
- Full nationwide rollout and multi-tenant production hardening.
- Clinical decision automation or diagnosis recommendations.
- Integration with all external hospital/EMR systems in pilot phase.
- Dedicated mobile applications (iOS/Android native).
- Custom foundation model training/fine-tuning pipeline.

## 5. Solution Architecture and Technical Design
### 5.1 Architecture Pattern
- API-first modular backend with separate auth/content/AI concerns.
- Stateless service runtime with horizontal scaling capability.
- Managed or self-hosted data components for transactional and semantic workloads.

### 5.2 Logical Flow
1. User authenticates via backend `auth` module and receives JWT tokens.
2. User uploads protocol content via backend `content` module.
3. User accesses available lectures via backend `content` module.
4. AI indexing job stores embeddings in Qdrant.
5. User asks protocol question via RAG API.
6. Backend `ai` module retrieves context from Qdrant and produces grounded answer.
7. Optional summary/quiz outputs generated and stored.

### 5.3 Deployment Model
- Environments: `stage` and `prod`.
- Public entry: ALB/API gateway over HTTPS.
- Services and data layers deployed in private subnets/network zones where applicable.
- CI/CD-driven deployments with release promotion rules.

## 6. Workstreams and Deliverables
### Workstream A: Foundation and Platform Setup
- Deliverables:
  - repository and branch strategy
  - baseline service templates
  - environment configuration templates

### Workstream B: Authentication and Access
- Deliverables:
  - JWT auth flows
  - access control middleware for protected endpoints
  - session/token lifecycle APIs

### Workstream C: Content Management
- Deliverables:
  - protocol lifecycle APIs (add/delete/rename/get/list)
  - metadata persistence and storage integration
  - basic validation and error handling

### Workstream D: AI Workflows
- Deliverables:
  - vector indexing endpoint and background jobs
  - RAG chat API with source references
  - summary and quiz generation endpoints

### Workstream E: Frontend Integration
- Deliverables:
  - authentication screens and session handling
  - protocol management UI flows
  - AI interaction views (chat, summary, quiz)

### Workstream F: Observability and Delivery
- Deliverables:
  - monitoring dashboards
  - alerting rules and channels
  - CI/CD pipeline with test and release gates

## 7. Implementation Approach and Timeline
### Phase 1 (Weeks 1-2): Foundation
- Environment setup, CI/CD baseline, service skeletons, schema draft.

### Phase 2 (Weeks 3-4): Core APIs
- Auth and content APIs complete with integration tests.

### Phase 3 (Weeks 5-6): AI Integration
- Indexing pipeline, RAG endpoint, summary/quiz endpoints.

### Phase 4 (Weeks 7-8): Frontend + UAT Readiness
- End-to-end frontend integration, pilot test cycles, hardening fixes.

## 8. Roles and Responsibilities (RACI-Style)
- Product Owner (MoH): Accept scope, prioritize pilot goals, approve sign-off criteria.
- Project Manager: Timeline governance, risk tracking, stakeholder reporting.
- Solution Architect: Architecture integrity, NFR alignment, technical decisions.
- Security Engineer: Auth controls, encryption strategy, audit requirements.
- Backend Engineers: Auth/content APIs, integration contracts, reliability fixes.
- AI Engineer: Retrieval and generation quality, prompt and indexing tuning.
- Frontend Engineer: User workflows, API integration, UX consistency.
- DevOps Engineer: IaC, CI/CD, observability stack, release orchestration.
- QA Engineer: Test planning, execution, regression and acceptance evidence.

## 9. Environments, Infrastructure, and Access Dependencies
- Required environments:
  - `stage`: integration and UAT preparation.
  - `prod` (pilot): controlled pilot usage.
- Access dependencies:
  - cloud project/account access and IAM roles.
  - container registry and secret management access.
  - controlled access to logs, dashboards, and alert channels.

## 10. Integration and Data Requirements
### 10.1 Internal Service Integration
- `frontend` -> `backend` via `/api/v1`.
- `backend ai/indexing modules` -> Qdrant, Celery (for indexing), RabbitMQ (broker), Redis (request cache), object storage, and metadata store.
- `backend content module` -> object storage and PostgreSQL (`lectures` and related metadata).

### 10.2 Data Requirements
- Mandatory metadata fields: title, owner, language, created timestamp, version status.
- Content ingestion supports document-oriented payloads (PDF and text-first formats in pilot).
- Retention and deletion policies must support controlled purge for pilot data lifecycle.

## 11. Security, Compliance, and Risk Controls
### 11.1 Security Controls
- JWT access/refresh token model with expiration and refresh rotation.
- Access control middleware on all protected routes for authenticated users.
- HTTPS/TLS on all external and service communication paths.
- Encryption at rest for object and database storage.
- Audit logging for privileged actions and auth events.

### 11.2 Compliance and Privacy Assumptions
- Pilot data is handled with least-privilege access and environment isolation.
- Any handling of sensitive clinical data requires approved governance policy and legal review.

### 11.3 Risk Register (Pilot Technical)
- Retrieval quality risk due to poor source documents.
  - Mitigation: ingestion validation and chunking optimization.
- AI latency risk under peak load.
  - Mitigation: caching, async processing, autoscaling plan.
- Integration drift risk between frontend and backend contracts.
  - Mitigation: versioned API contracts and automated integration tests.

## 12. Quality Assurance and Testing Strategy
- Unit testing on core service logic and validation layers.
- API contract tests for auth/content/AI endpoints.
- Integration tests for storage, vector indexing, and RAG flow.
- End-to-end tests for critical user journeys.
- Non-functional tests:
  - latency/load tests for target concurrent usage.
  - security tests for auth and authorization flows.

## 13. Acceptance Criteria and Sign-Off Process
### 13.1 Technical Acceptance Criteria
- All in-scope APIs implemented and documented.
- Auth and access control enforced across protected resources.
- Protocol indexing and RAG chat operational with source citation.
- Summary and quiz generation available for indexed protocols.
- Monitoring/alerting live with agreed baseline thresholds.
- CI/CD pipeline executes tests and deployment gates successfully.

### 13.2 UAT and Sign-Off
- UAT runbook executed with predefined scenarios and evidence capture.
- Open critical defects at sign-off milestone: 0.
- Formal sign-off stakeholders:
  - Product Owner (MoH)
  - Delivery Lead (Freedom Cloud)
  - Solution Architect

## 14. Assumptions, Constraints, and Dependencies
### Assumptions
- Required cloud infrastructure and access are available on schedule.
- Pilot user cohort and sample protocol dataset are provided by stakeholders.
- AI provider and vector DB are available in target environment.

### Constraints
- Pilot timeline and budget boundaries.
- Scope limited to web-based flows and selected integrations only.

### Dependencies
- Security/compliance approval checkpoints.
- Availability of subject matter reviewers for AI quality validation.

## 15. Governance, Reporting, and Communication Cadence
- Weekly delivery checkpoint with status, blockers, and risk updates.
- Bi-weekly architecture and quality review.
- Decision log maintained for scope and architecture changes.
- Escalation path defined for critical production-like incidents in pilot.

## 16. Change Control and Issue Management
- All scope changes must be submitted through change requests.
- Each request must include impact on timeline, cost, and technical risk.
- Change approval required from Product Owner and Delivery Lead.
- Defect triage model:
  - Severity 1: immediate response and hotfix workflow.
  - Severity 2: fixed in current sprint window.
  - Severity 3: backlog prioritization.

## 17. Support Model and Handover
- Pilot support window with L1/L2 operational coverage model.
- Handover package includes:
  - architecture and deployment documentation
  - API specification and examples
  - runbooks (operations, incident, rollback)
  - known limitations and enhancement backlog


### 18.1 Reference Technical Stack
- Frontend: React/Next.js, TypeScript.
- Services: Go + FastAPI.
- Data: PostgreSQL, Qdrant, Celery (indexing), RabbitMQ (Celery broker), Redis (request cache), MinIO (S3-compatible object storage).
- Platform: containerized deployment on AWS-compatible topology.
