package db

import (
	"database/sql"
	"log"
	"sync"

	"github.com/go-sql-driver/mysql"
)

var dataDb *sql.DB
var dataDBOnce = &sync.Once{}

func GetDataDBConnection() *sql.DB {
	dataDBOnce.Do(func() {
		cfg := mysql.Config{
			User:      "root",
			Net:       "tcp",
			Addr:      "127.0.0.1:3306",
			DBName:    "quidax-go",
			ParseTime: true,
		}
		// Get a database handle.
		var err error
		dataDb, err = sql.Open("mysql", cfg.FormatDSN())
		if err != nil {
			log.Fatal(err)
		}

		pingErr := dataDb.Ping()
		if pingErr != nil {
			log.Fatal(pingErr)
		}
	})

	return dataDb
}
