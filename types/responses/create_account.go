package responses

import "github.com/2HgO/quidax-go/models"

type CreateAccountResponseData struct {
	User  *models.Account     `json:"user"`
	Token *models.AccessToken `json:"token"`
}
