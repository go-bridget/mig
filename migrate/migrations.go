package migrate

import (
	"embed"
)

//go:embed *.sql
var migrationsFS embed.FS
