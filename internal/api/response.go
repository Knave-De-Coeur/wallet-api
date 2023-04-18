package api

import (
	"github.com/shopspring/decimal"
)

// MessageResponse is a generic response struct that'll be marshalled to json and sent to the requester
type MessageResponse struct {
	Message string `json:"message"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type BalanceResponse struct {
	UserID   int             `json:"user_ID"`
	WalletID int             `json:"wallet_ID"`
	Balance  decimal.Decimal `json:"balance"`
}

type CreditResponse struct {
	UserID   int             `json:"user_ID"`
	WalletID int             `json:"wallet_ID"`
	Balance  decimal.Decimal `json:"balance"`
}

type DebitResponse struct {
	UserID   int             `json:"user_ID"`
	WalletID int             `json:"wallet_ID"`
	Balance  decimal.Decimal `json:"balance"`
}

func GenerateMessageResponse(message string, res interface{}, err error) *MessageResponse {

	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}

	return &MessageResponse{
		Message: message,
		Result:  res,
		Error:   errorMessage,
	}
}
