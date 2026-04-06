--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO api, public;


--
-- Name: idempotency_keys; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key text,
    method text NOT NULL,
    path text NOT NULL,
    user_id uuid,
    request_hash text,
    status text DEFAULT 'processing' NOT NULL,
    response_status integer,
    response_headers jsonb,
    response_body bytea,
    created_at timestamptz DEFAULT now() NOT NULL,
    completed_at timestamptz,
    expires_at timestamptz DEFAULT (now() + '24:00:00'::interval) NOT NULL,
    CONSTRAINT idempotency_keys_pkey PRIMARY KEY (key),
    CONSTRAINT idempotency_keys_status_check CHECK (status IN ('processing'::text, 'completed'::text, 'failed'::text))
);


COMMENT ON TABLE idempotency_keys IS 'Stores idempotency keys for safe request retries';


COMMENT ON COLUMN api.idempotency_keys.key IS 'Client-provided idempotency key (typically UUID)';


COMMENT ON COLUMN api.idempotency_keys.status IS 'processing: request in progress, completed: response cached, failed: error occurred';

--
-- Name: idx_idempotency_keys_expires_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);

--
-- Name: idx_idempotency_keys_method_path; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_method_path ON idempotency_keys (method, path);

--
-- Name: idx_idempotency_keys_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_user_id ON idempotency_keys (user_id) WHERE (user_id IS NOT NULL);

--
-- Name: idempotency_keys; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE idempotency_keys TO service_role;

