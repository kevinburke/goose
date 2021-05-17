package goosedb

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Run a migration specified in raw SQL.
//
// Sections of the script can be annotated with a special comment,
// starting with "-- +goose" to specify whether the section should
// be applied during an Up or Down migration
//
// All statements following an Up or Down directive are grouped together
// until another direction directive is found.
func runSQLMigration(conf *DBConf, db *sql.DB, scriptFile string, v int64, direction bool) error {
	f, err := os.Open(scriptFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// find each statement, checking annotations for up/down direction
	// and execute each of them in the current transaction.
	// Commits the transaction if successfully applied each statement and
	// records the version into the version table or returns an error and
	// rolls back the transaction.
	stmts, err := splitSQLStatements(f, direction)
	if err != nil {
		return err
	}

	// Choose query strategy
	singleQueryOutsideTxn := false
	for _, query := range stmts {
		if cannotRunInTransaction(query) {
			if len(stmts) > 1 {
				log.Fatalf("Query %s cannot run in a transaction, but was paired with other queries; run it in isolation", query)
			} else {
				singleQueryOutsideTxn = true
			}
		}
	}

	if singleQueryOutsideTxn {
		if _, err = db.Exec(stmts[0]); err != nil {
			log.Fatalf("FAIL %s (%v), quitting migration.", filepath.Base(scriptFile), err)
			return err
		}
		stmt := conf.Driver.Dialect.insertVersionSql()
		if _, err := db.Exec(stmt, v, direction); err != nil {
			log.Printf("WARNING: Executed single query %s but could not commit the version bump %s; error was %s\n", stmts[0], stmt, err.Error())
		}
		return err
	}

	txn, err := db.Begin()
	if err != nil {
		log.Fatal("db.Begin:", err)
	}

	for _, query := range stmts {
		if _, err = txn.Exec(query); err != nil {
			txn.Rollback()
			log.Fatalf("FAIL %s (%v), quitting migration.", filepath.Base(scriptFile), err)
			return err
		}
	}

	// Update the version table for the given migration,
	// and finalize the transaction.
	// XXX: drop goose_db_version table on some minimum version number?
	stmt := conf.Driver.Dialect.insertVersionSql()
	if _, err := txn.Exec(stmt, v, direction); err != nil {
		txn.Rollback()
		return err
	}
	return txn.Commit()
}

const sqlCmdPrefix = "-- +goose "

// Checks the line to see if the line has a statement-ending semicolon
// or if the line contains a double-dash comment.
func endsWithSemicolon(line string) bool {

	prev := ""
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		prev = word
	}

	return strings.HasSuffix(prev, ";")
}

var concurrentIdxRx = regexp.MustCompile(`^CREATE (UNIQUE )?INDEX CONCURRENTLY\s`)

func cannotRunInTransaction(query string) bool {
	upQuery := strings.TrimSpace(strings.ToUpper(query))
	lines := strings.Split(upQuery, "\n")
	for i, line := range lines {
		trimLine := strings.TrimSpace(line)
		if trimLine == "" || strings.HasPrefix(trimLine, "--") {
			if i+1 > len(lines) {
				// every line is a comment
				return false
			}
			upQuery = strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
		} else {
			break
		}
	}
	// There are probably more potential cases here.
	return concurrentIdxRx.MatchString(upQuery) ||
		(strings.HasPrefix(upQuery, "ALTER TYPE ") && strings.Contains(upQuery, " ADD "))
}

// Split the given sql script into individual statements.
//
// The base case is to simply split on semicolons, as these
// naturally terminate a statement.
//
// However, more complex cases like pl/pgsql can have semicolons
// within a statement. For these cases, we provide the explicit annotations
// 'StatementBegin' and 'StatementEnd' to allow the script to
// tell us to ignore semicolons.
func splitSQLStatements(r io.Reader, direction bool) ([]string, error) {
	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)

	// track the count of each section
	// so we can diagnose scripts with no annotations
	upSections := 0
	downSections := 0

	statementEnded := false
	ignoreSemicolons := false
	directionIsActive := false

	stmts := make([]string, 0)
	for scanner.Scan() {

		line := scanner.Text()

		// handle any goose-specific commands
		if strings.HasPrefix(line, sqlCmdPrefix) {
			cmd := strings.TrimSpace(line[len(sqlCmdPrefix):])
			switch cmd {
			case "Up":
				//lint:ignore S1002 would rather write it this way.
				directionIsActive = (direction == true)
				upSections++

			case "Down":
				//lint:ignore S1002 would rather write it this way.
				directionIsActive = (direction == false)
				downSections++

			case "StatementBegin":
				if directionIsActive {
					ignoreSemicolons = true
				}

			case "StatementEnd":
				if directionIsActive {
					//lint:ignore S1002 would rather write it this way.
					statementEnded = (ignoreSemicolons == true)
					ignoreSemicolons = false
				}
			}
		}

		if !directionIsActive {
			continue
		}

		if _, err := buf.WriteString(line + "\n"); err != nil {
			return nil, fmt.Errorf("io err: %v", err)
		}

		// Wrap up the two supported cases: 1) basic with semicolon; 2) psql statement
		// Lines that end with semicolon that are in a statement block
		// do not conclude statement.
		if (!ignoreSemicolons && endsWithSemicolon(line)) || statementEnded {
			statementEnded = false
			stmts = append(stmts, buf.String())
			buf.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning migration: %v", err)
	}

	// diagnose likely migration script errors
	if ignoreSemicolons {
		return nil, fmt.Errorf("saw '-- +goose StatementBegin' with no matching '-- +goose StatementEnd'")
	}

	if bufferRemaining := strings.TrimSpace(buf.String()); len(bufferRemaining) > 0 {
		return nil, fmt.Errorf("unexpected unfinished SQL query: %s. Missing a semicolon?", bufferRemaining)
	}

	if upSections == 0 && downSections == 0 {
		return nil, fmt.Errorf(`no Up/Down annotations found, so no statements were executed.
See https://github.com/kevinburke/goose for details`)
	}

	return stmts, nil
}
