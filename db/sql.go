package db

import (
	"database/sql"
	"sync"

	"github.com/2HgO/quidax-go/config"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var dataDb *sql.DB
var dataDBOnce = &sync.Once{}

type dbLogger struct {
	log *zap.Logger
}

func (d *dbLogger) Print(v ...any) {
	d.log.Sugar().Info(v...)
}

func GetDataDBConnection(log *zap.Logger) *sql.DB {
	log.Sugar().Info()
	dataDBOnce.Do(func() {
		cfg := mysql.Config{
			User:      "root",
			Net:       "tcp",
			Addr:      config.DATA_DB_URL, //"127.0.0.1:3306"
			DBName:    "quidax-go",
			ParseTime: true,
			Logger:    &dbLogger{log: log},
		}
		// Get a database handle.
		var err error
		dataDb, err = sql.Open("mysql", cfg.FormatDSN())
		if err != nil {
			log.Sugar().Fatalln(err)
		}

		pingErr := dataDb.Ping()
		if pingErr != nil {
			log.Sugar().Fatalln(pingErr)
		}
	})

	return dataDb
}
