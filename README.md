# goose

This is a fork of bitbucket.org/liamstask/goose that implements several new
features:

- Support for db queries that cannot be run in a transaction (CREATE INDEX
  CONCURRENTLY, ALTER TYPE)
- Explicit versions/tags
- Dependency vendoring
- Updated versions of dependencies; fixes [a sqlite3 warning][warning] on the
latest version of Macs.

[warning]: https://github.com/mattn/go-sqlite3/issues/336

I removed support for the Go migration mode.

goose is a database migration tool.

You can manage your database's evolution by creating incremental SQL scripts.

# Install

    $ make install

This will install the `goose` binary to your `$GOPATH/bin` directory.

You can also build goose into your own applications by importing
`github.com/kevinburke/goose/lib/goose`. Documentation is available at
[godoc.org](http://godoc.org/github.com/kevinburke/goose/lib/goose).

NOTE: the API is still new, and may undergo some changes.

# Usage

goose provides several commands to help manage your database schema.

## create

Create a SQL migration:

    $ goose create AddSomeColumns
    $ goose: created db/migrations/20130106093224_AddSomeColumns.sql

Edit the newly created script to define the behavior of your migration.

## up

Apply all available migrations.

    $ goose up
    $ goose: migrating db environment 'development', current version: 0, target: 3
    $ OK    001_basics.sql
    $ OK    002_next.sql
    $ OK    003_and_again.sql

### option: pgschema

Use the `pgschema` flag with the `up` command specify a postgres schema.

    $ goose -pgschema=my_schema_name up
    $ goose: migrating db environment 'development', current version: 0, target: 3
    $ OK    001_basics.sql
    $ OK    002_next.sql
    $ OK    003_and_again.sql

## down

Roll back a single migration from the current version.

    $ goose down
    $ goose: migrating db environment 'development', current version: 3, target: 2
    $ OK    003_and_again.sql

## redo

Roll back the most recently applied migration, then run it again.

    $ goose redo
    $ goose: migrating db environment 'development', current version: 3, target: 2
    $ OK    003_and_again.sql
    $ goose: migrating db environment 'development', current version: 2, target: 3
    $ OK    003_and_again.sql

## status

Print the status of all migrations:

    $ goose status
    $ goose: status for environment 'development'
    $   Applied At                  Migration
    $   =======================================
    $   Sun Jan  6 11:25:03 2013 -- 001_basics.sql
    $   Sun Jan  6 11:25:03 2013 -- 002_next.sql
    $   Pending                  -- 003_and_again.sql

## dbversion

Print the current version of the database:

    $ goose dbversion
    $ goose: dbversion 002


`goose -h` provides more detailed info on each command.


# Migrations

goose supports migrations written in SQL - see the `goose create` command above
for details on how to generate them.

## SQL Migrations

A sample SQL migration looks like:

```sql
-- +goose Up
CREATE TABLE post (
    id int NOT NULL,
    title text,
    body text,
    PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE post;
```

Notice the annotations in the comments. Any statements following `-- +goose Up` will be executed as part of a forward migration, and any statements following `-- +goose Down` will be executed as part of a rollback.

By default, SQL statements are delimited by semicolons - in fact, query statements must end with a semicolon to be properly recognized by goose.

More complex statements (PL/pgSQL) that have semicolons within them must be annotated with `-- +goose StatementBegin` and `-- +goose StatementEnd` to be properly recognized. For example:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION histories_partition_creation( DATE, DATE )
returns void AS $$
DECLARE
  create_query text;
BEGIN
  FOR create_query IN SELECT
      'CREATE TABLE IF NOT EXISTS histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || ' ( CHECK( created_at >= timestamp '''
      || TO_CHAR( d, 'YYYY-MM-DD 00:00:00' )
      || ''' AND created_at < timestamp '''
      || TO_CHAR( d + INTERVAL '1 month', 'YYYY-MM-DD 00:00:00' )
      || ''' ) ) inherits ( histories );'
    FROM generate_series( $1, $2, '1 month' ) AS d
  LOOP
    EXECUTE create_query;
  END LOOP;  -- LOOP END
END;         -- FUNCTION END
$$
language plpgsql;
-- +goose StatementEnd
```

# Configuration

goose expects you to maintain a folder (typically called "db"), which contains the following:

* a `dbconf.yml` file that describes the database configurations you'd like to use
* a folder called "migrations" which contains `.sql` and/or `.go` scripts that implement your migrations

You may use the `-path` option to specify an alternate location for the folder containing your config and migrations.

A sample `dbconf.yml` looks like

```yml
development:
    driver: postgres
    open: user=liam dbname=tester sslmode=disable
```

Here, `development` specifies the name of the environment, and the `driver` and `open` elements are passed directly to database/sql to access the specified database.

You may include as many environments as you like, and you can use the `-env` command line option to specify which one to use. goose defaults to using an environment called `development`.

goose will expand environment variables in the `open` element. For an example, see the Heroku section below.

## Database Drivers

Currently, available dialects are: "postgres", "mysql", or "sqlite3".

Because migrations written in SQL are executed directly by the goose binary,
only drivers compiled into goose may be used for these migrations.

## Queries that require a transaction

Some Postgres migrations (CREATE INDEX CONCURRENTLY, ALTER TYPE) cannot be run
in a transaction. `goose` has a special mode that can detect these queries and
run them outside of a transaction. To avoid partially-applied transactions,
we require that these can't-run-in-transaction queries consist of a single
statement per up/down block, e.g. you can't do `ALTER TYPE ...; ALTER TYPE
...;`.

# Contributors

Thank you!

* Josh Bleecher Snyder (josharian)
* Abigail Walthall (ghthor)
* Daniel Heath (danielrheath)
* Chris Baynes (chris_baynes)
* Michael Gerow (gerow)
* Vytautas Šaltenis (rtfb)
* James Cooper (coopernurse)
* Gyepi Sam (gyepisam)
* Matt Sherman (clipperhouse)
* runner_mei
* John Luebs (jkl1337)
* Luke Hutton (lukehutton)
* Kevin Gorjan (kevingorjan)
* Brendan Fosberry (Fozz)
* Nate Guerin (gusennan)
