package requests

type UpdateWebhookURLRequest struct {
	CallbackURL string  `json:"callback_url"`
	WebhookKey  *string `json:"webhook_key"`
}
