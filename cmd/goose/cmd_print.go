package main

import (
	"flag"
	"fmt"
	"log"
)

var printCmd = &Command{
	Name:    "print",
	Usage:   "",
	Summary: "Print psql-compatible db configuration to the command line",
	Help:    `Intended for converting dbconf.yml to a psql repl.`,
	Run:     printRun,
	Flag:    *flag.NewFlagSet("print", flag.ExitOnError),
}

func printRun(cmd *Command, args ...string) {
	conf, err := dbConfFromFlags()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(conf.Driver.OpenStr)
}
