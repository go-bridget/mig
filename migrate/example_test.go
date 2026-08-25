package migrate_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing/fstest"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/go-bridget/mig/migrate"
)

// exampleDB returns an in-memory sqlite handle for the examples to migrate.
func exampleDB() *sqlx.DB {
	handle, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	return sqlx.NewDb(handle, "sqlite")
}

// ExampleManager_Apply applies every migration of a project, and prints the run
// it returns: the file, the index of the last statement that ran, and the status
// recorded for it.
func ExampleManager_Apply() {
	fsys := fstest.MapFS{
		"0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\nCREATE TABLE post (id INTEGER);\n",
		)},
		"0002-comment.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE comment (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}

	applied, err := manager.Apply(context.Background())
	if err != nil {
		panic(err)
	}

	for _, item := range applied {
		fmt.Println(item.Filename, item.StatementIndex, item.Status)
	}

	// Output:
	// 0001-user.up.sql 1 ok
	// 0002-comment.up.sql 0 ok
}

// ExampleManager_Apply_failure shows what a failed run reports. The second
// statement of the file fails, so Apply returns its error, and the entry of the
// file holds the index of the statement before it - where the next Apply
// resumes.
func ExampleManager_Apply_failure() {
	fsys := fstest.MapFS{
		"0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\nCREATE TABLE user (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}

	applied, err := manager.Apply(context.Background())
	if err == nil {
		panic("the second statement creates the same table twice")
	}

	failed := applied[0]
	fmt.Println("stopped in", failed.Filename, "after statement", failed.StatementIndex)
	fmt.Println("the entry carries the error:", errors.Is(failed.Err(), err))

	// Output:
	// stopped in 0001-user.up.sql after statement 0
	// the entry carries the error: true
}

// ExampleManager_Apply_noMigrations shows the one failure that otherwise looks
// like a clean run: a filesystem the pattern matches nothing in.
func ExampleManager_Apply_noMigrations() {
	manager, err := migrate.NewManager(exampleDB(), fstest.MapFS{}, "example")
	if err != nil {
		panic(err)
	}

	_, err = manager.Apply(context.Background())

	fmt.Println(errors.Is(err, migrate.ErrNoMigrations))

	// Output:
	// true
}

// ExampleManager_Load selects the migrations of one directory. The pattern
// names a directory, so it is matched against the whole path, and the migration
// sitting elsewhere is left out of the run.
func ExampleManager_Load() {
	fsys := fstest.MapFS{
		"schema/0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\n",
		)},
		"vendor/0001-other.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE other (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}
	manager.Load("schema/*.up.sql")

	applied, err := manager.Apply(context.Background())
	if err != nil {
		panic(err)
	}

	for _, item := range applied {
		fmt.Println(item.Filename)
	}

	// Output:
	// schema/0001-user.up.sql
}

// ExampleManager_Load_anyDepth selects migrations wherever they sit. The pattern
// names no directory, so it is matched against base names, which is what the
// default Pattern does for an embed.FS handed over whole.
func ExampleManager_Load_anyDepth() {
	fsys := fstest.MapFS{
		"schema/0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\n",
		)},
		"vendor/0001-other.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE other (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}
	manager.Load("*.up.sql")

	applied, err := manager.Apply(context.Background())
	if err != nil {
		panic(err)
	}

	for _, item := range applied {
		fmt.Println(item.Filename)
	}

	// Output:
	// schema/0001-user.up.sql
	// vendor/0001-other.up.sql
}
