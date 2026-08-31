-- Identity core: accounts and refresh-token sessions.
-- A single account can act as both freelancer and client; capabilities are
-- explicit boolean flags rather than a role enum.

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             CITEXT NOT NULL UNIQUE,
    password_hash     TEXT NOT NULL,
    display_name      TEXT NOT NULL DEFAULT '',
    avatar_url        TEXT NOT NULL DEFAULT '',
    country           TEXT NOT NULL DEFAULT '',
    timezone          TEXT NOT NULL DEFAULT 'UTC',
    is_freelancer     BOOLEAN NOT NULL DEFAULT FALSE,
    is_client         BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin          BOOLEAN NOT NULL DEFAULT FALSE,
    status            TEXT NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'suspended', 'deleted')),
    email_verified_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Refresh-token sessions. The access token is a short-lived JWT held only by the
-- client; the refresh token is stored hashed so a DB leak cannot mint sessions.
CREATE TABLE auth_sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash    TEXT NOT NULL UNIQUE,
    user_agent    TEXT NOT NULL DEFAULT '',
    ip            TEXT NOT NULL DEFAULT '',
    expires_at    TIMESTAMPTZ NOT NULL,
    revoked_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
