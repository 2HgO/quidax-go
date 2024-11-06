package services

import (
	"database/sql"

	tdb "github.com/tigerbeetle/tigerbeetle-go"
	"go.uber.org/zap"
)

type service struct {
	transactionDB  tdb.Client
	dataDB         *sql.DB
	accountService AccountService
	swapService    InstantSwapService
	walletService  WalletService
	log            *zap.Logger
}

var Ledgers = map[uint32]string{
	1: "ngn",
	2: "usdt",
	3: "usdc",
	4: "eth",
	5: "bnb",
	6: "sol",
	7: "btc",
}

var LedgerIDs = map[string]uint32{
	"ngn":  1,
	"usdt": 2,
	"usdc": 3,
	"eth":  4,
	"bnb":  5,
	"sol":  6,
	"btc":  7,
}
