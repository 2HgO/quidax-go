package db

import (
	"log"

	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func GetTxDBConnection() tdb.Client {
	client, err := tdb.NewClient(tdb_types.ToUint128(0), []string{"3003"})
	if err != nil {
		log.Panicln(err)
	}

	return client
}
