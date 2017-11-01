package main

import (
	"flag"
	"log"

	"github.com/kevinburke/goose/lib/goose"
)

var downCmd = &Command{
	Name:    "down",
	Flag:    *flag.NewFlagSet("down", flag.ExitOnError),
	Usage:   "usage: down",
	Summary: "Roll back the version by 1",
	Help:    `Execute the "down" command for the most recently applied migration`,
	Run:     downRun,
}

func downRun(_ *Command, args ...string) {
	conf, err := dbConfFromFlags()
	if err != nil {
		log.Fatal(err)
	}

	current, err := goose.GetDBVersion(conf)
	if err != nil {
		log.Fatal(err)
	}

	previous, err := goose.GetPreviousDBVersion(conf.MigrationsDir, current)
	if err != nil {
		log.Fatal(err)
	}

	if err = goose.RunMigrations(conf, conf.MigrationsDir, previous); err != nil {
		log.Fatal(err)
	}
}
