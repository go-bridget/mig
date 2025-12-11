# Commit

Stores information about commits in branches.

| Name         | Type         | Key | Comment                   |
|--------------|--------------|-----|---------------------------|
| id           | int          | PRI | Commit ID                 |
| branch_id    | int          | MUL | Branch ID                 |
| commit_hash  | varchar(40)  |     | Commit hash               |
| author       | varchar(255) |     | Commit author             |
| message      | text         |     | Commit message            |
| committed_at | timestamp    |     | Commit timestamp          |
| created_at   | timestamp    |     | Record creation timestamp |
| updated_at   | timestamp    |     | Record update timestamp   |
