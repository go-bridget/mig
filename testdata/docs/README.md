# asset

Stores asset information for each commit

| Name       | Type         | Key | Comment                   |
|------------|--------------|-----|---------------------------|
| id         | int          | PRI | Asset ID                  |
| commit_id  | int          | MUL | Commit ID                 |
| filename   | varchar(255) |     | Filename                  |
| contents   | longtext     |     | File contents             |
| created_at | timestamp    |     | Record creation timestamp |
| updated_at | timestamp    |     | Record update timestamp   |

# branch

Stores information about branches in repositories

| Name          | Type         | Key | Comment                   |
|---------------|--------------|-----|---------------------------|
| id            | int          | PRI | Branch ID                 |
| repository_id | int          | MUL | Repository ID             |
| name          | varchar(255) |     | Branch name               |
| created_at    | timestamp    |     | Record creation timestamp |
| updated_at    | timestamp    |     | Record update timestamp   |

# commit

Stores information about commits in branches

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

# migrations

Migration log of applied migrations

| Name            | Type         | Key | Comment                        |
|-----------------|--------------|-----|--------------------------------|
| project         | varchar(16)  | PRI | Microservice or project name   |
| filename        | varchar(255) | PRI | yyyy-mm-dd-HHMMSS.sql          |
| statement_index | int          |     | Statement number from SQL file |
| status          | text         |     | ok or full error message       |

# repository

Stores basic information about repositories

| Name       | Type         | Key | Comment                   |
|------------|--------------|-----|---------------------------|
| id         | int          | PRI | Repository ID             |
| name       | varchar(255) |     | Repository name           |
| url        | varchar(255) |     | Repository URL            |
| created_at | timestamp    |     | Record creation timestamp |
| updated_at | timestamp    |     | Record update timestamp   |
