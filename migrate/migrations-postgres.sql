CREATE TABLE IF NOT EXISTS migrations (
    project varchar(16) NOT NULL,
    filename varchar(255) NOT NULL,
    statement_index int NOT NULL,
    status text NOT NULL,
    PRIMARY KEY (project, filename)
);

COMMENT ON TABLE migrations IS 'Migration log of applied migrations';
COMMENT ON COLUMN migrations.project IS 'Microservice or project name';
COMMENT ON COLUMN migrations.filename IS 'yyyy-mm-dd-HHMMSS.sql';
COMMENT ON COLUMN migrations.statement_index IS 'Statement number from SQL file';
COMMENT ON COLUMN migrations.status IS 'ok or full error message';
