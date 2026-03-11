-- 000001_create_users_table.up.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       username VARCHAR(255) UNIQUE NOT NULL,
                       email TEXT UNIQUE NOT NULL,
                       first_name VARCHAR(255) NOT NULL,
                       last_name VARCHAR(255) NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       refresh_token TEXT,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);