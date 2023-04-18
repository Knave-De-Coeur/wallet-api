package services

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
			redisClientMock.ExpectGet(fmt.Sprintf("%d-balance", test.Input.WalletId)).RedisNil()
			sqlMock.ExpectQuery(regexp.QuoteMeta(
				`SELECT id, user_id, name, funds FROM "wallets" 
                                WHERE id = ? AND user_id = ?
                                ORDER BY id
                                LIMIT 1`)).
				WithArgs(test.Input.WalletId, test.Input.UserId).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
					AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 100))
			res, err := walletService.Balance(test.Input.UserId, test.Input.WalletId)
			if test.ExpectedErr {
				require.Error(t, err)
			}

			assert.Equal(t, test.ExpectedResult, res)
		})
	}
}

func TestCredit(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *api.CreditRequest
		ExpectedResult *api.CreditResponse
		ExpectedErr    bool
	}{
		{
			Name: "Simple test",
			Input: &api.CreditRequest{
				UserId:   1,
				WalletId: 1,
			},
			ExpectedResult: &api.CreditResponse{
				UserID:   1,
				WalletID: 1,
				Balance:  decimal.NewFromInt(100),
			},
			ExpectedErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			redisClientMock.ExpectGet(fmt.Sprintf("%d-balance", test.Input.WalletId)).RedisNil()
			sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, name, funds FROM "wallets" WHERE id = ? AND user_id = ? ORDER BY id LIMIT 1`)).
				WithArgs(test.Input.WalletId, test.Input.UserId).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
					AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 100))
			res, err := walletService.Credit(test.Input)
			if test.ExpectedErr {
				require.Error(t, err)
			}

			assert.Equal(t, test.ExpectedResult, res)
		})
	}
}

func TestDebit(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *api.DebitRequest
		ExpectedResult *api.DebitResponse
		ExpectedErr    bool
	}{
		{
			Name: "Simple test",
			Input: &api.DebitRequest{
				UserId:   1,
				WalletId: 1,
			},
			ExpectedResult: &api.DebitResponse{
				UserID:   1,
				WalletID: 1,
				Balance:  decimal.NewFromInt(100),
			},
			ExpectedErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			redisClientMock.ExpectGet(fmt.Sprintf("%d-balance", test.Input.WalletId)).RedisNil()
			sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, name, funds FROM "wallets" WHERE id = ? AND user_id = ? ORDER BY id LIMIT 1`)).
				WithArgs(test.Input.WalletId, test.Input.UserId).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
					AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 100))
			res, err := walletService.Debit(test.Input)
			if test.ExpectedErr {
				require.Error(t, err)
			}

			assert.Equal(t, test.ExpectedResult, res)
		})
	}
}
