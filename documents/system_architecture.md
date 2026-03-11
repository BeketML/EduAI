# System Architecture Diagram

## Document Purpose

This document provides a production-level system architecture view of the AI-Powered Educational Platform, derived from BRD, AVD, SOW, database schema, and project overview.

---

## 1. Comprehensive System Architecture Diagram (Mermaid)

```mermaid
flowchart TB
  subgraph actors ["External Actors"]
    user[User]
    operator[Platform Operator]
  end

  subgraph frontend_layer ["Frontend Layer"]
    frontend[Frontend Web App\nReact/Next.js, TypeScript]
  end

  subgraph gateway_layer ["Gateway / Infrastructure"]
    alb[API Gateway / ALB\nHTTPS]
  end

  subgraph backend_layer ["Backend Layer"]
    backend[Backend Service\n(modular: auth, content, ai, indexing)]
  end

  subgraph data_layer ["Data & Storage Layer"]
    pg[(PostgreSQL\nusers, lectures, chats,\nmessages, quiz records)]
    qdrant[(Qdrant\nVector DB / Embeddings)]
    minio[(MinIO\nS3-compatible Object Storage)]
  end

  subgraph messaging_layer ["Messaging (Indexing Only)"]
    kafka[Kafka\nIndexing Events]
  end

  subgraph external ["External Integrations"]
    llm[LLM Provider\nGeneration API]
  end

  subgraph observability ["Observability"]
    logs[Centralized Logs\nELK/EFK / CloudWatch]
    metrics[Prometheus / Grafana]
  end

  user -->|Web usage| frontend
  operator -->|Content operations| frontend
  frontend -->|HTTPS /api/v1| alb
  alb --> backend

  backend -->|SQL| pg
  backend -->|S3 API| minio
  backend -->|Vector search| qdrant
  backend -->|Publish index job| kafka
  backend -->|Generate| llm

  backend -.->|Logs/Metrics| logs
  logs --> metrics
```

---

## 2. Layered View (Simplified)

```mermaid
flowchart LR
  subgraph L1 ["Layer 1: Clients"]
    U[User]
    PO[Platform Operator]
  end

  subgraph L2 ["Layer 2: Frontend"]
    FE[Frontend]
  end

  subgraph L3 ["Layer 3: Gateway"]
    GW[ALB / API Gateway]
  end

  subgraph L4 ["Layer 4: Backend Service"]
    B[Backend\n(auth, content, ai, indexing)]
  end

  subgraph L5 ["Layer 5: Data"]
    PG[(PostgreSQL)]
    Q[(Qdrant)]
    M[(MinIO)]
  end

  subgraph L6 ["Layer 6: Messaging"]
    K[Kafka]
  end

  U --> FE
  PO --> FE
  FE --> GW
  GW --> B
  B --> PG
  B --> M
  B --> Q
  B --> K
```

---

## 3. Data Flow: Typical Request Flows

### 3.1 User Login

```
User -> Frontend (login form)
  -> API Gateway (POST /api/v1/auth/login)
  -> Backend (auth module)
  -> PostgreSQL (validate user, session)
  -> Backend (auth module, issue JWT access + refresh)
  -> Frontend (store tokens, redirect)
```

### 3.2 List Lectures (Authenticated)

```
User -> Frontend (lecture list)
  -> API Gateway (GET /api/v1/lectures, Bearer token)
  -> Backend (auth module validates JWT)
  -> Backend (content module)
  -> PostgreSQL (lectures + metadata)
  -> Backend (content module, paginated response)
  -> Frontend (render list)
```

### 3.3 Upload Lecture (Platform Operator)

```
Operator -> Frontend (upload)
  -> API Gateway (POST /api/v1/lectures, Bearer token)
  -> Backend (auth module + RBAC, operator role)
  -> Backend (content module: create metadata, obtain upload link / file_key)
  -> PostgreSQL (insert lecture row)
  -> MinIO (store file via frontend or signed URL)
  -> Backend (content module, confirm)
  -> Frontend (success)
```

### 3.4 Trigger Indexing (Platform Operator)

```
Operator -> Frontend (index lecture)
  -> API Gateway (POST /api/v1/ai/lectures/{lecture_id}/index)
  -> Backend (auth module + RBAC, operator)
  -> Backend (ai/indexing module publishes event to Kafka, return job_id 202)
  -> Kafka (topic)
  -> Indexing Worker (consume)
  -> MinIO (read lecture file)
  -> Indexing Worker (chunk, embed)
  -> Qdrant (store vectors)
  -> (optional) PostgreSQL (job status)
  -> Frontend (poll job status or webhook)
```

### 3.5 RAG Chat (User)

```
User -> Frontend (send message)
  -> API Gateway (POST /api/v1/ai/chat/rag, Bearer token)
  -> Auth (validate JWT)
  -> Backend (ai module, synchronous: retrieve context from Qdrant, call LLM, build response)
  -> Qdrant (semantic search for context)
  -> LLM Provider (generate answer with context)
  -> Backend (ai module, persist message, return response)
  -> PostgreSQL (messages insert)
  -> Frontend (display answer + source_documents)
```

### 3.6 Summary / Quiz Generation (Operator or User)

```
Client -> API Gateway (POST /api/v1/ai/lectures/{lecture_id}/summaries or /quizzes)
  -> Auth + RBAC
  -> Backend (ai module)
  -> (optional) MinIO / Content Service (lecture content)
  -> LLM Provider (generate summary or quiz)
  -> Backend (ai module, return payload; quiz may be stored in PG)
  -> Frontend (display)
```

---

## 4. Service Roles Summary

| Service | Role |
|--------|------|
| **Frontend** | SPA/SSR UI (React/Next.js). Handles auth token storage, routing, lecture list/detail, RAG chat UI, summary/quiz views. Calls backend via `/api/v1`. |
| **Backend** | Single backend application with internal modules: **auth** (identity and token lifecycle, JWT validation, RBAC), **content** (lecture CRUD, metadata, file_key lifecycle, lecture listings, MinIO integration), **ai** (RAG chat, summary/quiz generation, LLM integration, chat/message storage), **indexing** (publishes/consumes indexing jobs via Kafka, chunking and embeddings into Qdrant). |

---

## 5. Components Summary

| Component | Type | Purpose |
|-----------|------|---------|
| **API Gateway / ALB** | Infrastructure | Single HTTPS entry, TLS termination, routing to backend service. |
| **Backend Service** | Application | Handles all auth, content, AI, and indexing logic behind a unified `/api/v1` API. |
| **PostgreSQL** | SQL Database | Users, lectures metadata, chats, messages, quiz results. Source of truth for transactional data. |
| **Qdrant** | Vector Database | Stores embeddings; semantic search for RAG context retrieval. |
| **MinIO** | Object Storage | S3-compatible store for lecture files (PDF/audio) and generated artifacts (e.g. summaries). |
| **Kafka** | Message Broker | Used only for indexing: backend indexing module publishes index jobs to a topic; background indexing logic consumes and processes (chunk, embed, write to Qdrant). |
| **LLM Provider** | External | External AI/LLM API for text generation (summaries, quizzes, RAG answers). |

---

## 6. Assumptions Made

1. **Auth middleware**: JWT validation and RBAC are applied at the backend service (and optionally at API Gateway). Documents state "protected endpoints" and "access control middleware"; the diagrams assume the gateway routes to a single backend service, and that backend modules enforce JWT and roles.

2. **Background workers**: Indexing is event-driven via Kafka; indexing logic runs as background jobs of the backend (indexing module). RAG is handled synchronously by the backend AI module (no message broker for chat).

3. **Single PostgreSQL**: Database schema describes tables for users, lectures, chats, messages. There is a single PostgreSQL instance used by different backend modules (auth: users; content: lectures; ai: chats, messages, quiz data).

4. **RAG**: RAG is synchronous. AI Service performs retrieval from Qdrant and LLM call directly within the request; no message broker is used for chat.

5. **File upload path**: Lecture "file_key" is stored by Content Service; actual file upload may be direct to MinIO (signed URL) or via Content Service. Diagram shows Content Service writing metadata to PG and MinIO as the storage backend; upload path is not refined further.

6. **Observability**: BRD/AVD mention ELK/EFK, Prometheus/Grafana, and SOW mentions CloudWatch Logs. Diagram shows a generic "Centralized Logs" and "Prometheus/Grafana" as observability; no assumption on a single vendor.

7. **LLM Provider**: Referenced as "External AI provider/model runtime" and "LLM provider integration"; no specific product (e.g. OpenAI, Bedrock). Drawn as single external "LLM Provider" box.

---

## 7. Communication Matrix

| From | To | Protocol / Mechanism |
|------|-----|------------------------|
| User / Operator | Frontend | HTTPS (browser) |
| Frontend | API Gateway | HTTPS, REST /api/v1 |
| API Gateway | Auth Service | HTTP/HTTPS (internal) |
| API Gateway | Content Service | HTTP/HTTPS (internal) |
| API Gateway | AI Service | HTTP/HTTPS (internal) |
| Auth Service | PostgreSQL | SQL (TCP) |
| Content Service | PostgreSQL | SQL (TCP) |
| Content Service | MinIO | S3 API (HTTP) |
| AI Service | PostgreSQL | SQL (TCP) |
| AI Service | Qdrant | gRPC/HTTP (vector API) |
| AI Service | MinIO | S3 API (HTTP) |
| AI Service | Kafka | Kafka protocol (produce, indexing only) |
| AI Service | LLM Provider | HTTPS (REST or vendor API) |
| Indexing Worker | Kafka | Kafka protocol (consume) |
| Indexing Worker | Qdrant | gRPC/HTTP |
| Indexing Worker | MinIO | S3 API |
| All services | Logs / Metrics | HTTP or agent (e.g. Prometheus scrape, log forwarder) |

---

## 8. ASCII System Diagram

```
                    EXTERNAL ACTORS
    +------------------+              +------------------+
    |      User        |              | Platform Operator |
    +--------+---------+              +--------+---------+
             |                                  |
             | HTTPS (web)                      | HTTPS (web)
             v                                  v
    +----------------------------------------------+
    |         Frontend (React/Next.js, TS)          |
    +----------------------------------------------+
             |
             | HTTPS /api/v1
             v
    +----------------------------------------------+
    |            API Gateway / ALB                  |
    +--------+------------------+-------------------+
             |                  |                   |
             v                  v                   v
    +-------------+   +----------------+   +----------------+
    |Auth Service |   |Content Service |   |   AI Service   |
    |    (Go)     |   |   (FastAPI)     |   |   (FastAPI)    |
    +------+------+   +--------+-------+   +--------+-------+
           |                   |                     |
           |                   |                     +---> Kafka ----> Indexing Worker ---> Qdrant, MinIO
           |                   |                     |
           |                   |                     RAG: AI Service does Qdrant + LLM synchronously (no broker)
           |                   |                     |
           v                   v                     v
    +------------+      +-----------+      +-----------------+
    | PostgreSQL |      |   MinIO   |      | Qdrant (vectors)|
    | users,     |      | S3-compat |      +-----------------+
    | lectures,  |      | lectures, |
    | chats,     |      | artifacts |             ^
    | messages   |      +-----------+             +-- Indexing Worker, AI Service (RAG sync)
    +------------+            ^
           ^                  |
           |                  +-- Content Service, AI Service
           +-- Auth, Content, AI Service

    EXTERNAL:  LLM Provider <-- AI Service
```

---

## 9. AWS Architecture Diagram (diagrams package)

The project is designed to run on AWS (EC2/ECS/EKS per AVD). The following mapping is used for an AWS-style diagram:

| Component        | AWS / diagram element        |
|------------------|------------------------------|
| API Gateway/ALB  | ALB                          |
| Frontend         | ECS                          |
| Auth/Content/AI  | ECS                          |
| Indexing Worker  | ECS                          |
| PostgreSQL       | RDS                          |
| MinIO            | S3 (S3-compatible)           |
| Qdrant           | GenericDatabase (vector DB)  |
| Kafka            | ManagedStreamingForKafka     |
| LLM              | Bedrock (or external API)    |
| Logs             | CloudWatch Logs              |

To generate the PNG diagram locally (requires Linux/macOS or WSL; `signal.SIGALRM` is not available on Windows):

```bash
pip install diagrams
# from repo root
python scripts/generate_aws_diagram.py
```

Output: `generated-diagrams/eduai-aws-architecture.png`. The script is [scripts/generate_aws_diagram.py](../scripts/generate_aws_diagram.py).

---

This document and the diagrams above summarize the system architecture as a production-style microservices design with clear layers, data flows, and communication patterns derived from the project documentation.
