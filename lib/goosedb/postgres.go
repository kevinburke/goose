package goosedb

import (
	"database/sql"

	"github.com/jackc/pgx/v5/stdlib"
)

// pgx/v5's stdlib driver registers itself with database/sql under the name
// "pgx". Historically goose used github.com/lib/pq, which registers as
// "postgres", and existing dbconf.yml files in the wild rely on
// `driver: postgres`. Register the pgx driver under that legacy name as
// well so those configurations keep working unchanged after the migration
// off lib/pq.
func init() {
	sql.Register("postgres", &stdlib.Driver{})
}
