package services

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"wallet-api/internal/api"
)

func TestBalance(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *api.BalanceRequest
		ExpectedResult *api.BalanceResponse
		ExpectedErr    bool
	}{
		{
			Name: "Simple test",
			Input: &api.BalanceRequest{
				UserId:   1,
				WalletId: 1,
			},
			ExpectedResult: &api.BalanceResponse{
				UserID:   1,
				WalletID: 1,
				Balance:  decimal.NewFromInt(100),
			},
			ExpectedErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			res, err := walletService.Balance(test.Input.UserId, test.Input.WalletId)
			if test.ExpectedErr {
				require.Error(t, err)
			}

			assert.Equal(t, test.ExpectedResult, res)
		})
	}
}

func TestCredit(t *testing.T) {

}

func TestDebit(t *testing.T) {

}
