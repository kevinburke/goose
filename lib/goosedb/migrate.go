package goosedb

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/kevinburke/goose/lib/goose"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/godrv"
)

var ErrTableDoesNotExist = errors.New("goosedb: table does not exist")

type migrationSorter []*goose.Migration

// helpers so we can use pkg sort
func (ms migrationSorter) Len() int           { return len(ms) }
func (ms migrationSorter) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms migrationSorter) Less(i, j int) bool { return ms[i].Version < ms[j].Version }

func (ms migrationSorter) Sort(direction bool) {
	// sort ascending or descending by version
	if direction {
		sort.Sort(ms)
	} else {
		sort.Sort(sort.Reverse(ms))
	}

	// now that we're sorted in the appropriate direction,
	// populate next and previous for each migration
	for i, m := range ms {
		prev := int64(-1)
		if i > 0 {
			prev = ms[i-1].Version
			ms[i-1].Next = m.Version
		}
		ms[i].Previous = prev
	}
}

// Runs migration on a specific database instance.
func RunMigrationsOnDb(conf *DBConf, migrationsDir string, target int64, db *sql.DB) (err error) {
	current, err := EnsureDBVersion(conf, db)
	if err != nil {
		return err
	}

	migrations, err := goose.CollectMigrations(migrationsDir, current, target)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		fmt.Printf("goose: no migrations to run. current version: %d\n", current)
		return nil
	}

	ms := migrationSorter(migrations)
	direction := current < target
	ms.Sort(direction)

	fmt.Printf("goose: migrating db environment '%v', current version: %d, target: %d\n",
		conf.Env, current, target)

	for _, m := range ms {

		switch filepath.Ext(m.Source) {
		case ".sql":
			err = runSQLMigration(conf, db, m.Source, m.Version, direction)
		}

		if err != nil {
			return fmt.Errorf("FAIL %w, quitting migration", err)
		}

		fmt.Println("OK   ", filepath.Base(m.Source))
	}

	return nil
}

func RunMigrations(conf *DBConf, migrationsDir string, target int64) error {
	db, err := OpenDBFromDBConf(conf)
	if err != nil {
		return err
	}
	defer db.Close()

	return RunMigrationsOnDb(conf, migrationsDir, target, db)
}

// wrapper for EnsureDBVersion for callers that don't already have
// their own DB instance
func GetDBVersion(conf *DBConf) (int64, error) {

	db, err := OpenDBFromDBConf(conf)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	version, err := EnsureDBVersion(conf, db)
	if err != nil {
		return -1, err
	}

	return version, nil
}

// Create the goose_db_version table
// and insert the initial 0 value into it
func createVersionTable(conf *DBConf, db *sql.DB) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}

	d := conf.Driver.Dialect

	if _, err := txn.Exec(d.createVersionTableSql()); err != nil {
		txn.Rollback()
		return err
	}

	version := 0
	applied := true
	if _, err := txn.Exec(d.insertVersionSql(), version, applied); err != nil {
		txn.Rollback()
		return err
	}

	return txn.Commit()
}

// EnsureDBVersion retrieves the current version for this DB, creating and
// initializing the DB version table if it doesn't exist.
func EnsureDBVersion(conf *DBConf, db *sql.DB) (int64, error) {

	rows, err := conf.Driver.Dialect.dbVersionQuery(db)
	if err != nil {
		if err == ErrTableDoesNotExist {
			return 0, createVersionTable(conf, db)
		}
		return 0, err
	}
	defer rows.Close()

	// The most recent record for each migration specifies
	// whether it has been applied or rolled back.
	// The first version we find that has been applied is the current version.

	toSkip := make([]int64, 0)

	for rows.Next() {
		var row goose.MigrationRecord
		if err = rows.Scan(&row.VersionId, &row.IsApplied); err != nil {
			return 0, fmt.Errorf("error scanning rows: %w", err)
		}

		// have we already marked this version to be skipped?
		skip := false
		for _, v := range toSkip {
			if v == row.VersionId {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		// if version has been applied we're done
		if row.IsApplied {
			return row.VersionId, nil
		}

		// latest version of migration has not been applied.
		toSkip = append(toSkip, row.VersionId)
	}

	panic("failure in EnsureDBVersion()")
}
