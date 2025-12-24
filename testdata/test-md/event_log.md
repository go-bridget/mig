# Event Log

Event Log.

| Name              | Type      | Key | Comment           |
|-------------------|-----------|-----|-------------------|
| id                | integer   | PRI | ID                |
| event_id          | integer   | MUL | Event ID          |
| worker_id         | text      | MUL | Worker ID         |
| action            | text      |     | Action            |
| status_code       | integer   |     | Status Code       |
| error_message     | text      |     | Error Message     |
| execution_time_ms | integer   |     | Execution Time Ms |
| created_at        | timestamp | MUL | Created At        |
