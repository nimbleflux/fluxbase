--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO system, public;


--
-- Name: rate_limits; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS rate_limits (
    key text,
    count bigint DEFAULT 1 NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT rate_limits_pkey PRIMARY KEY (key)
);


COMMENT ON TABLE rate_limits IS 'Distributed rate limiting storage for multi-instance deployments';


COMMENT ON COLUMN system.rate_limits.key IS 'Rate limit key (e.g., "login:192.168.1.1")';


COMMENT ON COLUMN system.rate_limits.count IS 'Number of requests in the current window';


COMMENT ON COLUMN system.rate_limits.expires_at IS 'When this rate limit window expires';

--
-- Name: idx_rate_limits_expires_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rate_limits_expires_at ON rate_limits (expires_at);

--
-- Name: rate_limits; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE rate_limits TO service_role;

