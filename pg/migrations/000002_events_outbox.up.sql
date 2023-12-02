CREATE TABLE IF NOT EXISTS events_outbox (
  event_id UUID PRIMARY KEY REFERENCES events (event_id),
  published BOOLEAN NOT NULL DEFAULT FALSE,
  aggregate_type VARCHAR(255) NOT NULL,
  aggregate_version INT NOT NULL
);

CREATE INDEX idx_events_outbox_unpublished ON events_outbox (published) WHERE published = FALSE;
CREATE INDEX idx_events_outbox_aggregate_version ON events_outbox (aggregate_version);
