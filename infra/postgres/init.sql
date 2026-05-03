-- All Phase 0 tables. Created by docker-entrypoint on first container start.

CREATE TABLE IF NOT EXISTS sessions (
    id                       text         PRIMARY KEY,
    user_id                  text         NOT NULL,
    agent_name               text         NOT NULL,
    platform_conversation_id text,                                 -- nullable: Mock has no remote conv
    created_at               timestamptz  NOT NULL DEFAULT now(),
    ended_at                 timestamptz
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS messages (
    id          text         PRIMARY KEY,
    session_id  text         NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role        text         NOT NULL CHECK (role IN ('user','assistant')),
    text        text         NOT NULL,
    created_at  timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_session ON messages (session_id, created_at);

CREATE TABLE IF NOT EXISTS audit_log (
    id           bigserial    PRIMARY KEY,
    session_id   text         NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    run_id       text         NOT NULL,
    kind         text         NOT NULL,                    -- canonical event type
    occurred_at  timestamptz  NOT NULL DEFAULT now(),
    payload      jsonb        NOT NULL                     -- the event JSON
);

CREATE INDEX IF NOT EXISTS idx_audit_session ON audit_log (session_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_audit_run     ON audit_log (run_id);

CREATE TABLE IF NOT EXISTS agent_registry (
    name         text   PRIMARY KEY,
    version      text   NOT NULL,
    description  text,
    enabled      boolean NOT NULL DEFAULT true,
    adapter_kind text   NOT NULL,             -- e.g. 'mock', 'goclaw'
    config       jsonb  NOT NULL DEFAULT '{}'  -- per-adapter settings (base_url, api_key_ref, ...)
);

-- Phase 0 seed: one Mock-backed agent.
-- scenario_dir is resolved relative to the server's working directory; run from backend/.
INSERT INTO agent_registry (name, version, description, enabled, adapter_kind, config) VALUES
    ('mock-trino-flavored', '0.1.0', 'Scripted Trino-flavored conversations for MVP smoke',
     true, 'mock', '{"scenario_dir":"internal/adapters/mock/scenarios"}')
ON CONFLICT (name) DO NOTHING;
