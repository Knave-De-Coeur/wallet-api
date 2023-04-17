package api

// MessageResponse is a generic response struct that'll be marshalled to json and sent to the requester
type MessageResponse struct {
	Message string `json:"message"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type BalanceResponse struct {
	User    User
	Balance int64 `json:"balance"`
}

type CreditResponse struct {
	User    User
	Balance int64 `json:"balance"`
}

type DebitResponse struct {
	User    User
	Balance int64 `json:"balance"`
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
