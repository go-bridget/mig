# Event Queue API - Design Document & Usage Guide

## Overview

This document describes a distributed event queue system supporting asynchronous job processing. Process A publishes events with tags; Process B subscribes to tags and processes events. Worker actions are audited in `event_log` with support for automatic retry on failure.

## Data Model

### Event States

```
PENDING → PROCESSING → COMPLETED
       ↓
       FAILED → PENDING (after 5-min backoff, max 3 retries)
```

### Worker Tracking

Each worker is identified by a composite ID: `{hostname}:{process_id}`. This allows:
- Tracking which worker picked up an event
- Detecting stale workers via hostname/pid
- Correlating all actions by a single worker instance

### Retry Strategy

On failure:
1. Event status remains FAILED (read-only marker of last attempt)
2. Reset status to PENDING for reprocessing
3. Set `next_retry_at = now() + 5 minutes`
4. Increment `retry_count`
5. Worker must check `retry_count < max_retries` before picking up

---

## API Endpoints

### 1. POST /events - Create Job (Process A)

Process A posts a new event for distribution to interested workers.

**Request:**
```json
{
  "title": "Send Email Notification",
  "description": "Send welcome email to new user",
  "tags": "email,priority-high,notification",
  "payload": {
    "user_id": 12345,
    "email": "user@example.com",
    "template": "welcome",
    "variables": {
      "name": "Alice",
      "signup_date": "2025-12-24"
    }
  }
}
```

**Response (201 Created):**
```json
{
  "id": 1001,
  "title": "Send Email Notification",
  "description": "Send welcome email to new user",
  "tags": "email,priority-high,notification",
  "status": "PENDING",
  "payload": {
    "user_id": 12345,
    "email": "user@example.com",
    "template": "welcome",
    "variables": {
      "name": "Alice",
      "signup_date": "2025-12-24"
    }
  },
  "retry_count": 0,
  "max_retries": 3,
  "next_retry_at": null,
  "created_at": "2025-12-24T10:30:00Z",
  "updated_at": "2025-12-24T10:30:00Z"
}
```

---

### 2. GET /events/subscribe - Poll for Jobs (Process B)

Process B polls for events matching desired tags. Returns one event at a time with status atomically set to PROCESSING and worker_id recorded.

**Request:**
```
GET /events/subscribe?tags=email,notification&worker_id=worker-02:8742
```

**Query Parameters:**
- `tags` (required): Comma-separated tags to subscribe to. Event matches if it contains ANY tag.
- `worker_id` (required): Unique identifier for this worker (format: `{hostname}:{pid}`)

**Response (200 OK) - Event found:**
```json
{
  "id": 1001,
  "title": "Send Email Notification",
  "tags": "email,priority-high,notification",
  "status": "PROCESSING",
  "payload": {
    "user_id": 12345,
    "email": "user@example.com",
    "template": "welcome",
    "variables": {
      "name": "Alice",
      "signup_date": "2025-12-24"
    }
  },
  "retry_count": 0,
  "max_retries": 3,
  "created_at": "2025-12-24T10:30:00Z"
}
```

**Response (204 No Content) - No matching events:**
```
(empty body)
```

---

### 3. POST /events/{id}/complete - Mark Success (Process B)

Worker reports successful completion, writes audit log, and marks event complete.

**Request:**
```json
{
  "worker_id": "worker-02:8742",
  "execution_time_ms": 1250,
  "status_code": 200
}
```

**Response (200 OK):**
```json
{
  "event_id": 1001,
  "worker_id": "worker-02:8742",
  "action": "COMPLETED",
  "status_code": 200,
  "execution_time_ms": 1250,
  "created_at": "2025-12-24T10:31:15Z"
}
```

**Side Effects:**
- Event status → COMPLETED
- Event updated_at → now
- event_log entry created with action=COMPLETED

---

### 4. POST /events/{id}/fail - Report Failure & Retry (Process B)

Worker reports failure, writes audit log, and schedules retry if under max_retries.

**Request:**
```json
{
  "worker_id": "worker-02:8742",
  "execution_time_ms": 5000,
  "status_code": 500,
  "error_message": "Connection timeout: mail server unreachable. Will retry."
}
```

**Response (200 OK) - Retry scheduled:**
```json
{
  "event_id": 1001,
  "worker_id": "worker-02:8742",
  "action": "FAILED",
  "status_code": 500,
  "error_message": "Connection timeout: mail server unreachable. Will retry.",
  "execution_time_ms": 5000,
  "retry_scheduled": true,
  "next_retry_at": "2025-12-24T10:36:15Z",
  "created_at": "2025-12-24T10:31:15Z"
}
```

**Response (400 Bad Request) - Max retries exceeded:**
```json
{
  "error": "Max retries exceeded",
  "retry_count": 3,
  "max_retries": 3
}
```

**Side Effects:**
- Event status → PENDING (if retry scheduled) or FAILED (if max reached)
- Event retry_count += 1
- Event next_retry_at = now() + 5 minutes (if retry scheduled)
- event_log entry created with action=FAILED

---

### 5. GET /events/{id} - Get Event Details

Retrieve current state of an event, including full audit trail.

**Request:**
```
GET /events/1001?include_logs=true
```

**Query Parameters:**
- `include_logs` (optional, default=false): Include full event_log history

**Response (200 OK):**
```json
{
  "id": 1001,
  "title": "Send Email Notification",
  "description": "Send welcome email to new user",
  "tags": "email,priority-high,notification",
  "status": "COMPLETED",
  "payload": { ... },
  "retry_count": 0,
  "max_retries": 3,
  "next_retry_at": null,
  "created_at": "2025-12-24T10:30:00Z",
  "updated_at": "2025-12-24T10:31:15Z",
  "logs": [
    {
      "id": 5001,
      "worker_id": "worker-02:8742",
      "action": "PICKED",
      "created_at": "2025-12-24T10:30:45Z"
    },
    {
      "id": 5002,
      "worker_id": "worker-02:8742",
      "action": "COMPLETED",
      "status_code": 200,
      "execution_time_ms": 1250,
      "created_at": "2025-12-24T10:31:15Z"
    }
  ]
}
```

---

### 6. GET /events?status=PENDING&limit=10 - List Events

List events with filtering and pagination.

**Query Parameters:**
- `status` (optional): Filter by status (PENDING, PROCESSING, COMPLETED, FAILED)
- `tags` (optional): Filter by tags (comma-separated, matches if ANY match)
- `limit` (optional, default=20): Maximum events to return
- `offset` (optional, default=0): Pagination offset

**Response (200 OK):**
```json
{
  "events": [
    {
      "id": 1001,
      "title": "Send Email Notification",
      "status": "COMPLETED",
      "tags": "email,priority-high,notification",
      "retry_count": 0,
      "created_at": "2025-12-24T10:30:00Z",
      "updated_at": "2025-12-24T10:31:15Z"
    },
    {
      "id": 1002,
      "title": "Process Payment",
      "status": "PENDING",
      "tags": "payment,priority-high",
      "retry_count": 1,
      "next_retry_at": "2025-12-24T10:36:00Z",
      "created_at": "2025-12-24T10:25:00Z",
      "updated_at": "2025-12-24T10:31:00Z"
    }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0
}
```

---

## Event Log Lifecycle Example

For a successful event with one retry:

```json
[
  {
    "id": 5001,
    "event_id": 1001,
    "worker_id": "worker-01:7234",
    "action": "PICKED",
    "created_at": "2025-12-24T10:30:45Z"
  },
  {
    "id": 5002,
    "event_id": 1001,
    "worker_id": "worker-01:7234",
    "action": "STARTED",
    "created_at": "2025-12-24T10:30:46Z"
  },
  {
    "id": 5003,
    "event_id": 1001,
    "worker_id": "worker-01:7234",
    "action": "FAILED",
    "status_code": 503,
    "error_message": "Service unavailable, will retry",
    "execution_time_ms": 8000,
    "created_at": "2025-12-24T10:30:54Z"
  },
  {
    "id": 5004,
    "event_id": 1001,
    "worker_id": "worker-02:8742",
    "action": "PICKED",
    "created_at": "2025-12-24T10:36:10Z"
  },
  {
    "id": 5005,
    "event_id": 1001,
    "worker_id": "worker-02:8742",
    "action": "COMPLETED",
    "status_code": 200,
    "execution_time_ms": 1250,
    "created_at": "2025-12-24T10:36:25Z"
  }
]
```

---

## Implementation Notes

### Atomicity & Concurrency

- **Subscribe operation:** Use database transaction with FOR UPDATE lock to atomically fetch and transition PENDING → PROCESSING
- **Retry scheduling:** Database timestamp functions (CURRENT_TIMESTAMP) ensure consistency across workers in different timezones
- **Event log:** Append-only; no updates required after insert

### Worker Registration

Workers do not need explicit registration. Worker identity emerges naturally from event_log entries. Stale workers are detected by analyzing:
- Events stuck in PROCESSING for > X seconds
- PICKED actions without matching STARTED/COMPLETED/FAILED within Y seconds

### Scaling

- Index on `status` + `next_retry_at` for efficient retry polling
- Index on `tags` for subscription queries (may require full-table scan if many events; consider tag bitmap in future)
- event_log grows unbounded; consider archival strategy for production
- Sharding by `event_id % N` for horizontal scaling across multiple databases

---

## Example: Process A Publishes, Process B Consumes

**Process A (Job Publisher):**
```
POST /events
{
  "title": "Batch Report Generation",
  "tags": "reporting,batch",
  "payload": {
    "report_type": "sales",
    "month": "2025-12"
  }
}
→ Returns event_id=2001
```

**Process B (Worker 1):**
```
GET /events/subscribe?tags=reporting&worker_id=worker-03:9100
→ Returns event_id=2001, status=PROCESSING

[Process executes for 15 seconds]

POST /events/2001/complete
{
  "worker_id": "worker-03:9100",
  "execution_time_ms": 15000,
  "status_code": 200
}
→ Event status → COMPLETED
```

**Process B (Worker 2 - Different subscriptions):**
```
GET /events/subscribe?tags=payment&worker_id=worker-04:9101
→ Returns 204 No Content (no payment events available)

[Waits 5 seconds and retries]

GET /events/subscribe?tags=payment&worker_id=worker-04:9101
→ Polls until events match
```

---

## Error Handling

### Idempotency
- POST /events/{id}/complete and /events/{id}/fail should be idempotent (safe to retry if network fails)
- Use event_id + worker_id + action as idempotency key
- Return 200 OK if action already recorded; don't duplicate event_log entry

### Orphaned Events
- Events in PROCESSING state for > 30 seconds with no event_log updates suggest worker crash
- Administrative endpoint to reset: `POST /events/{id}/reset?reason=worker_timeout`

### Payload Size
- Limit event.payload to 1 MB (adjust based on storage constraints)
- For large payloads, store in S3/blob store and reference via URI in payload
