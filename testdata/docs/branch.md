# Branch

Stores information about branches in repositories.

| Name          | Type         | Key | Comment                   |
|---------------|--------------|-----|---------------------------|
| id            | int(11)      | PRI | Branch ID                 |
| repository_id | int(11)      | MUL | Repository ID             |
| name          | varchar(255) |     | Branch name               |
| created_at    | timestamp    |     | Record creation timestamp |
| updated_at    | timestamp    |     | Record update timestamp   |
