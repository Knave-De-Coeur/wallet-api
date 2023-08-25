package services

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/magiconair/properties/assert"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wallet-api/internal/api"
	"wallet-api/internal/utils"
)

type balanceTestCase struct {
	Name           string
	Input          *api.BalanceRequest
	ExpectedResult *api.BalanceResponse
	ExpectedErr    bool
	SqlMock        func(test balanceTestCase) bool
	RedisMock      func(test balanceTestCase) bool
}

type creditTestCase struct {
	Name           string
	Input          *api.CreditRequest
	ExpectedResult *api.CreditResponse
	ExpectedErr    bool
	SqlMock        func(test creditTestCase) bool
	RedisMock      func(test creditTestCase) bool
}

type debitTestCase struct {
	Name           string
	Input          *api.DebitRequest
	ExpectedResult *api.DebitResponse
	ExpectedErr    bool
	SqlMock        func(test debitTestCase) bool
	RedisMock      func(test debitTestCase) bool
}

func TestBalance(t *testing.T) {
	emptyFunds, _ := decimal.NewFromString("0.00")

	testCases := []balanceTestCase{
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
			SqlMock: func(test balanceTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
						AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 10000))
				return true
			},
			RedisMock: func(test balanceTestCase) bool {
				key := utils.GenerateRedisKey(test.Input.WalletId)
				redisClientMock.ExpectGet(key).RedisNil()
				redisClientMock.Regexp().ExpectSet(key, `^[0-9]`, 60*time.Minute).SetVal("ok")
				return true
			},
		},
		{
			Name: "Return empty balance",
			Input: &api.BalanceRequest{
				UserId:   1,
				WalletId: 1,
			},
			ExpectedResult: &api.BalanceResponse{
				UserID:   1,
				WalletID: 1,
				Balance:  emptyFunds,
			},
			ExpectedErr: false,
			SqlMock: func(test balanceTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
						AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 0))
				return true
			},
			RedisMock: func(test balanceTestCase) bool {
				key := utils.GenerateRedisKey(test.Input.WalletId)
				redisClientMock.ExpectGet(key).RedisNil()
				redisClientMock.Regexp().ExpectSet(key, `^[0-9]`, 60*time.Minute).SetVal("ok")
				return true
			},
		},
		{
			Name: "Incorrect user_id or wallet_id",
			Input: &api.BalanceRequest{
				UserId:   1,
				WalletId: 9,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock: func(test balanceTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnError(gorm.ErrRecordNotFound)
				return true
			},
			RedisMock: func(test balanceTestCase) bool {
				key := utils.GenerateRedisKey(test.Input.WalletId)
				redisClientMock.ExpectGet(key).RedisNil()
				return true
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			// set up the mock results
			sqlM := test.RedisMock(test)
			redisM := test.SqlMock(test)

			res, err := walletService.Balance(test.Input.UserId, test.Input.WalletId)
			if sqlM {
				if sqlMockErr := sqlMock.ExpectationsWereMet(); sqlMockErr != nil {
					t.Errorf("there were unfulfilled expectations: %s", sqlMockErr)
				}
			}
			if redisM {
				if redisMockErr := redisClientMock.ExpectationsWereMet(); redisMockErr != nil {
					t.Errorf("there were unfulfilled expectations: %s", redisMockErr)
				}
			}
			if test.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, res.Balance.String(), test.ExpectedResult.Balance.String())
			}
		})
	}
}

func TestCredit(t *testing.T) {
	emptyFunds, _ := decimal.NewFromString("0.00")
	amount, _ := decimal.NewFromString("14.65")
	amountExpected, _ := decimal.NewFromString("114.65")
	negativeAmount := amount.Neg()

	testCases := []creditTestCase{
		{
			Name: "Simple test, add funds",
			Input: &api.CreditRequest{
				UserId:   1,
				WalletId: 1,
				Amount:   amount,
			},
			ExpectedResult: &api.CreditResponse{
				UserID:   1,
				WalletID: 1,
				Balance:  amountExpected,
			},
			ExpectedErr: false,
			SqlMock: func(test creditTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
						AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 10000))

				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(regexp.QuoteMeta("UPDATE `wallets` SET `funds`=? WHERE `id` = ?")).
					WithArgs(11465, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				return true
			},
			RedisMock: func(test creditTestCase) bool {
				key := utils.GenerateRedisKey(test.Input.WalletId)
				redisClientMock.Regexp().ExpectSet(key, `^[0-9]`, 60*time.Minute).SetVal("ok")
				return true
			},
		},
		{
			Name: "Attempt to add Negative funds",
			Input: &api.CreditRequest{
				UserId:   1,
				WalletId: 1,
				Amount:   negativeAmount,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock:        func(test creditTestCase) bool { return false },
			RedisMock:      func(test creditTestCase) bool { return false },
		},
		{
			Name: "Attempt to add Empty Funds",
			Input: &api.CreditRequest{
				UserId:   1,
				WalletId: 1,
				Amount:   emptyFunds,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock:        func(test creditTestCase) bool { return false },
			RedisMock:      func(test creditTestCase) bool { return false },
		},
		{
			Name: "Incorrect user_id or wallet_id",
			Input: &api.CreditRequest{
				UserId:   1,
				WalletId: 9,
				Amount:   amount,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock: func(test creditTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnError(gorm.ErrRecordNotFound)
				return true
			},
			RedisMock: func(test creditTestCase) bool { return false },
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			// set up the mock results
			sqlM := test.RedisMock(test)
			redisM := test.SqlMock(test)

			res, err := walletService.Credit(test.Input)

			if sqlM {
				if sqlMockErr := sqlMock.ExpectationsWereMet(); sqlMockErr != nil {
					t.Errorf("there were unfulfilled expectations: %s", sqlMockErr)
				}
			}
			if redisM {
				if redisMockErr := redisClientMock.ExpectationsWereMet(); redisMockErr != nil {
					t.Errorf("there were unfulfilled expectations: %s", redisMockErr)
				}
			}

			if test.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, res.Balance.String(), test.ExpectedResult.Balance.String())
			}
		})
	}
}

func TestDebit(t *testing.T) {
	emptyFunds, _ := decimal.NewFromString("0.00")
	massiveAmount, _ := decimal.NewFromString("100.01")
	amount, _ := decimal.NewFromString("14.65")
	amountExpected, _ := decimal.NewFromString("85.35")
	negativeAmount := amount.Neg()

	testCases := []debitTestCase{
		{
			Name: "Simple test, deduct funds",
			Input: &api.DebitRequest{
				UserId:   1,
				WalletId: 1,
				Amount:   amount,
			},
			ExpectedResult: &api.DebitResponse{
				UserID:   1,
				WalletID: 1,
				Balance:  amountExpected,
			},
			ExpectedErr: false,
			SqlMock: func(test debitTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
						AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 10000))

				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(regexp.QuoteMeta("UPDATE `wallets` SET `funds`=? WHERE `id` = ?")).
					WithArgs(8535, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				return true
			},
			RedisMock: func(test debitTestCase) bool {
				key := utils.GenerateRedisKey(test.Input.WalletId)
				redisClientMock.Regexp().ExpectSet(key, `^[0-9]`, 60*time.Minute).SetVal("ok")
				return true
			},
		},
		{
			Name: "Attempt to deduct amount larger than wallet contains",
			Input: &api.DebitRequest{
				UserId:   1,
				WalletId: 1,
				Amount:   massiveAmount,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock: func(test debitTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "funds"}).
						AddRow(test.Input.WalletId, test.Input.UserId, "Wallet 1", 10000))
				return true
			},
			RedisMock: func(test debitTestCase) bool { return false },
		},
		{
			Name: "Attempt to deduct Negative funds",
			Input: &api.DebitRequest{
				UserId:   1,
				WalletId: 1,
				Amount:   negativeAmount,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock:        func(test debitTestCase) bool { return false },
			RedisMock:      func(test debitTestCase) bool { return false },
		},
		{
			Name: "Attempt to deduct Empty Funds",
			Input: &api.DebitRequest{
				UserId:   1,
				WalletId: 1,
				Amount:   emptyFunds,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock:        func(test debitTestCase) bool { return false },
			RedisMock:      func(test debitTestCase) bool { return false },
		},
		{
			Name: "Incorrect user_id or wallet_id",
			Input: &api.DebitRequest{
				UserId:   1,
				WalletId: 9,
				Amount:   amount,
			},
			ExpectedResult: nil,
			ExpectedErr:    true,
			SqlMock: func(test debitTestCase) bool {
				sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`user_id`,`name`,`funds` FROM `wallets` WHERE `id` = ? AND `user_id` = ? ORDER BY `wallets`.`id` LIMIT 1")).
					WithArgs(test.Input.WalletId, test.Input.UserId).
					WillReturnError(gorm.ErrRecordNotFound)
				return true
			},
			RedisMock: func(test debitTestCase) bool { return false },
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			// set up the mock results
			sqlM := test.RedisMock(test)
			redisM := test.SqlMock(test)

			res, err := walletService.Debit(test.Input)

			if sqlM {
				if sqlMockErr := sqlMock.ExpectationsWereMet(); sqlMockErr != nil {
					t.Errorf("there were unfulfilled expectations: %s", sqlMockErr)
				}
			}
			if redisM {
				if redisMockErr := redisClientMock.ExpectationsWereMet(); redisMockErr != nil {
					t.Errorf("there were unfulfilled expectations: %s", redisMockErr)
				}
			}

			if test.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, res.Balance.String(), test.ExpectedResult.Balance.String())
			}
		})
	}
}
