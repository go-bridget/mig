# Event

Event.

| Name          | Type      | Key | Comment       |
|---------------|-----------|-----|---------------|
| id            | integer   | PRI | ID            |
| title         | text      |     | Title         |
| description   | text      |     | Description   |
| tags          | text      | MUL | Tags          |
| status        | enum      | MUL | Status        |
| payload       | text      |     | Payload       |
| retry_count   | integer   |     | Retry Count   |
| max_retries   | integer   |     | Max Retries   |
| next_retry_at | timestamp | MUL | Next Retry At |
| created_at    | timestamp |     | Created At    |
| updated_at    | timestamp |     | Updated At    |
