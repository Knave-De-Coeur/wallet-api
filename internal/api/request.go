package api

import (
	"github.com/shopspring/decimal"
)

// LoginRequest is the parsed struct of the /login endpoint
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BalanceRequest struct {
	UserId   int `json:"user_id"`
	WalletId int `json:"wallet_id"`
}

type CreditRequest struct {
	UserId   int             `json:"user_id"`
	WalletId int             `json:"wallet_id"`
	Amount   decimal.Decimal `json:"amount"`
}

type DebitRequest struct {
	UserId   int             `json:"user_id"`
	WalletId int             `json:"wallet_id"`
	Amount   decimal.Decimal `json:"amount"`
}
