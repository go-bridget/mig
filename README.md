# mig

Mig is a database SQL statement based migration utility. It's short for migrate.

It's used in production on several projects, both personal and
professional. The tool provides controlled migrations for production
environments.

- Well tested and used: mysql.
- Early, but functional: sqlite.

Status: active use, maintained on occasion.

## Goals

- One way automatic or on-demand SQL migrations,
- Documentation and Code generation from DB schema

The intent of the tool is to provide a simple configuration file based
setup for database schema and access, so it may be deployed in CI jobs
and automated for production environments.

Additionally, it provides schema migrations for the configured databases,
so the migrations themselves can be tested from CI jobs, and can generate
source code and documentation for the final schema.

## Usage

~~~text
Usage: mig (command) [--flags]
Available commands:

   create     Create database schema SQL
   migrate    Apply SQL migrations to database
   docs       Generate markdown docs from DB schema
   lint       Check schema for best practices and comments
   gen        Generate source code from DB schema
   version    Print version
~~~

## Lint

You can use mig to "lint" your database schema, by default:

- a table must have a comment defined,
- a column must have a comment defined
- neither tables nor columns may be prefixed or suffixed with `_`
- table and column names must not use SQL reserved words

### Column/table names

While casing isn't enforced, the encouraged way to name tables and column
names is in lowercase, with `_` as a delimiter. In the case of generating
Go code, "table_name" will be generated as `TableName`.

### Comments

In order to generate documentation and have the database schema readable
without that documentation at hand, comments are enforced on tables and
columns.

### Table names

This rule enforces a thought process where you think about a single
record from a table. For example, if you wanted to use a table called
`dogs`, a single record of that table is a `dog`. As such, a typed object
would be named `Dog`, while a set of dogs would be `[]Dog` (possibly
aliased to `Dogs` in code).

Edge cases: a singular noun may end in a `s`, for example, `bus`. While
it's particularly up to you, a few suggestions for naming the table
apply:

- `bus_entry`
- `stats_entry`
- `statistics_entry`

You may choose other appropriate suffixes, e.g. `_item`, `_record`,...

### Reserved words

SQL servers reserve quite a few keywords for use in SQL statements, and
it's bad practice to use them as table or column names. While we can
generally quote table and column names in statements, it's often
preferable to write simpler sql - if you're not using reserved words,
then you don't need to. The linter will warn you if you're using any of
them as column names or table names.

In the most often case, when you have a `type` column in tables, it's
suggested that you rename the column to `kind`, `kind_of` or similar.

### Prefix/Suffix relationship tables

This isn't enforced by the linter, but it's suggested to prefix or suffix
any relationship tables with `rel_` or `_rel`:

- `rel_company_bus_entry` (preferred)
- `company_bus_entry_rel`

Same plurality and reserved word rules apply for relationship tables.
