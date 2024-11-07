package models

import (
	"encoding/json"
	"strings"

	"github.com/2HgO/quidax-go/errors"
)

type Withdrawal struct {
	ID              string
	WalletID        string
	AccountID       string
	Ref             *string
	TransactionNote string
	Narration       string
	Reason          *string
	Status          WithdrawalStatus
	Recipient       *Recipient
}

type WithdrawalStatus uint8

const (
	Pending_WithdrawalStatus WithdrawalStatus = iota
	Completed_WithdrawalStatus
	Failed_WithdrawalStatus
)

func (w WithdrawalStatus) String() string {
	switch w {
	case Pending_WithdrawalStatus:
		return "pending"
	case Completed_WithdrawalStatus:
		return "completed"
	case Failed_WithdrawalStatus:
		return "failed"
	default:
		panic("unreachabled")
	}
}

func (w *WithdrawalStatus) UnmarshalJSON(input []byte) error {
	if w == nil {
		w = new(WithdrawalStatus)
	}
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`)
	switch strInput {
	case "pending":
		*w = Pending_WithdrawalStatus
	case "completed":
		*w = Completed_WithdrawalStatus
	case "failed":
		*w = Failed_WithdrawalStatus
	default:
		return errors.NewValidationError("invalid withdrawal status")
	}
	return nil
}

func (w WithdrawalStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}

type RecipientType uint8

const (
	Internal_RecipientType RecipientType = iota
	CoinAddress_RecipientType
)

func (r RecipientType) String() string {
	switch r {
	case Internal_RecipientType:
		return "internal"
	case CoinAddress_RecipientType:
		return "coin_address"
	default:
		panic("unreachable")
	}
}

func (r RecipientType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

type Recipient struct {
	Type    RecipientType     `json:"type"`
	Details *RecipientDetails `json:"details"`
}

type RecipientDetails struct {
	Name           *string `json:"name"`
	DestinationTag *string `json:"destination_tag"`
	Address        *string `json:"address"`
}
