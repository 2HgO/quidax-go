package services

import (
	"context"
	"database/sql"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/models"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	// "github.com/google/uuid"
	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"go.uber.org/zap"
)

type WalletService interface {
	FetchUserWallets(ctx context.Context, req *requests.FetchUserWalletsRequest) (*responses.Response[[]*responses.UserWalletResponseData], error)
	FetchUserWallet(ctx context.Context, req *requests.FetchUserWalletRequest) (*responses.Response[*responses.UserWalletResponseData], error)
}

func NewWalletService(txDatabase tdb.Client, dataDatabase *sql.DB, accountService AccountService, log *zap.Logger) WalletService {
	return &walletService{
		service: service{
			transactionDB:  txDatabase,
			dataDB:         dataDatabase,
			accountService: accountService,
			log:            log,
		},
	}
}

type walletService struct {
	service
}

func (w *walletService) FetchUserWallets(ctx context.Context, req *requests.FetchUserWalletsRequest) (*responses.Response[[]*responses.UserWalletResponseData], error) {
	parent := ctx.Value("user").(*models.Account)
	if req.UserID == parent.ID {
		req.UserID = "me"
	}

	user, err := w.accountService.FetchAccountDetails(ctx, &requests.FetchAccountDetailsRequest{UserID: req.UserID})
	if err != nil {
		return nil, err
	}

	rows, err := sq.
		Select("id", "account_id", "token").
		From("wallets").
		Where(sq.Eq{"account_id": user.Data.ID}).
		RunWith(w.dataDB).
		QueryContext(ctx)
	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	var walletsMap = map[string]*models.Wallet{}
	for rows.Next() {
		wallet := &models.Wallet{}
		err = rows.Scan(&wallet.ID, &wallet.AccountID, &wallet.Token)
		if err != nil {
			return nil, errors.HandleDataDBError(err)
		}
		walletsMap[wallet.ID] = wallet
	}

	res, err := w.transactionDB.QueryAccounts(tdb_types.QueryFilter{UserData128: tdb_types.BytesToUint128(uuid.MustParse(user.Data.ID)), Limit: uint32(len(Ledgers))})
	if err != nil {
		return nil, err
	}

	data := make([]*responses.UserWalletResponseData, len(res))
	for i := range res {
		credits := res[i].CreditsPosted.BigInt()
		debits := res[i].DebitsPosted.BigInt()
		pendingDebits := res[i].DebitsPending.BigInt()
		balance := credits.Sub(&credits, &debits)
		balance = balance.Sub(balance, &pendingDebits)
		wallet := walletsMap[res[i].ID.String()]
		data[i] = &responses.UserWalletResponseData{
			ID:            wallet.ID,
			Currency:      wallet.Token,
			Balance:       float64(balance.Uint64()) * 1e-9,
			LockedBalance: float64(pendingDebits.Uint64()) * 1e-9,
			User:          user.Data,
		}
	}

	return &responses.Response[[]*responses.UserWalletResponseData]{
		Status: "successful",
		Data:   data,
	}, nil
}

func (w *walletService) FetchUserWallet(ctx context.Context, req *requests.FetchUserWalletRequest) (*responses.Response[*responses.UserWalletResponseData], error) {
	parent := ctx.Value("user").(*models.Account)
	if req.UserID == parent.ID {
		req.UserID = "me"
	}

	user, err := w.accountService.FetchAccountDetails(ctx, &requests.FetchAccountDetailsRequest{UserID: req.UserID})
	if err != nil {
		return nil, err
	}

	row := sq.
		Select("id", "account_id", "token").
		From("wallets").
		Where(sq.Eq{"account_id": user.Data.ID, "token": req.Currency}).
		RunWith(w.dataDB).
		QueryRowContext(ctx)
	if row == nil {
		return nil, errors.NewNotFoundError("wallet not found")
	}

	wallet := &models.Wallet{}
	err = row.Scan(&wallet.ID, &wallet.AccountID, &wallet.Token)
	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	walletId, err := tdb_types.HexStringToUint128(wallet.ID)
	if err != nil {
		return nil, err
	}
	res, err := w.transactionDB.LookupAccounts([]tdb_types.Uint128{walletId})
	if err != nil {
		return nil, err
	}
	if len(res) < 1 {
		return nil, errors.NewNotFoundError("wallet not found")
	}

	credits := res[0].CreditsPosted.BigInt()
	debits := res[0].DebitsPosted.BigInt()
	pendingDebits := res[0].DebitsPending.BigInt()
	balance := credits.Sub(&credits, &debits)
	balance = balance.Sub(balance, &pendingDebits)
	data := &responses.UserWalletResponseData{
		ID:            wallet.ID,
		Currency:      wallet.Token,
		Balance:       float64(balance.Uint64()) * 1e-9,
		LockedBalance: float64(pendingDebits.Uint64()) * 1e-9,
		User:          user.Data,
	}

	return &responses.Response[*responses.UserWalletResponseData]{
		Status: "successful",
		Data:   data,
	}, nil
}
