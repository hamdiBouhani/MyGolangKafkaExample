-- Create Debezium replication user
CREATE USER debezium WITH REPLICATION LOGIN PASSWORD 'debezium';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO debezium;

-- Sample table for CDC
CREATE TABLE IF NOT EXISTS user_events (
    id          BIGSERIAL PRIMARY KEY,
    user_id     TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO user_events (user_id, event_type) VALUES
    ('user-123', 'LOGIN'),
    ('user-456', 'LOGOUT'),
    ('user-789', 'SIGNUP');