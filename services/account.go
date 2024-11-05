package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lucsky/cuid"
	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"golang.org/x/crypto/bcrypt"

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
	UpdateWebHookURL(ctx context.Context, req *requests.UpdateWebhookURLRequest) (*responses.Response[any], error)
	GenerateToken(ctx context.Context, req *requests.GenerateTokenRequest) (*responses.Response[*models.AccessToken], error)
	GetAccountByAccessToken(ctx context.Context, token string) (*models.Account, error)
}

func NewAccountService(txDatabase tdb.Client, dataDatabase *sql.DB) AccountService {
	return &accountService{
		service: service{
			transactionDB: txDatabase,
			dataDB:        dataDatabase,
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
				History:                    true,
				DebitsMustNotExceedCredits: true,
				Linked:                     len(wallets) < (len(Ledgers) - 2),
			}.ToUint16(),
			Ledger:      ledgerId,
			Code:        1,
			UserData128: tdb_types.BytesToUint128(accountID),
		})
	}

	txRes, err := a.transactionDB.CreateAccounts(wallets)
	if err != nil {
		return nil, err
	}
	if len(txRes) > 0 {
		return nil, errors.New("failed to create user wallet")
	}
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
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
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
	stmt := sq.
		Select("id", "sn", "display_name", "email", "first_name", "last_name", "callback_url", "created_at", "updated_at").
		From("accounts")

	switch req.UserID {
	case "me":
		stmt = stmt.Where(sq.Eq{"is_main_account": true, "id": parent.ID})
	default:
		stmt = stmt.Where(sq.Or{sq.Eq{"id": req.UserID, "parent_id": parent.ID}, sq.Eq{"id": req.UserID, "is_main_account": true}})
	}

	row := stmt.
		Limit(1).
		RunWith(a.dataDB).
		QueryRow()

	if row == nil {
		return nil, errors.New("user not found")
	}
	var account = &models.Account{}
	err := row.Scan(&account.ID, &account.SN, &account.DisplayName, &account.Email, &account.FirstName, &account.LastName, &account.CallbackURL, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &responses.Response[*models.Account]{
		Status: "successful",
		Data:   account,
	}, nil
}

func (a *accountService) GetAccountByAccessToken(ctx context.Context, token string) (*models.Account, error) {
	row := sq.
		Select("accounts.id", "email", "callback_url").
		From("access_tokens").
		Join("accounts on access_tokens.account_id = accounts.id").
		Where(sq.Eq{"token": token}).
		Limit(1).
		RunWith(a.dataDB).
		QueryRow()

	if row == nil {
		return nil, errors.New("token not found")
	}
	var account = &models.Account{}
	err := row.Scan(&account.ID, &account.Email, &account.CallbackURL)
	if err != nil {
		return nil, err
	}

	return account, nil
}
