package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/2HgO/quidax-go/models"
	"github.com/2HgO/quidax-go/types/responses"
	"go.uber.org/zap"
)

type WebhookService interface {
	SendWalletUpdatedEvent(models.WebhookDetails, *responses.UserWalletResponseData) (self WebhookService)
	SendInstantSwapCompletedEvent(models.WebhookDetails, *responses.InstantSwapResponseData) (self WebhookService)
	SendInstantSwapFailedEvent(models.WebhookDetails, *responses.InstantSwapResponseData) (self WebhookService)
	SendInstantSwapReversedEvent(models.WebhookDetails, *responses.InstantSwapResponseData) (self WebhookService)
	SendWithdrawalSuccessfulEvent(models.WebhookDetails, *responses.WithdrawalResponseData) (self WebhookService)
	SendWithdrawalRejectedEvent(models.WebhookDetails, *responses.WithdrawalResponseData) (self WebhookService)
	SendDepositSuccessfulEvent(models.WebhookDetails, *responses.DepositResponseData) (self WebhookService)
}

type webhookService struct {
	service
}

func NewWebhookService(accountService AccountService, log *zap.Logger) WebhookService {
	return &webhookService{
		service{
			accountService: accountService,
			log:            log,
		},
	}
}

func (w *webhookService) doRequest(url string, body *bytes.Buffer, key *string) (error, bool) {
	time.Sleep(time.Second * 5)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err, false
	}

	if key != nil {
		now := time.Now().Unix()
		data := strings.ReplaceAll(body.String(), "/", "\\/")
		payload := fmt.Sprintf("%d.%s", now, data)
		// todo
		mac := hmac.New(sha256.New, []byte(*key))
		if _, err := mac.Write([]byte(payload)); err != nil {
			return err, false
		}
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("quidax-signature", fmt.Sprintf("ts=%d,sig=%s", now, signature))
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if res != nil {
		resData, _ := io.ReadAll(res.Body)
		w.log.Info("response from callback", zap.String("Response Data", string(resData)))
	}
	return err, (res != nil && res.StatusCode < 300)
}

func (w *webhookService) sendEvent(whDetails models.WebhookDetails, eventType models.WebhookEvent, eventData any) (self WebhookService) {
	if whDetails.CallbackURL == nil {
		return w
	}
	w.log.Info("dispatching event...", zap.String("Event Type", eventType.String()))

	event := &models.Webhook{
		Event: eventType,
		Data:  eventData,
	}

	data, err := json.Marshal(event)
	if err != nil {
		// todo
		w.log.Error("encoding request body", zap.Error(err))
		return w
	}

	err, ok := w.doRequest(*whDetails.CallbackURL, bytes.NewBuffer(data), whDetails.WebhookKey)
	if err != nil {
		//todo
		w.log.Error("dispatching request", zap.Error(err))
		return w
	}

	if ok {
		return w
	}

	// todo: schedule event for single retry
	return w
}

func (w *webhookService) SendWalletUpdatedEvent(whDetails models.WebhookDetails, wallet *responses.UserWalletResponseData) (self WebhookService) {
	return w.sendEvent(whDetails, models.WalletUpdated_WebhookEvent, wallet)
}

func (w *webhookService) SendInstantSwapCompletedEvent(whDetails models.WebhookDetails, swap *responses.InstantSwapResponseData) (self WebhookService) {
	return w.sendEvent(whDetails, models.SwapTransactionCompleted_WebhookEvent, swap)
}

func (w *webhookService) SendInstantSwapFailedEvent(whDetails models.WebhookDetails, swap *responses.InstantSwapResponseData) (self WebhookService) {
	return w.sendEvent(whDetails, models.SwapTransactionFailed_WebhookEvent, swap)
}

func (w *webhookService) SendInstantSwapReversedEvent(whDetails models.WebhookDetails, swap *responses.InstantSwapResponseData) (self WebhookService) {
	return w.sendEvent(whDetails, models.SwapTransactionReversed_WebhookEvent, swap)
}

func (w *webhookService) SendWithdrawalSuccessfulEvent(whDetails models.WebhookDetails, withdrawal *responses.WithdrawalResponseData) (self WebhookService) {
	return w.sendEvent(whDetails, models.WithdrawalSuccessful_WebhookEvent, withdrawal)
}

func (w *webhookService) SendWithdrawalRejectedEvent(whDetails models.WebhookDetails, withdrawal *responses.WithdrawalResponseData) (self WebhookService) {
	return w.sendEvent(whDetails, models.WithdrawalRejected_WebhookEvent, withdrawal)
}

func (w *webhookService) SendDepositSuccessfulEvent(whDetails models.WebhookDetails, data *responses.DepositResponseData) (self WebhookService) {
	w.sendEvent(whDetails, models.DepositConfirmation_WebhookEvent, data)
	time.Sleep(time.Second * 5)
	return w.sendEvent(whDetails, models.DepositSuccessful_WebhookEvent, data)
}
