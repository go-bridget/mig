# Asset

Stores asset information for each commit

| Name       | Type         | Key | Comment                   |
|------------|--------------|-----|---------------------------|
| id         | int(11)      | PRI | Asset ID                  |
| commit_id  | int(11)      | MUL | Commit ID                 |
| filename   | varchar(255) |     | Filename                  |
| contents   | longtext     |     | File contents             |
| created_at | timestamp    |     | Record creation timestamp |
| updated_at | timestamp    |     | Record update timestamp   |

# Branch

Stores information about branches in repositories

| Name          | Type         | Key | Comment                   |
|---------------|--------------|-----|---------------------------|
| id            | int(11)      | PRI | Branch ID                 |
| repository_id | int(11)      | MUL | Repository ID             |
| name          | varchar(255) |     | Branch name               |
| created_at    | timestamp    |     | Record creation timestamp |
| updated_at    | timestamp    |     | Record update timestamp   |

# Commit

Stores information about commits in branches

| Name         | Type         | Key | Comment                   |
|--------------|--------------|-----|---------------------------|
| id           | int(11)      | PRI | Commit ID                 |
| branch_id    | int(11)      | MUL | Branch ID                 |
| commit_hash  | varchar(40)  |     | Commit hash               |
| author       | varchar(255) |     | Commit author             |
| message      | text         |     | Commit message            |
| committed_at | timestamp    |     | Commit timestamp          |
| created_at   | timestamp    |     | Record creation timestamp |
| updated_at   | timestamp    |     | Record update timestamp   |

# Migrations

Migration log of applied migrations

| Name            | Type         | Key | Comment                        |
|-----------------|--------------|-----|--------------------------------|
| project         | varchar(16)  | PRI | Microservice or project name   |
| filename        | varchar(255) | PRI | yyyy-mm-dd-HHMMSS.sql          |
| statement_index | int(11)      |     | Statement number from SQL file |
| status          | text         |     | ok or full error message       |

# Repository

Stores basic information about repositories

| Name       | Type         | Key | Comment                   |
|------------|--------------|-----|---------------------------|
| id         | int(11)      | PRI | Repository ID             |
| name       | varchar(255) |     | Repository name           |
| url        | varchar(255) |     | Repository URL            |
| created_at | timestamp    |     | Record creation timestamp |
| updated_at | timestamp    |     | Record update timestamp   |
