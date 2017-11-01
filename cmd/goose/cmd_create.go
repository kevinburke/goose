package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kevinburke/goose/lib/goose"
)

var createCmd = &Command{
	Name:    "create",
	Usage:   "create <migration-name>",
	Summary: "Create the scaffolding for a new migration",
	Help:    `Create a file with a new migration. The file will have the given name`,
	Run:     createRun,
	Flag:    *flag.NewFlagSet("create", flag.ExitOnError),
}

func createRun(cmd *Command, args ...string) {
	if len(args) < 1 {
		log.Fatal("goose create: migration name required")
	}

	migrationType := "sql"

	conf, err := dbConfFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(conf.MigrationsDir, 0755); err != nil {
		log.Fatal(err)
	}

	n, err := goose.CreateMigration(args[0], migrationType, conf.MigrationsDir, time.Now().UTC())
	if err != nil {
		log.Fatal(err)
	}

	a, e := filepath.Abs(n)
	if e != nil {
		log.Fatal(e)
	}

	fmt.Println("goose: created", a)
}
