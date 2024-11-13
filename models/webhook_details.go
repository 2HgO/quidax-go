package models

type WebhookDetails struct {
	ID          string
	CallbackURL *string
	WebhookKey  *string
}
