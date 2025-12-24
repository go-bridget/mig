-- Event queue tables for distributed job processing
-- Supports process A posting jobs and process B subscribing to process them

-- Event status enumeration
CREATE TYPE event_status AS ENUM ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED');

-- Core event table: task definitions and state management
CREATE TABLE event (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    tags TEXT NOT NULL,
    status event_status NOT NULL DEFAULT 'PENDING',
    payload JSONB,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Event log: audit trail of worker execution
CREATE TABLE event_log (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL,
    worker_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    status_code INT,
    error_message TEXT,
    execution_time_ms INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES event(id) ON DELETE CASCADE
);

-- Indexes for common queries
CREATE INDEX idx_event_status ON event(status);
CREATE INDEX idx_event_tags ON event(tags);
CREATE INDEX idx_event_next_retry_at ON event(next_retry_at);
CREATE INDEX idx_event_log_event_id ON event_log(event_id);
CREATE INDEX idx_event_log_worker_id ON event_log(worker_id);
CREATE INDEX idx_event_log_created_at ON event_log(created_at);

COMMENT ON TABLE event IS 'Core event table for distributed job queue system';
COMMENT ON COLUMN event.id IS 'Unique event identifier';
COMMENT ON COLUMN event.title IS 'Event title or job name';
COMMENT ON COLUMN event.description IS 'Detailed event description';
COMMENT ON COLUMN event.tags IS 'Comma-separated tags for routing';
COMMENT ON COLUMN event.status IS 'Current event state';
COMMENT ON COLUMN event.payload IS 'Event payload with job-specific data';
COMMENT ON COLUMN event.retry_count IS 'Number of failed attempts';
COMMENT ON COLUMN event.max_retries IS 'Maximum allowed retry attempts';
COMMENT ON COLUMN event.next_retry_at IS 'When to retry after failure';
COMMENT ON COLUMN event.created_at IS 'Event creation timestamp';
COMMENT ON COLUMN event.updated_at IS 'Last update timestamp';

COMMENT ON TABLE event_log IS 'Audit trail of worker actions on events';
COMMENT ON COLUMN event_log.id IS 'Log entry identifier';
COMMENT ON COLUMN event_log.event_id IS 'Reference to event';
COMMENT ON COLUMN event_log.worker_id IS 'Unique worker identifier';
COMMENT ON COLUMN event_log.action IS 'Action performed PICKED STARTED COMPLETED FAILED';
COMMENT ON COLUMN event_log.status_code IS 'HTTP-like status code for outcome';
COMMENT ON COLUMN event_log.error_message IS 'Error details if action failed';
COMMENT ON COLUMN event_log.execution_time_ms IS 'Milliseconds taken to execute';
COMMENT ON COLUMN event_log.created_at IS 'When action occurred';
