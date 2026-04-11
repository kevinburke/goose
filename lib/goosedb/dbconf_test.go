package goosedb

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	dbconf, err := NewConfig("postgres", "user=liam dbname=tester sslmode=disable", "sql/migrations")
	if err != nil {
		t.Fatal(err)
	}

	got := []string{dbconf.MigrationsDir, dbconf.Env, dbconf.Driver.Name, dbconf.Driver.OpenStr}
	want := []string{"sql/migrations", "production", "postgres", "user=liam dbname=tester sslmode=disable"}

	for i, s := range got {
		if s != want[i] {
			t.Errorf("Unexpected DBConf value. got %v, want %v", s, want[i])
		}
	}

	if reflect.TypeOf(dbconf.Driver.Dialect) != reflect.TypeFor[*PostgresDialect]() {
		t.Errorf("Unexpected dialect type: got %T", dbconf.Driver.Dialect)
	}
}

func TestNewConfigRejectsUnknownDriver(t *testing.T) {
	_, err := NewConfig("oracle", "ignored", "sql/migrations")
	if err == nil {
		t.Fatal("expected error for invalid driver")
	}
}

func TestNewConfigCustom(t *testing.T) {
	driver := DBDriver{
		Name:    "custom",
		OpenStr: "dsn",
		Import:  "github.com/example/customdriver",
		Dialect: &Sqlite3Dialect{},
	}

	dbconf, err := NewConfigCustom(driver, "sql/migrations")
	if err != nil {
		t.Fatal(err)
	}

	if dbconf.Env != "production" {
		t.Errorf("Unexpected env: got %q want %q", dbconf.Env, "production")
	}
	if !reflect.DeepEqual(dbconf.Driver, driver) {
		t.Errorf("Unexpected driver: got %#v want %#v", dbconf.Driver, driver)
	}
}

func TestNewConfigCustomRejectsInvalidDriver(t *testing.T) {
	_, err := NewConfigCustom(DBDriver{Name: "broken"}, "sql/migrations")
	if err == nil {
		t.Fatal("expected error for invalid custom driver")
	}
}

func TestBasics(t *testing.T) {

	dbconf, err := NewDBConf("../../db-sample", "test", "")
	if err != nil {
		t.Fatal(err)
	}

	got := []string{dbconf.MigrationsDir, dbconf.Env, dbconf.Driver.Name, dbconf.Driver.OpenStr}
	want := []string{"../../db-sample/migrations", "test", "postgres", "user=liam dbname=tester sslmode=disable"}

	for i, s := range got {
		if s != want[i] {
			t.Errorf("Unexpected DBConf value. got %v, want %v", s, want[i])
		}
	}
}

func TestImportOverride(t *testing.T) {

	dbconf, err := NewDBConf("../../db-sample", "customimport", "")
	if err != nil {
		t.Fatal(err)
	}

	got := dbconf.Driver.Import
	want := "github.com/custom/driver"
	if got != want {
		t.Errorf("bad custom import. got %v want %v", got, want)
	}
}

func TestDriverSetFromEnvironmentVariable(t *testing.T) {
	databaseUrlEnvVariableKey := "DB_DRIVER"
	databaseUrlEnvVariableVal := "sqlite3"
	databaseOpenStringKey := "DATABASE_URL"
	databaseOpenStringVal := "db.db"

	t.Setenv(databaseUrlEnvVariableKey, databaseUrlEnvVariableVal)
	t.Setenv(databaseOpenStringKey, databaseOpenStringVal)

	dbconf, err := NewDBConf("../../db-sample", "environment_variable_config", "")
	if err != nil {
		t.Fatal(err)
	}

	got := reflect.TypeOf(dbconf.Driver.Dialect)
	want := reflect.TypeFor[*Sqlite3Dialect]()

	if got != want {
		t.Errorf("Not able to read the driver type from environment variable."+
			"got %v want %v", got, want)
	}

	gotOpenString := dbconf.Driver.OpenStr
	wantOpenString := databaseOpenStringVal

	if gotOpenString != wantOpenString {
		t.Errorf("Not able to read the open string from the environment."+
			"got %v want %v", gotOpenString, wantOpenString)
	}
}
