package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/2HgO/quidax-go/config"
	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func GetTxDBConnection() tdb.Client {
	addr := strings.Split(config.TX_DB_URL, ",")
	client, err := tdb.NewClient(tdb_types.ToUint128(0), addr) //3003
	if err != nil {
		fmt.Println(err.Error())
		log.Panicln(err)
	}

	return client
}
