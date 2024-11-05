package requests

type GenerateTokenRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
