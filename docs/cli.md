# mig CLI

The `mig` binary applies SQL migrations, and generates documentation and
source code from a database schema. It supports sqlite, mysql and
postgres.

## Install

~~~sh
go install github.com/go-bridget/mig@latest
~~~

This installs `mig` into `$(go env GOBIN)`, or `$(go env GOPATH)/bin` when
`GOBIN` is unset. The main package sits in the root of the module, so there
is no `./cmd/...` path to install.

From a checkout:

~~~sh
task            # go mod tidy, goimports, go install, tests
task build      # build into build/ and build the docker image
~~~

`task build` writes `build/mig` and builds the `go-bridget/mig` image from
`docker/Dockerfile`, which is the binary with `mig` as the entrypoint:

~~~sh
docker run --rm go-bridget/mig version
~~~

## Connecting to a database

Every command that talks to a database takes `--db-dsn`. The scheme
selects the driver:

| Scheme | Driver | Example |
| --- | --- | --- |
| `sqlite://` or `file://` | modernc.org/sqlite | `sqlite://test.db`, `sqlite://:memory:` |
| `mysql://` | go-sql-driver/mysql | `mysql://root:test@tcp(localhost:3306)/events` |
| `postgres://` or `postgresql://` | jackc/pgx | `postgres://postgres:mig@localhost:5432/events?sslmode=disable` |

A DSN without a scheme is read as mysql when it contains `@tcp(` or
`@unix(`, and as sqlite otherwise. For mysql, `collation=utf8mb4_general_ci`,
`parseTime=true` and `loc=Local` are appended when they are not already in
the DSN.

Connecting is retried every 2 seconds, for 100 tries or 2 minutes, whichever
comes first, so a command can be started against a database server that is
still coming up. An empty DSN is not retried and fails with `empty dsn`.

Flags also read their value from the environment: an environment variable is
lowercased and its underscores become hyphens, so `DB_DSN` fills `--db-dsn`
and `SKIP_COMMENTS` fills `--skip-comments`. A variable without an
underscore, or one that does not name a flag of the command being run, is
ignored.

~~~sh
export DB_DSN='mysql://root:test@tcp(localhost:3306)/events'
mig lint
~~~

`MIG_DB_DSN` fills `--db-dsn` as well, and is the one the importable `db`
package reads. The order is `--db-dsn`, then `DB_DSN`, then `MIG_DB_DSN`.

## Commands

~~~text
Usage: mig (command) [--flags]
Available commands:

   create     Create database schema SQL
   migrate    Apply SQL migrations to database
   docs       Generate markdown docs from DB schema
   filter     Filter schema YAML to driver-independent fields
   lint       Lint database schema
   gen        Generate source code from DB schema
   version    Print version
~~~

### create

Prints, or runs, the `CREATE DATABASE` statement for a project. The project
name is the first argument, or `--project`, and is required. Connect to a
database the server already has, such as `mysql` or `postgres`, and name the
one to create:

~~~sh
mig create events --db-dsn "mysql://root:test@tcp(localhost:3306)/mysql" --apply
~~~

| Flag | Default | Description |
| --- | --- | --- |
| `--db-dsn` | none | DSN for the database connection |
| `--project` | none | Project name, when not given as the first argument |
| `--apply` | `false` | `false` prints the query, `true` runs it |

The statement is printed either way. The name is quoted with `"` for
postgres and with a backtick for every other driver. With `--apply`, an
error from the server is printed as `notice: ...` and the command still
exits zero, so creating a database that exists is not a failure.

### migrate

Applies the `*.up.sql` files of a project. The project name is the first
argument, or `--project`, and is the key recorded in the migrations table.
It is required, and has to fit the 16 characters of the `project` column:

~~~sh
mig migrate events --path schema --db-dsn "sqlite://test.db" --apply
~~~

| Flag | Default | Description |
| --- | --- | --- |
| `--db-dsn` | none | DSN for the database connection |
| `--path` | `schema` | Directory holding the migrations |
| `--filename`, `-f` | none | Single file as the migration source |
| `--glob` | `*.up.sql` | Glob selecting migrations under the path |
| `--project` | none | Project name, when not given as the first argument |
| `--apply` | `false` | `false` prints the migrations, `true` runs them |
| `--list` | `false` | Print what the migrations table records and exit |
| `--verbose` | `false` | Print the migrations before applying them |

#### Selecting the migrations

Files are found by walking `--path` and are applied in lexical order of
their path. The glob syntax is `path.Match`'s, and naming a directory is
what decides the depth: a pattern with a `/` in it is matched against the
whole path, one without it against base names, so the default finds
migrations at any depth. Files shorter than two bytes are skipped, two
bytes being the shortest statement there is, the `--` of a comment.

`--filename` wins over `--path` and brings its own glob: the named file is
applied whatever it is called, and the rest of its directory stays out of
the run.

Matching no file at all is an error, `migrate: no migrations found`, so a
`--path` pointing at the wrong directory fails instead of reporting a clean
run. A file that parses to no statements is `migrate: migration has no
statements`.

#### Printing

Without `--apply` and without `--list`, the migrations are printed and no
database is opened. The output is the SQL of the run, each file headed by a
comment naming it and each statement by its index, which is a script that
applies the project by hand:

~~~text
-- Migrations file: 2025-12-24-000001-event-queue-schema.up.sql

-- Statement index: 0
CREATE TABLE event (...);
~~~

A file is split into statements on a `;` that ends a line, so a semicolon in
the middle of a line does not terminate one. Comments are stripped to the
end of their line before the split, and a `uuid()` call is replaced by a
fresh UUID literal, which lets a migration seed a row on a database with no
UUID function and means two prints of that migration do not produce the same
bytes. The index printed for each statement is the one recorded in the
migrations table.

`--verbose` prints all of this before applying.

#### Applying

Each file is applied in one transaction holding a lock keyed by project and
filename, which is what makes it safe to run the same migrations from
several processes at once. The lock is `pg_advisory_xact_lock` on postgres
and `GET_LOCK` with a 30 second timeout on mysql; on sqlite the transaction
is exclusive anyway. The migrations table is created if it is missing, by
`--apply` and by `--list` both.

The row written for a file records the index of the last statement that ran
whether it succeeded or not, so a later `--apply` resumes a failed file from
the statement after that one, and a file that grew statements since the last
run has only the new ones applied. Apply stops at the first file that fails.

`--apply` and `--list` both print one line per migration file: the filename,
the index of the last statement that ran, and either `ok` or the error of
the statement that failed. A file that has never run prints `-1` and `-`.
Rows of the migrations table whose file is gone are printed too, in the same
lexical order.

~~~sh
mig migrate events --path schema --db-dsn "sqlite://test.db" --list
~~~

~~~text
2025-12-24-000001-event-queue-schema.up.sql  7   ok
2026-01-14-000002-event-index.up.sql         0   ok
2026-02-01-000003-event-priority.up.sql      -1  -
~~~

### docs

Introspects the database and writes the schema out. The default is one
markdown file per table under `--output`, named after the lowercased table
title with spaces as underscores, so `event_log` is written to
`event_log.md`. Each path is printed as it is written:

~~~sh
mig docs --output=docs/schema --db-dsn "sqlite://test.db"
~~~

| Flag | Default | Description |
| --- | --- | --- |
| `--db-dsn` | none | DSN for the database connection |
| `--output` | `docs` | Output folder, created if missing |
| `--output-file` | none | Write one file under `--output` instead of one per table |
| `--yaml` | `false` | Output YAML |
| `--json` | `false` | Output JSON |

`--yaml` and `--json` write the whole schema as one document. Without
`--output-file` it goes to stdout; with it, to that name under `--output`.
Given both, `--yaml` wins.

~~~sh
mig docs --output=testdata/docs-sqlite --output-file=schema.yaml --yaml --db-dsn "sqlite://test.db"
~~~

A markdown table has a Name, Type, Key and Comment column. Type is the
normalized data type, the mysql name for it: a sqlite `INTEGER` is reported
as `bigint` and a `TEXT` as `varchar`, which is what makes the schemas of
two drivers comparable. Key is `PRI` for a primary key column, `MUL` for one
in another index or with a name ending in `_id`, and `UNI` for a mysql
unique column. A column with no comment is documented with its name
converted to a readable title, which is what sqlite and postgres columns
always get: only mysql carries column comments in its schema.

A table whose comment is `ignore` is left out of all three formats. That is
a mysql table comment or a postgres `COMMENT ON TABLE`; sqlite has no table
comments, so a sqlite table cannot be excluded this way.

### filter

Reads a schema YAML file, or stdin when no file is given, and prints it back
with only the fields that mean the same thing on every driver. Dropped: the
driver-specific column type, the key marker and the index names, which are
generated per driver. What is kept is the table name and comment, for each
column the name, comment, data type, size and enum values, and for each
index its columns and whether it is primary or unique:

~~~sh
mig filter testdata/docs-sqlite/schema.yaml > sqlite.yaml
mig filter testdata/docs-mysql/schema.yaml > mysql.yaml
dyff between -i sqlite.yaml mysql.yaml
~~~

### lint

Introspects the database and checks the schema. Every violation is printed
and the command exits non-zero, which is what makes it usable as a CI step.

~~~sh
mig lint --db-dsn "sqlite://test.db"
~~~

| Flag | Default | Description |
| --- | --- | --- |
| `--db-dsn` | none | DSN for the database connection |
| `--skip-comments` | `false` | Skip validating table and column comments |
| `--skip-plural` | `false` | Skip validating table names for singular form |

The rules, per table:

- The table has a comment, and so does every column. A column named `id` is
  exempt. Only the column rule can fail, and only against mysql: sqlite and
  postgres columns are given the column name as a title during
  introspection, and a table with no comment is given its own name on every
  driver, which is already enough to pass.
- The table name does not end in `s`, which is the whole of the singular
  check. `--skip-plural` turns it off, and the `migrations` table is exempt.
- Neither the table name nor a column name is prefixed or suffixed with `_`.
- Neither the table name nor a column name is one of the MySQL reserved
  words, compared case-insensitively.

A table whose comment is `ignore` is skipped entirely.

The reasoning behind the rules, and the naming conventions that go with
them, are in the [README](../README.md#lint).

### gen

Generates source code from the schema of the connected database. Go is the
only language implemented:

~~~sh
mig gen --lang=go --db-dsn "sqlite://test.db" --output model
~~~

| Flag | Default | Description |
| --- | --- | --- |
| `--db-dsn` | none | DSN for the database connection |
| `--lang` | `go` | Programming language |
| `--output` | `model` | Output folder for the generated types |
| `--go.fill-json` | `false` | Fill JSON tags |
| `--go.skip-json` | `false` | Skip JSON tags |

The output is one gofmt-ed file, `<output>/types.mig.go`, whose path is
printed, in a package named after the last element of `--output`. It carries
a
`Code generated by go-bridget/mig. DO NOT EDIT.` header, a `QueryOption`
interface with a `QueryConfig` implementing it, and one struct per table.
A table whose comment is `ignore` is skipped.

Names are converted to camel case, so table `event_log` becomes `EventLog`
and column `retry_count` becomes `RetryCount`. Initialisms are uppercased,
so `id` becomes `ID`. Each field carries the column comment as its doc
comment and a `db` tag with the column name. The JSON tag is `json:"-"` by
default, `--go.fill-json` puts the column name in it, and `--go.skip-json`
leaves the tag off:

~~~go
type Event struct {
	// ID
	ID int64 `db:"id" json:"-"`

	// Retry Count
	RetryCount int64 `db:"retry_count" json:"-"`
}
~~~

### version

Prints the binary name, the build version and the build time. The version
and the time are set with `-ldflags` at build time, as `task build` does,
and are empty in a `go install` build:

~~~sh
mig version
~~~

## A migration run

The commands compose into the run a CI job makes. Create the database, apply
the migrations, check the result, then generate the documentation and the
models:

~~~sh
export DB_DSN='mysql://root:test@tcp(localhost:13306)/events'

mig create events --db-dsn "mysql://root:test@tcp(localhost:13306)/mysql" --apply
mig migrate events --path testdata/events/mysql --apply
mig lint
mig docs --output=testdata/docs-mysql --output-file=schema.yaml --yaml
mig gen --lang=go --output=testdata/gen_go
~~~

`Taskfile.yml` runs this against sqlite, mysql and postgres under
`task test`, with the servers coming from `compose.yml`.
