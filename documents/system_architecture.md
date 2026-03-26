# System Architecture Diagram

## Document Purpose

This document provides a production-level system architecture view of the AI-Powered Educational Platform, derived from BRD, AVD, SOW, database schema, and project overview.

---

## 1. Comprehensive System Architecture Diagram (Mermaid)

```mermaid
flowchart TB
  subgraph actors ["External Actors"]
    user[User]
    user2[User]
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
    redis[(Redis\nRequest cache)]
    rabbitmq[(RabbitMQ\nCelery broker)]
  end

  subgraph messaging_layer ["Background / Cache"]
    celery[Celery\nIndexing Workers]
  end

  subgraph external ["External Integrations"]
    llm[LLM Provider\nGeneration API]
  end

  subgraph observability ["Observability"]
    logs[Centralized Logs\nELK/EFK / CloudWatch]
    metrics[Prometheus / Grafana]
  end

  user -->|Web usage| frontend
  user2 -->|Content operations| frontend
  frontend -->|HTTPS /api/v1| alb
  alb --> backend

  backend -->|SQL| pg
  backend -->|S3 API| minio
  backend -->|Vector search| qdrant
  backend -->|Cache| redis
  backend -->|Enqueue index task| celery
  celery -->|AMQP broker| rabbitmq
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
    U2[User]
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
    R[Redis]
  end

  subgraph L6 ["Layer 6: Background"]
    C[Celery]
  end

  U --> FE
  U2 --> FE
  FE --> GW
  GW --> B
  B --> PG
  B --> M
  B --> Q
  B --> R
  B --> C
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

### 3.3 Upload Lecture (User)

```
User -> Frontend (upload)
  -> API Gateway (POST /api/v1/lectures, Bearer token)
  -> Backend (auth module + auth check)
  -> Backend (content module: create metadata, obtain upload link / file_key)
  -> PostgreSQL (insert lecture row)
  -> MinIO (store file via frontend or signed URL)
  -> Backend (content module, confirm)
  -> Frontend (success)
```

### 3.4 Trigger Indexing (User)

```
User -> Frontend (index lecture)
  -> API Gateway (POST /api/v1/ai/lectures/{lecture_id}/index)
  -> Backend (auth module + auth check)
  -> Backend (ai/indexing module enqueues task to Celery, return job_id 202)
  -> Celery (queue, broker RabbitMQ)
  -> Celery worker (consumes task)
  -> MinIO (read lecture file)
  -> Celery worker (chunk, embed)
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

### 3.6 Summary / Quiz Generation (User)

```
Client -> API Gateway (POST /api/v1/ai/lectures/{lecture_id}/summaries or /quizzes)
  -> Auth check
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
| **Backend** | Single backend application with internal modules: **auth** (identity and token lifecycle, JWT validation, authentication checks), **content** (lecture CRUD, metadata, file_key lifecycle, lecture listings, MinIO integration), **ai** (RAG chat, summary/quiz generation, LLM integration, chat/message storage), **indexing** (publishes/consumes indexing jobs via Celery, chunking and embeddings into Qdrant). |

---

## 5. Components Summary

| Component | Type | Purpose |
|-----------|------|---------|
| **API Gateway / ALB** | Infrastructure | Single HTTPS entry, TLS termination, routing to backend service. |
| **Backend Service** | Application | Handles all auth, content, AI, and indexing logic behind a unified `/api/v1` API. |
| **PostgreSQL** | SQL Database | Users, lectures metadata, chats, messages, quiz results. Source of truth for transactional data. |
| **Qdrant** | Vector Database | Stores embeddings; semantic search for RAG context retrieval. |
| **MinIO** | Object Storage | S3-compatible store for lecture files (PDF/audio) and generated artifacts (e.g. summaries). |
| **Redis** | Cache | Request caching. |
| **RabbitMQ** | Message Broker | Broker for Celery task delivery and queue routing. |
| **Celery** | Task queue | Indexing: backend enqueues index jobs; Celery workers consume, chunk, embed, write to Qdrant. |
| **LLM Provider** | External | External AI/LLM API for text generation (summaries, quizzes, RAG answers). |

---

## 6. Assumptions Made

1. **Auth middleware**: JWT validation and authentication checks are applied at the backend service (and optionally at API Gateway). Documents state "protected endpoints" and "access control middleware"; the diagrams assume the gateway routes to a single backend service.

2. **Background workers**: Indexing is event-driven via Celery (RabbitMQ as broker); indexing logic runs as Celery workers. RAG is handled synchronously by the backend AI module (no message broker for chat).

3. **Single PostgreSQL**: Database schema describes tables for users, lectures, chats, messages. There is a single PostgreSQL instance used by different backend modules (auth: users; content: lectures; ai: chats, messages, quiz data).

4. **RAG**: RAG is synchronous. AI Service performs retrieval from Qdrant and LLM call directly within the request; no message broker is used for chat.

5. **File upload path**: Lecture "file_key" is stored by Content Service; actual file upload may be direct to MinIO (signed URL) or via Content Service. Diagram shows Content Service writing metadata to PG and MinIO as the storage backend; upload path is not refined further.

6. **Observability**: BRD/AVD mention ELK/EFK, Prometheus/Grafana, and SOW mentions CloudWatch Logs. Diagram shows a generic "Centralized Logs" and "Prometheus/Grafana" as observability; no assumption on a single vendor.

7. **LLM Provider**: Referenced as "External AI provider/model runtime" and "LLM provider integration"; no specific product (e.g. OpenAI, Bedrock). Drawn as single external "LLM Provider" box.

---

## 7. Communication Matrix

| From | To | Protocol / Mechanism |
|------|-----|------------------------|
| User | Frontend | HTTPS (browser) |
| Frontend | API Gateway | HTTPS, REST /api/v1 |
| API Gateway | Backend Service | HTTP/HTTPS (internal) |
| Backend Service | PostgreSQL | SQL (TCP) |
| Backend Service | MinIO | S3 API (HTTP) |
| Backend Service | Qdrant | gRPC/HTTP (vector API) |
| Backend | Redis | Redis protocol (cache) |
| Backend | Celery | Enqueue task |
| Celery | RabbitMQ | AMQP publish/consume |
| Backend Service | LLM Provider | HTTPS (REST or vendor API) |
| Celery Worker | RabbitMQ | AMQP consume task |
| Celery Worker | Qdrant | gRPC/HTTP |
| Celery Worker | MinIO | S3 API |
| All services | Logs / Metrics | HTTP or agent (e.g. Prometheus scrape, log forwarder) |

---

## 8. ASCII System Diagram

```
                    EXTERNAL ACTORS
    +------------------+              +------------------+
    |                 User                       |
    +-------------------+------------------------+
                        |
                        | HTTPS (web)
                        v
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
   |           Backend Service (modular)             |
   |     auth, content, ai, indexing modules         |
    +----------------------+-------------------------+
                           |
                           +---> Celery (broker RabbitMQ) ----> Celery Worker ---> Qdrant, MinIO
                           |
                           RAG: backend ai module does Qdrant + LLM synchronously (no broker)
                           |
                           v
    +------------+      +-----------+      +-----------------+
    | PostgreSQL |      |   MinIO   |      | Qdrant (vectors)|
    | users,     |      | S3-compat |      +-----------------+
    | lectures,  |      | lectures, |
    | chats,     |      | artifacts |             ^
    | messages   |      +-----------+             +-- Indexing Worker, AI Service (RAG sync)
    +------------+            ^
           ^                  |
           |                  +-- Backend Service
           +-- Backend Service

    EXTERNAL:  LLM Provider <-- Backend Service
```

---

## 9. AWS Architecture Diagram (diagrams package)

The project is designed to run on AWS (EC2/ECS/EKS per AVD). The following mapping is used for an AWS-style diagram:

| Component        | AWS / diagram element        |
|------------------|------------------------------|
| API Gateway/ALB  | ALB                          |
| Frontend         | ECS                          |
| Backend service (auth/content/ai/indexing modules) | ECS |
| Celery workers (indexing) | ECS                  |
| PostgreSQL       | RDS                          |
| MinIO            | S3 (S3-compatible)           |
| Qdrant           | GenericDatabase (vector DB)  |
| Redis (cache) | ElastiCache or self-managed Redis |
| RabbitMQ (Celery broker) | Amazon MQ (RabbitMQ) or self-managed RabbitMQ |
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

This document and the diagrams above summarize the system architecture as a production-style modular backend design with clear layers, data flows, and communication patterns derived from the project documentation.
