# incoming

Incoming stats log, writes only

| Name             | Type            | Key | Comment                             |
|------------------|-----------------|-----|-------------------------------------|
| id               | bigint unsigned | PRI | Tracking ID                         |
| property         | varchar(32)     |     | Property name (human readable, a-z) |
| property_section | int unsigned    |     | Property Section ID                 |
| property_id      | int unsigned    |     | Property Item ID                    |
| remote_ip        | varchar(255)    |     | Remote IP from user making request  |
| stamp            | datetime        |     | Timestamp of request                |
