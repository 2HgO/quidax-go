package models

type AccessToken struct {
	// ? maybe change to uuid.UUID
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AccountID   string `json:"account_id"`
	Token       string `json:"token"`
}
