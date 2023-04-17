package pkg

type Wallet struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	User   User   `json:"user"`
	Name   string `json:"name"`
	Funds  int    `json:"funds"`
}
