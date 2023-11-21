SET SCHEMA 'event_store';

CREATE TABLE IF NOT EXISTS events (
  event_id UUID PRIMARY KEY,
  event_type TEXT NOT NULL,
  event_issued_by TEXT NOT NULL,
  event_issued_at TIMESTAMP NOT NULL,

  aggregate_id UUID NOT NULL,
  aggregate_type TEXT NOT NULL,
  aggregate_version INT NOT NULL,

  event_data JSONB
);

CREATE INDEX IF NOT EXISTS events_aggregate_id_idx ON events (aggregate_id);
CREATE INDEX IF NOT EXISTS events_aggregate_type_idx ON events (aggregate_type);
