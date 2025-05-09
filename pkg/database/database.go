package database

import (
	"database/sql"
	"github.com/sakojpa/tasker/config"
	_ "modernc.org/sqlite"
	"os"
)

var dbConn *sql.DB

const (
	SCHEMA string = `create table scheduler (
  id integer not null constraint scheduler_id_pk primary key autoincrement,
  date CHAR(8) default '' not null,
  title varchar(64) default '' not null,
  comment text default '',
  repeat varchar(128)
);
create index scheduler_date_index on scheduler (date);
`
)

// Init initializes the database connection or creates a new database if it doesn't exist.
func Init(c *config.Config) error {
	_, err := os.Stat(c.DB.FilePath)
	if err != nil {
		if err := dbCreate(c); err != nil {
			return err
		}
	} else {
		dbConn, err = dbConnect(c)
	}
	return nil
}

func dbCreate(c *config.Config) error {
	dbFile, err := os.Create(c.DB.FilePath)
	if err != nil {
		return err
	}
	defer dbFile.Close()
	dbConn, err = dbConnect(c)
	if err != nil {
		return err
	}
	_, err = dbConn.Exec(SCHEMA)
	if err != nil {
		return err
	}
	return nil
}

func dbConnect(c *config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", c.DB.FilePath)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// DbClose closes the database connection safely.
func DbClose() error {
	if dbConn != nil {
		err := dbConn.Close()
		if err != nil {
			return err
		}
		dbConn = nil
	}
	return nil
}
