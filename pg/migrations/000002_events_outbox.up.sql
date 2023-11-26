CREATE TABLE IF NOT EXISTS events_outbox (
  event_id UUID PRIMARY KEY REFERENCES events (event_id),
  published BOOLEAN NOT NULL DEFAULT FALSE,
  aggregate_version INT NOT NULL
);
