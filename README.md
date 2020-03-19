# mig

Mig is a database SQL statement based migration utility. It's short for migrate.

## Goals

1. Define schemas, users and permissions,
2. One way automatic or on-demand SQL migrations,
3. Documentation and Code generation from DB schema

The intent of the tool is to provide a simple configuration file based setup
for database schema and access, so it may be deployed in CI jobs and automated
for production environments.

Additionally, it provides schema migrations for the configured databases, so
the migrations themselves can be tested from CI jobs, and can generate source
code and documentation for the final schema.
