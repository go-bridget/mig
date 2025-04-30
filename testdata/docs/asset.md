# Asset

Stores asset information for each commit

| Name       | Type         | Key | Comment                   |
|------------|--------------|-----|---------------------------|
| id         | int          | PRI | Asset ID                  |
| commit_id  | int          | MUL | Commit ID                 |
| filename   | varchar(255) |     | Filename                  |
| contents   | longtext     |     | File contents             |
| created_at | timestamp    |     | Record creation timestamp |
| updated_at | timestamp    |     | Record update timestamp   |
