-- Event queue tables for distributed job processing
-- Supports process A posting jobs and process B subscribing to process them

-- Core event table: task definitions and state management
CREATE TABLE event (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'Unique event identifier',
    title VARCHAR(255) NOT NULL COMMENT 'Event title or job name',
    description TEXT COMMENT 'Detailed event description',
    tags VARCHAR(1000) NOT NULL COMMENT 'Comma-separated tags for routing',
    status ENUM('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED') NOT NULL DEFAULT 'PENDING' COMMENT 'Current event state',
    payload JSON COMMENT 'Event payload with job-specific data',
    retry_count INT NOT NULL DEFAULT 0 COMMENT 'Number of failed attempts',
    max_retries INT NOT NULL DEFAULT 3 COMMENT 'Maximum allowed retry attempts',
    next_retry_at TIMESTAMP NULL COMMENT 'When to retry after failure',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Event creation timestamp',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    INDEX idx_event_status (status),
    INDEX idx_event_tags (tags(100)),
    INDEX idx_event_next_retry_at (next_retry_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Core event table for distributed job queue system';

-- Event log: audit trail of worker execution
CREATE TABLE event_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'Log entry identifier',
    event_id BIGINT NOT NULL COMMENT 'Reference to event',
    worker_id VARCHAR(255) NOT NULL COMMENT 'Unique worker identifier',
    action VARCHAR(50) NOT NULL COMMENT 'Action performed PICKED STARTED COMPLETED FAILED',
    status_code INT COMMENT 'HTTP-like status code for outcome',
    error_message TEXT COMMENT 'Error details if action failed',
    execution_time_ms INT COMMENT 'Milliseconds taken to execute',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'When action occurred',
    FOREIGN KEY (event_id) REFERENCES event(id) ON DELETE CASCADE,
    INDEX idx_event_log_event_id (event_id),
    INDEX idx_event_log_worker_id (worker_id),
    INDEX idx_event_log_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Audit trail of worker actions on events';
