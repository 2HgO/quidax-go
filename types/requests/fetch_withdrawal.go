package requests

type FetchWithdrawalRequest struct {
	UserID       string `uri:"user_id"`
	WithdrawalID string `uri:"withdrawal_id"`
	Reference    string `uri:"reference"`
}
