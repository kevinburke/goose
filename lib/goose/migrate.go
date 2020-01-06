package goose

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	ErrNoPreviousVersion = errors.New("no previous version found")
)

type MigrationRecord struct {
	VersionId int64
	TStamp    time.Time
	IsApplied bool // was this a result of up() or down()
}

type Migration struct {
	Version  int64
	Next     int64  // next version, or -1 if none
	Previous int64  // previous version, -1 if none
	Source   string // path to .go or .sql script
}

func newMigration(v int64, src string) *Migration {
	return &Migration{v, -1, -1, src}
}

// CollectMigrations collects and returns all of the valid looking migration
// scripts in dirpath. Set current to 0 and target to a very large number to
// collect all migrations in the directory.
func CollectMigrations(dirpath string, current, target int64) ([]*Migration, error) {
	// extract the numeric component of each migration,
	// filter out any uninteresting files,
	// and ensure we only have one file per migration version.
	m := make([]*Migration, 0)
	err := filepath.Walk(dirpath, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if v, e := NumericComponent(name); e == nil {

			for _, g := range m {
				if v == g.Version {
					log.Fatalf("more than one file specifies the migration for version %d (%s and %s)",
						v, g.Source, filepath.Join(dirpath, name))
				}
			}

			if versionFilter(v, current, target) {
				m = append(m, newMigration(v, name))
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// versionFilter returns true if v is greater than current and less than or
// equal to target.
func versionFilter(v, current, target int64) bool {
	if target > current {
		return v > current && v <= target
	}

	if target < current {
		return v <= current && v > target
	}

	return false
}

// NumericComponent returns the integer part of name, if it exists, or an error
// if name does not resemble a Goose migration file.
//
// The function looks for migration scripts with names in the form:
//
//      XXX_descriptivename.ext
//
// where XXX specifies the version number and ext specifies the type of the
// migration.
func NumericComponent(name string) (int64, error) {
	base := filepath.Base(name)

	if ext := filepath.Ext(base); ext != ".sql" {
		return 0, errors.New("goose: not a recognized migration file type")
	}

	idx := strings.Index(base, "_")
	if idx < 0 {
		return 0, errors.New("goose: no separator found")
	}

	n, e := strconv.ParseInt(base[:idx], 10, 64)
	if e == nil && n <= 0 {
		return 0, errors.New("goose: migration IDs must be greater than zero")
	}

	return n, e
}

func GetPreviousDBVersion(dirpath string, version int64) (previous int64, err error) {

	previous = -1
	sawGivenVersion := false

	filepath.Walk(dirpath, func(name string, info os.FileInfo, walkerr error) error {

		if !info.IsDir() {
			if v, e := NumericComponent(name); e == nil {
				if v > previous && v < version {
					previous = v
				}
				if v == version {
					sawGivenVersion = true
				}
			}
		}

		return nil
	})

	if previous == -1 {
		if sawGivenVersion {
			// the given version is (likely) valid but we didn't find
			// anything before it.
			// 'previous' must reflect that no migrations have been applied.
			previous = 0
		} else {
			err = ErrNoPreviousVersion
		}
	}

	return
}

// helper to identify the most recent possible version
// within a folder of migration scripts
func GetMostRecentDBVersion(dirpath string) (version int64, err error) {

	version = -1

	filepath.Walk(dirpath, func(name string, info os.FileInfo, walkerr error) error {
		if walkerr != nil {
			return walkerr
		}

		if !info.IsDir() {
			if v, e := NumericComponent(name); e == nil {
				if v > version {
					version = v
				}
			}
		}

		return nil
	})

	if version == -1 {
		err = errors.New("no valid version found")
	}

	return
}

// CreateMigration creates a new migration and writes it to a new file in dir.
// The path to the file will be returned.
func CreateMigration(name, migrationType, dir string, t time.Time) (path string, err error) {

	if migrationType != "sql" {
		return "", errors.New("migration type must be 'sql'")
	}

	timestamp := t.Format("20060102150405")
	filename := fmt.Sprintf("%v_%v.%v", timestamp, name, migrationType)

	fpath := filepath.Join(dir, filename)

	var tmpl *template.Template
	if migrationType == "sql" {
		tmpl = sqlMigrationTemplate
	}

	path, err = writeTemplateToFile(fpath, tmpl, timestamp)

	return
}

var sqlMigrationTemplate = template.Must(template.New("goose.sql-migration").Parse(
	`-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
`))
