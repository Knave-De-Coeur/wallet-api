package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"wallet-api/internal/api"
	"wallet-api/internal/pkg"
	"wallet-api/internal/utils"
)

type WalletService struct {
	DBConn      *gorm.DB
	Cache       *redis.Client
	logger      *zap.Logger
	settings    WalletServiceSettings
	UserService *UserService
}

// WalletServiceSettings used to affect code flow
type WalletServiceSettings struct {
	Port              int
	Hostname          string
	RedisCacheTimeout int
}

type WalletServices interface {
	Balance(userID, walletID int) (*api.BalanceResponse, error)
	Credit(creditReq *api.CreditRequest) (*api.CreditResponse, error)
	Debit(debitReq *api.DebitRequest) (*api.DebitResponse, error)
	getUserWalletByID(userID, walletID int) (*pkg.Wallet, error)
	updateUserWalletByID(wallet *pkg.Wallet) error
}

func NewWalletService(dbConn *gorm.DB, rc *redis.Client, logger *zap.Logger, settings WalletServiceSettings, userService *UserService) *WalletService {
	return &WalletService{
		DBConn:      dbConn,
		Cache:       rc,
		logger:      logger,
		settings:    settings,
		UserService: userService,
	}
}

func (w *WalletService) Balance(userID, walletID int) (*api.BalanceResponse, error) {

	redisKey := utils.GenerateRedisKey(walletID)
	walletBalance, err := w.Cache.Get(context.TODO(), redisKey).Result()
	if err != nil && err != redis.Nil {
		w.logger.Error(
			"something went wrong getting the wallet from cache",
			zap.Error(err),
			zap.Int64("user_id", int64(userID)),
			zap.Int64("wallet_id", int64(walletID)),
		)
		return nil, err
	}

	if walletBalance != "" {
		wBalance, err := decimal.NewFromString(walletBalance)
		if err != nil {
			w.logger.Error(
				"something parsing the wallet balance",
				zap.Error(err),
				zap.Int64("user_id", int64(userID)),
				zap.Int64("wallet_id", int64(walletID)),
			)
			return nil, err
		}
		return &api.BalanceResponse{
			UserID:   userID,
			WalletID: walletID,
			Balance:  wBalance.Div(decimal.NewFromInt(100)),
		}, nil
	}

	wallet, err := w.getUserWalletByID(userID, walletID)
	if err != nil {
		return nil, err
	}

	w.logger.Debug("wallet grabbed", zap.Any("wallet", wallet))

	err = w.Cache.Set(context.TODO(), redisKey, wallet.Funds, time.Duration(w.settings.RedisCacheTimeout*int(time.Minute))).Err()
	if err != nil {
		w.logger.Error(
			"something went wrong saving the balance in cache",
			zap.Error(err),
			zap.Any("wallet-balance", wallet.Funds),
		)
		return nil, err
	}

	return &api.BalanceResponse{
		UserID:   userID,
		WalletID: walletID,
		Balance:  decimal.NewFromInt(int64(wallet.Funds)).Div(decimal.NewFromInt(100)),
	}, nil
}

func (w *WalletService) getUserWalletByID(userID, walletID int) (*pkg.Wallet, error) {
	var wallet pkg.Wallet
	// Get all records
	res := w.DBConn.
		Select("id", "user_id", "name", "funds").
		Where(map[string]interface{}{"id": walletID, "user_id": userID}).
		First(&wallet)
	if res.Error != nil {
		w.logger.Error(
			"something went wrong getting the wallet",
			zap.Error(res.Error),
			zap.Int64("user_id", int64(userID)),
			zap.Int64("wallet_id", int64(walletID)),
		)
		return nil, res.Error
	}

	return &wallet, nil
}

func (w *WalletService) updateUserWalletByID(wallet *pkg.Wallet) error {
	// Get all records
	res := w.DBConn.Model(&wallet).Update("funds", wallet.Funds)
	if res.Error != nil {
		w.logger.Error(
			"something went wrong updating the wallet",
			zap.Error(res.Error),
			zap.Any("wallet", wallet),
		)
		return res.Error
	}

	return nil
}

func (w *WalletService) Credit(creditReq *api.CreditRequest) (*api.CreditResponse, error) {

	if creditReq.Amount.IsNegative() || creditReq.Amount.IsZero() {
		err := errors.New(pkg.WrongAmount)
		w.logger.Error(
			"attempted to credit invalid amount",
			zap.Error(err),
			zap.Any("creditReq", creditReq),
		)
		return nil, err
	}

	wallet, err := w.getUserWalletByID(creditReq.UserId, creditReq.WalletId)
	if err != nil {
		return nil, err
	}

	amountToAdd := creditReq.Amount.Mul(decimal.NewFromFloat(100))

	walletFunds := decimal.NewFromInt(int64(wallet.Funds)) // should already be in 100s

	newBalance := walletFunds.Add(amountToAdd)

	wallet.Funds, _ = strconv.Atoi(newBalance.String())

	if err = w.updateUserWalletByID(wallet); err != nil {
		return nil, err
	}

	err = w.Cache.Set(context.TODO(), utils.GenerateRedisKey(creditReq.WalletId), newBalance.String(), time.Duration(w.settings.RedisCacheTimeout*int(time.Minute))).Err()
	if err != nil {
		// in case of error just delete the cached amount
		_ = w.Cache.Del(context.TODO(), utils.GenerateRedisKey(creditReq.WalletId))
		w.logger.Error(
			"something went wrong updating the cached data",
			zap.Error(err),
			zap.Any("wallet", wallet),
		)
	}

	return &api.CreditResponse{
		UserID:   creditReq.UserId,
		WalletID: creditReq.WalletId,
		Balance:  newBalance.Div(decimal.NewFromFloat(100)),
	}, nil
}

func (w *WalletService) Debit(debitReq *api.DebitRequest) (*api.DebitResponse, error) {
	if debitReq.Amount.IsNegative() || debitReq.Amount.IsZero() {
		err := errors.New(pkg.WrongAmount)
		w.logger.Error(
			"attempted to debit invalid amount",
			zap.Error(err),
			zap.Any("debitReq", debitReq),
		)
		return nil, err
	}

	wallet, err := w.getUserWalletByID(debitReq.UserId, debitReq.WalletId)
	if err != nil {
		return nil, err
	}

	amountToAdd := debitReq.Amount.Mul(decimal.NewFromFloat(100))

	walletFunds := decimal.NewFromInt(int64(wallet.Funds)) // should already be in 100s

	newBalance := walletFunds.Sub(amountToAdd)

	if newBalance.IsNegative() {
		err = errors.New(pkg.NotEnoughFunds)
		w.logger.Error(
			"attempted to debit with insufficient funds",
			zap.Error(err),
			zap.Any("debitReq", debitReq),
		)
		return nil, err
	}

	wallet.Funds, _ = strconv.Atoi(newBalance.String())

	if err = w.updateUserWalletByID(wallet); err != nil {
		return nil, err
	}

	err = w.Cache.Set(context.TODO(), fmt.Sprintf("%d-balance", debitReq.UserId), newBalance.String(), time.Duration(w.settings.RedisCacheTimeout*int(time.Minute))).Err()
	if err != nil {
		_ = w.Cache.Del(context.TODO(), fmt.Sprintf("%d-balance", debitReq.UserId))
		w.logger.Error(
			"something went wrong updating the cached data, deleting",
			zap.Error(err),
			zap.Any("wallet", wallet),
		)
	}

	return &api.DebitResponse{
		UserID:   debitReq.UserId,
		WalletID: debitReq.WalletId,
		Balance:  newBalance.Div(decimal.NewFromFloat(100)),
	}, nil
}
