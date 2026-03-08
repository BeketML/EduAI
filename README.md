# AI-Powered Educational Platform for Lectures

## Overview
This project is a microservices-based educational platform for lecture content management and AI-assisted learning workflows.

Core capabilities:
- secure authentication and session lifecycle
- lecture content upload, storage, listing, and retrieval
- AI indexing to vector database
- RAG chat with source-grounded answers
- summary and quiz generation

## Architecture
Services:
- `auth-service` (Go)
- `content-service` (FastAPI)
- `ai-service` (FastAPI)
- `frontend` (React/Next.js, TypeScript)

Data and infrastructure:
- PostgreSQL (transactional metadata)
- Qdrant (vector embeddings / semantic retrieval)
- MinIO (object storage)
- Kafka (indexing async flow only; RAG is synchronous in AI Service)

## Access Model
- authenticated users can read existing content and use own AI chat interactions
- restricted platform-operator actions are used for content write operations and AI generation triggers

## API Surface (high level)
Auth:
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

Content:
- `POST /api/v1/lectures`
- `PATCH /api/v1/lectures/{lecture_id}`
- `DELETE /api/v1/lectures/{lecture_id}`
- `GET /api/v1/lectures`
- `GET /api/v1/lectures/{lecture_id}/content`

AI:
- `POST /api/v1/ai/lectures/{lecture_id}/index`
- `POST /api/v1/ai/chat/rag`
- `POST /api/v1/ai/lectures/{lecture_id}/summaries`
- `POST /api/v1/ai/lectures/{lecture_id}/quizzes`

## Database
Current MVP schema is documented in:
- `documents/database_schema.md`

Main tables:
- `users`
- `lectures`
- `chats`
- `messages`

## Project Documentation
- Architecture Vision: `documents/AVD.md`
- Scope of Work: `documents/SOW.md`
- Business Requirements: `documents/BRD.md`
- Project Overview: `documents/project_overview.md`
- ERD image: `documents/ERD.drawio.png`

## Current Repository State
This repository currently contains project documentation and architecture artifacts.
Implementation repositories/modules for services can be added using the same service boundaries described above.
