package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lucsky/cuid"
	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/models"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
)

type AccountService interface {
	CreateSubAccount(ctx context.Context, req *requests.CreateSubAccountRequest) (*responses.Response[*models.Account], error)
	EditSubAccountDetails(ctx context.Context, req *requests.EditSubAccountDetailsRequest) (*responses.Response[*models.Account], error)
	FetchAllSubAccounts(ctx context.Context, req *requests.FetchAllSubAccountsRequest) (*responses.Response[[]*models.Account], error)
	FetchAccountDetails(ctx context.Context, req *requests.FetchAccountDetailsRequest) (*responses.Response[*models.Account], error)

	CreateAccount(ctx context.Context, req *requests.CreateAccountRequest) (*responses.Response[*responses.CreateAccountResponseData], error)
	UpdateWebHookURL(ctx context.Context, req *requests.UpdateWebhookURLRequest) error
	GenerateToken(ctx context.Context, req *requests.GenerateTokenRequest) (*responses.Response[*models.AccessToken], error)
	GetAccountByAccessToken(ctx context.Context, token string) (*models.Account, error)
}

func NewAccountService(txDatabase tdb.Client, dataDatabase *sql.DB, log *zap.Logger) AccountService {
	return &accountService{
		service: service{
			transactionDB: txDatabase,
			dataDB:        dataDatabase,
			log:           log,
		},
	}
}

type accountService struct {
	AccountService
	service
}

func (a *accountService) CreateAccount(ctx context.Context, req *requests.CreateAccountRequest) (*responses.Response[*responses.CreateAccountResponseData], error) {
	now := time.Now()
	accountID := uuid.New()
	account := &models.Account{
		ID:          accountID.String(),
		SN:          cuid.New(),
		DisplayName: req.DisplayName,
		Email:       strings.ToLower(req.Email),
		FirstName:   strings.ToTitle(req.FirstName),
		LastName:    strings.ToTitle(req.LastName),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	tx, err := a.dataDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	// * create user account
	_, err = sq.
		Insert("accounts").
		Columns("id", "sn", "display_name", "email", "password", "first_name", "last_name", "created_at", "updated_at", "is_main_account").
		Values(account.ID, account.SN, account.DisplayName, account.Email, string(password), account.FirstName, account.LastName, now, now, true).
		RunWith(tx).
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	accessToken := &models.AccessToken{
		ID:          uuid.NewString(),
		Name:        "Default Token",
		Description: "default token for user requests",
		AccountID:   account.ID,
		Token:       "pub_test_" + cuid.Slug(),
	}

	// * create user access token to authenticate requests
	_, err = sq.
		Insert("access_tokens").
		Columns("id", "name", "description", "account_id", "token").
		Values(accessToken.ID, accessToken.Name, accessToken.Description, accessToken.AccountID, accessToken.Token).
		RunWith(tx).
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	wallets := make([]tdb_types.Account, 0, len(Ledgers))
	for ledgerId := range Ledgers {
		wallets = append(wallets, tdb_types.Account{
			ID: tdb_types.ID(),
			Flags: tdb_types.AccountFlags{
				History: true,
				// * allow debits to exceed credits for main account
				// ?todo: undo later and use system transaction account to monitor amount of value in circulation at a quick glance
				DebitsMustNotExceedCredits: true,
				Linked: len(wallets) < (len(Ledgers) - 1),
			}.ToUint16(),
			Ledger:      ledgerId,
			Code:        1,
			UserData128: tdb_types.BytesToUint128(accountID),
		})
	}

	// * create wallet accounts in financial transaction database
	txRes, err := a.transactionDB.CreateAccounts(wallets)
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}
	if len(txRes) > 0 {
		return nil, errors.NewUnknownError(txRes[0].Result.String())
	}

	// * store wallets ref in wallets collection
	walletsInsertStmt := sq.
		Insert("wallets").
		Columns("id", "account_id", "token")
	for _, wallet := range wallets {
		walletsInsertStmt = walletsInsertStmt.
			Values(wallet.ID.String(), account.ID, Ledgers[wallet.Ledger])
	}

	_, err = walletsInsertStmt.
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	return &responses.Response[*responses.CreateAccountResponseData]{
		Status:  "successful",
		Message: "Account Created successfully",
		Data: &responses.CreateAccountResponseData{
			User:  account,
			Token: accessToken,
		},
	}, nil
}

func (a *accountService) FetchAccountDetails(ctx context.Context, req *requests.FetchAccountDetailsRequest) (*responses.Response[*models.Account], error) {
	parent := ctx.Value("user").(*models.Account)
	if req.UserID == parent.ID {
		req.UserID = "me"
	}
	stmt := sq.
		Select("id", "sn", "display_name", "email", "first_name", "last_name", "callback_url", "created_at", "updated_at").
		From("accounts")

	switch req.UserID {
	case "me":
		stmt = stmt.Where(sq.Eq{"is_main_account": true, "id": parent.ID})
	default:
		if ctx.Value("skip_check") == nil {
			stmt = stmt.Where(sq.Eq{"id": req.UserID, "parent_id": parent.ID})
		} else {
			stmt = stmt.Where(sq.Eq{"id": req.UserID})
		}
	}

	row := stmt.
		Limit(1).
		RunWith(a.dataDB).
		QueryRowContext(ctx)

	if row == nil {
		return nil, errors.NewNotFoundError("user not found")
	}
	var account = &models.Account{}
	err := row.Scan(&account.ID, &account.SN, &account.DisplayName, &account.Email, &account.FirstName, &account.LastName, &account.CallbackURL, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	return &responses.Response[*models.Account]{
		Status: "successful",
		Data:   account,
	}, nil
}

func (a *accountService) GetAccountByAccessToken(ctx context.Context, token string) (*models.Account, error) {
	row := sq.
		Select("accounts.id", "email", "callback_url", "display_name").
		From("access_tokens").
		Join("accounts on access_tokens.account_id = accounts.id").
		Where(sq.Eq{"token": token}).
		Limit(1).
		RunWith(a.dataDB).
		QueryRowContext(ctx)

	if row == nil {
		return nil, errors.NewNotFoundError("token not found")
	}
	var account = &models.Account{}
	err := row.Scan(&account.ID, &account.Email, &account.CallbackURL, &account.DisplayName)
	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	return account, nil
}

func (a *accountService) UpdateWebHookURL(ctx context.Context, req *requests.UpdateWebhookURLRequest) error {
	parent := ctx.Value("user").(*models.Account)
	_, err := sq.
		Update("accounts").
		Set("callback_url", req.CallbackURL).
		Where(sq.Eq{"id": parent.ID, "is_main_account": true}).
		RunWith(a.dataDB).
		ExecContext(ctx)
	if err != nil {
		return errors.HandleDataDBError(err)
	}

	return nil
}

func (a *accountService) CreateSubAccount(ctx context.Context, req *requests.CreateSubAccountRequest) (*responses.Response[*models.Account], error) {
	parent := ctx.Value("user").(*models.Account)

	now := time.Now()
	accountID := uuid.New()
	account := &models.Account{
		ID:          accountID.String(),
		SN:          cuid.New(),
		DisplayName: parent.DisplayName,
		Email:       strings.ToLower(req.Email),
		FirstName:   strings.ToTitle(req.FirstName),
		LastName:    strings.ToTitle(req.LastName),
		ParentID:    &parent.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tx, err := a.dataDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	// * create sub account user
	_, err = sq.
		Insert("accounts").
		Columns("id", "sn", "display_name", "email", "first_name", "last_name", "created_at", "updated_at", "is_main_account", "parent_id").
		Values(account.ID, account.SN, parent.DisplayName, account.Email, account.FirstName, account.LastName, now, now, false, parent.ID).
		RunWith(tx).
		ExecContext(ctx)

	if err != nil {
		return nil, err
	}

	wallets := make([]tdb_types.Account, 0, len(Ledgers))
	for ledgerId := range Ledgers {
		wallets = append(wallets, tdb_types.Account{
			ID: tdb_types.ID(),
			Flags: tdb_types.AccountFlags{
				History:                    true,
				DebitsMustNotExceedCredits: true,
				Linked:                     len(wallets) < (len(Ledgers) - 1),
			}.ToUint16(),
			Ledger:      ledgerId,
			Code:        1,
			UserData128: tdb_types.BytesToUint128(accountID),
		})
	}

	// * create wallet accounts in financial transaction database
	txRes, err := a.transactionDB.CreateAccounts(wallets)
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}
	if len(txRes) > 0 {
		return nil, errors.NewUnknownError(txRes[0].Result.String())
	}

	// * store wallets ref in wallets collection
	walletsInsertStmt := sq.
		Insert("wallets").
		Columns("id", "account_id", "token")
	for _, wallet := range wallets {
		walletsInsertStmt = walletsInsertStmt.
			Values(wallet.ID.String(), account.ID, Ledgers[wallet.Ledger])
	}

	_, err = walletsInsertStmt.
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	return &responses.Response[*models.Account]{
		Status:  "successful",
		Message: "Account Created successfully",
		Data:    account,
	}, nil
}

func (a *accountService) EditSubAccountDetails(ctx context.Context, req *requests.EditSubAccountDetailsRequest) (*responses.Response[*models.Account], error) {
	parent := ctx.Value("user").(*models.Account)
	if req.UserID == parent.ID {
		req.UserID = "me"
	}

	stmt := sq.
		Update("accounts")

	if req.FirstName != "" {
		stmt = stmt.Set("first_name", req.FirstName)
	}
	if req.LastName != "" {
		stmt = stmt.Set("last_name", req.LastName)
	}
	if req.PhoneNumber != "" {
		// todo: add phone number to accounts schema
		// stmt = stmt.Set("first_name", req.FirstName)
	}

	switch req.UserID {
	case "me":
		stmt = stmt.Where(sq.Eq{"is_main_account": true, "id": parent.ID})
	default:
		stmt = stmt.Where(sq.Eq{"id": req.UserID, "parent_id": parent.ID})
	}

	_, err := stmt.RunWith(a.dataDB).ExecContext(ctx)

	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	return a.FetchAccountDetails(ctx, &requests.FetchAccountDetailsRequest{UserID: req.UserID})
}

func (a *accountService) FetchAllSubAccounts(ctx context.Context, req *requests.FetchAllSubAccountsRequest) (*responses.Response[[]*models.Account], error) {
	parent := ctx.Value("user").(*models.Account)

	rows, err := sq.
		Select("id", "sn", "display_name", "email", "first_name", "last_name", "callback_url", "created_at", "updated_at").
		From("accounts").
		Where("parent_id", parent.ID).
		RunWith(a.dataDB).
		QueryContext(ctx)
	if err != nil {
		return nil, errors.HandleDataDBError(err)
	}

	res := make([]*models.Account, 0)
	for rows.Next() {
		acc := &models.Account{}
		err := rows.Scan(&acc.ID, &acc.SN, &acc.DisplayName, &acc.Email, &acc.FirstName, &acc.LastName, &acc.CallbackURL, &acc.CreatedAt, &acc.UpdatedAt)
		if err != nil {
			return nil, errors.HandleDataDBError(err)
		}
		res = append(res, acc)
	}

	return &responses.Response[[]*models.Account]{
		Status: "successful",
		Data:   res,
	}, nil
}
