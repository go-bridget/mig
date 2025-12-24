-- Event queue tables for distributed job processing
-- Supports process A posting jobs and process B subscribing to process them

-- Core event table: task definitions and state management
CREATE TABLE event (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    tags TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED')),
    payload TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Event log: audit trail of worker execution
CREATE TABLE event_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL,
    worker_id TEXT NOT NULL,
    action TEXT NOT NULL,
    status_code INTEGER,
    error_message TEXT,
    execution_time_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES event(id) ON DELETE CASCADE
);

-- Indexes for common queries and foreign key relationships
CREATE INDEX idx_event_status ON event(status);
CREATE INDEX idx_event_tags ON event(tags);
CREATE INDEX idx_event_next_retry_at ON event(next_retry_at);
-- Foreign key indexes
CREATE INDEX idx_event_log_event_id ON event_log(event_id);
CREATE INDEX idx_event_log_worker_id ON event_log(worker_id);
CREATE INDEX idx_event_log_created_at ON event_log(created_at);
