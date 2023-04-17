package api

// LoginRequest is the parsed struct of the /login endpoint
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BalanceRequest struct {
	UserId   int `json:"userId"`
	WalletId int `json:"walletId"`
}

type CreditRequest struct {
	UserId   int `json:"userId"`
	WalletId int `json:"walletId"`
	Amount   int `json:"amount"`
}

type DebitRequest struct {
	UserId   int `json:"userId"`
	WalletId int `json:"walletId"`
	Amount   int `json:"amount"`
}
