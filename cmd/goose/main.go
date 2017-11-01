package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kevinburke/goose/lib/goose"
)

// global options. available to any subcommands.
var flagPath = flag.String("path", "db", "folder containing db info")
var flagEnv = flag.String("env", "development", "which DB environment to use")
var flagPgSchema = flag.String("pgschema", "", "which postgres-schema to migrate (default = none)")

// helper to create a DBConf from the given flags
func dbConfFromFlags() (dbconf *goose.DBConf, err error) {
	return goose.NewDBConf(*flagPath, *flagEnv, *flagPgSchema)
}

var commands = []*Command{
	upCmd,
	downCmd,
	redoCmd,
	statusCmd,
	createCmd,
	dbVersionCmd,
	versionCmd,
	helpCmd,
	initCmd,
}

var versionCmd = &Command{
	Name:    "version",
	Usage:   "Print the goose version",
	Summary: "Print the current version of the goose tool",
	Run:     versionRun,
	Help:    "Print the goose version",
	Flag:    *flag.NewFlagSet("version", flag.ExitOnError),
}

var helpCmd = &Command{
	Name:    "help",
	Usage:   "Print the help text",
	Summary: "Print the help text",
	Run:     helpRun,
	Help:    "Print the help text",
	Flag:    *flag.NewFlagSet("help", flag.ExitOnError),
}

var initCmd = &Command{
	Name:    "init",
	Usage:   "Create the scaffolding",
	Summary: "Create the migration scaffolding",
	Run:     initRun,
	Flag:    *flag.NewFlagSet("init", flag.ExitOnError),
}

var dbConfTpl = []byte(`# Database configuration file.
#
# Example configurations (uncomment and modify as you see fit):
#
# development:
#     driver: postgres
#     open: user=mypguser dbname=mydatabase sslmode=disable
#
# cluster:
#     driver: mysql
#     open: $DATABASE_URL
`)

func initRun(*Command, ...string) {
	wd, err := os.Getwd()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(2)
	}
	dbDir := filepath.Join(wd, "db")
	os.Mkdir(dbDir, 0700)
	conf := filepath.Join(dbDir, "dbconf.yml")
	if _, err := os.Stat(conf); os.IsNotExist(err) {
		if err := ioutil.WriteFile(conf, dbConfTpl, 0600); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(2)
		}
	}
}

func helpRun(*Command, ...string) {
	flag.Usage()
}

// The version of the goose tool.
const VERSION = "1.7" // Bump this by running "make release".

func versionRun(*Command, ...string) {
	fmt.Fprintf(os.Stderr, "goose version %s\n", VERSION)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 || args[0] == "-h" {
		flag.Usage()
		return
	}

	var cmd *Command
	name := args[0]
	for _, c := range commands {
		if strings.HasPrefix(c.Name, name) {
			cmd = c
			break
		}
	}

	if cmd == nil {
		fmt.Printf("error: unknown command %q\n", name)
		flag.Usage()
		os.Exit(1)
	}

	cmd.Exec(args[1:])
}

func usage() {
	fmt.Print(usagePrefix)
	flag.PrintDefaults()
	usageTmpl.Execute(os.Stdout, commands)
}

var usagePrefix = `goose is a database migration management system for Go projects.

Usage:
    goose [options] <subcommand> [subcommand options]

Options:
`
var usageTmpl = template.Must(template.New("usage").Parse(
	`
Commands:{{range .}}
    {{.Name | printf "%-10s"}} {{.Summary}}{{end}}
`))
