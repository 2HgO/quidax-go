package requests

type FetchWithdrawalRequest struct {
	UserID       string `uri:"user_id" validate:"required"`
	WithdrawalID string `uri:"withdrawal_id" validate:"required_without=Reference"`
	Reference    string `uri:"reference" validate:"required_without=WithdrawalID"`
}
