package models

type Wallet struct {
	ID string `json:"id"`
	// ? maybe change to uuid.UUID
	AccountID string `json:"account_id"`
	Token     string `json:"token"`
}
