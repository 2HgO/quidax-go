package requests

type EditSubAccountDetailsRequest struct {
	UserID      string `uri:"user_id" validate:"required"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}
