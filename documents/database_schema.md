# Database Schema

## Goal
DB schema for backend modules:
- `auth` module: users + login data
- `content` module: lecture metadata
- `ai` module: chat history per user

Redis is used elsewhere for request caching; RabbitMQ is used as Celery broker (no tables in this schema).

---

## 1) Tables

### `users` (auth module)
- `user_id` UUID (PK)
- `email` TEXT (UNIQUE, NOT NULL)
- `first_name` VARCHAR(255) (NOT NULL)
- `last_name` VARCHAR(255) (NOT NULL)
- `password_hash` VARCHAR(255) (NOT NULL)

### `lectures` (content module)
- `lecture_id` UUID (PK)
- `user_id` UUID (FK -> `users.user_id`, NOT NULL)
- `lecture_title` TEXT (NOT NULL)
- `language` TEXT (NOT NULL, default: `en`)
- `file_key` TEXT (NOT NULL, path/key in object storage)
- `created_date` BIGINT (NOT NULL, unix timestamp)
- `deleted` SMALLINT (NOT NULL, `0/1`)

### `chats` (ai module)
- `chat_id` UUID (PK)
- `user_id` UUID (FK -> `users.user_id`, NOT NULL)
- `chat_title` TEXT (NOT NULL)
- `created_date` BIGINT (NOT NULL, unix timestamp)
- `last_modified_date` BIGINT (NOT NULL, unix timestamp)
- `deleted` SMALLINT (NOT NULL, `0/1`)

### `messages` (ai module)
- `message_id` UUID (PK)
- `chat_id` UUID (FK -> `chats.chat_id`, NOT NULL)
- `created_date` TIMESTAMPTZ (NOT NULL)
- `last_modified_date` TIMESTAMPTZ (NOT NULL)
- `is_user` BOOLEAN (NOT NULL)
- `content` JSONB (NOT NULL)
- `source` TEXT (`web` or `script`)

This is enough for:
- registration/login and storing user info
- storing lecture metadata
- one user -> many chats
- each chat -> many messages (user query + AI response)

---

## 2) Minimal PostgreSQL SQL

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE lectures (
  lecture_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  lecture_title TEXT NOT NULL,
  language TEXT NOT NULL DEFAULT 'en',
  file_key TEXT NOT NULL,
  created_date BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM now())::BIGINT),
  deleted SMALLINT NOT NULL DEFAULT 0 CHECK (deleted IN (0, 1))
);

CREATE TABLE chats (
  chat_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  chat_title TEXT NOT NULL,
  created_date BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM now())::BIGINT),
  last_modified_date BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM now())::BIGINT),
  deleted SMALLINT NOT NULL DEFAULT 0 CHECK (deleted IN (0, 1))
);

CREATE TABLE messages (
  message_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chat_id UUID NOT NULL REFERENCES chats(chat_id) ON DELETE CASCADE,
  created_date TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_modified_date TIMESTAMPTZ NOT NULL DEFAULT now(),
  is_user BOOLEAN NOT NULL,
  content JSONB NOT NULL,
  source TEXT CHECK (source IN ('web', 'script'))
);

CREATE INDEX idx_lectures_user_id ON lectures(user_id);
CREATE INDEX idx_chats_user_id ON chats(user_id);
CREATE INDEX idx_messages_chat_id_created_date ON messages(chat_id, created_date);
```

---

## 3) `messages.content` JSON format (from ERD)

### If `is_user = true` (`UserContent`)
```json
{
  "text": "Explain matrix rank in simple terms"
}
```

### If `is_user = false` (`LLMContent`)
```json
{
  "text": "Matrix rank is the number of linearly independent rows or columns...",
  "source_documents": [
    "chunk_42",
    "chunk_17"
  ]
}
```

---

## 4) Authorization assumptions (single-role access)
- Authenticated users can read/write lecture content and use own chat interactions with AI.
- Ownership check:
  - content write actions are available to authenticated users.
  - chat/message actions are allowed only when `chats.user_id = token.user_id`.

---

## 5) Quiz storage (future or out of scope)

- The current schema does not define a dedicated table for quiz results or generated quizzes.
- Other documents (AVD, BRD, system_architecture) mention "quiz records" in PostgreSQL; this may refer to a future table or to quiz payloads returned by the API without persistent storage in MVP.
- If persistent quiz results are required, a separate table (e.g. `quizzes` or `quiz_results`) can be added to PostgreSQL; otherwise quiz data can remain API-only for the pilot.
