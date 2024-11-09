package models

import "time"

type Account struct {
	// ? maybe change to uuid.UUID
	ID          string     `json:"id"`
	SN          string     `json:"sn,omitempty"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email,omitempty"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`

	// internal fields
	IsMainAccount bool    `json:"-"`
	ParentID      *string `json:"-"`
	CallbackURL   *string `json:"-"`
	WebhookKey    *string `json:"-"`
	// ? maybe change to uuid.UUID
	Password *string `json:"-"`
}
