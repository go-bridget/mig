# mig - SQL statement migration and development tooling

Mig is a database SQL statement based migration utility. It's short for migrate.

It's used in production on several projects, both personal and
professional. The tool provides controlled migrations for production
environments. It's tested against sqlite, postgres and mysql.

## Status

The project is in active use and is being maintained. Breaking changes
are possible, however the tool lives for many years now, had added
features over time, and needs a cleanup or two.

It's expected minor breaking changes can occur in imported APIs, namely
the `/migrate` import got some usability/ergonomic updates.

Status: active use, maintenance.

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
   filter     Filter schema YAML to driver-independent fields
   lint       Check schema for best practices and comments
   gen        Generate source code from DB schema
   version    Print version
~~~

## Library

The `migrate` package applies the same migrations from Go. A `Manager` takes a
database handle, an `fs.FS` holding the `*.up.sql` files, and the project name
recorded in the migrations table:

~~~go
//go:embed schema/*.sql
var schema embed.FS

fsys, err := fs.Sub(schema, "schema")
if err != nil {
	return err
}

m, err := migrate.NewManager(db, fsys, "events")
if err != nil {
	return err
}

applied, err := m.Apply(ctx)
for _, item := range applied {
	fmt.Println(item.Filename, item.StatementIndex, item.Status)
}
~~~

`Apply` and `List` both return a `[]Migration`, one per migration file, carrying
the row of the migrations table: the index of the last statement that ran and
either `ok` or the error of the one that failed. A migration that has never run
has a `StatementIndex` of -1 and an empty `Status`. `List` applies nothing.

Printing the migrations of a project needs no database:

~~~go
err := migrate.Print(os.Stdout, os.DirFS("schema"))
~~~

The package logs nothing of its own. What a run did comes back from `Apply` and
`List` as a `[]Migration`, for the caller to print or log as they see fit.

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
columns. If a column doesn't have a comment, `mig docs` will convert the
column name into a readable title.

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

### Suggested practices

- Use soft deletes to prevent destructive `DELETE` operations. Various
  data management practices demand you retain this data for business
  reasons.
- Limit client permissions to avoid `DROP, `TRUNCATE`.
- Avoid foreign keys to prevent destructive `DELETE ... CASCADE;` and
  simplify backup/restore of single tables which don't depend on others.
- Exclusively use `?` placeholders or `:name` named placeholders to
  prevent awkward SQL query escape attack patterns.
- [github.com/titpetric/pdo](https://github.com/titpetric/pdo) for fluent
  CRUD with Go >= 1.27, generic function selects that simplify storage
  package implementations.
- Use `mig gen` to follow SQL schema as source of truth. Even if you don't
  use the migration features, the tool allows you to generate the data models,
  which you can simply populate with pdo.Client in a type driven way.
