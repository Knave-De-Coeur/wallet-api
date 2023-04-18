package services

import (
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"wallet-api/internal/api"
	"wallet-api/internal/pkg"
)

type WalletService struct {
	DBConn      *gorm.DB
	RedisClient *redis.Client
	logger      *zap.Logger
	settings    WalletServiceSettings
	UserService *UserService
}

// WalletServiceSettings used to affect code flow
type WalletServiceSettings struct {
	Port     int
	Hostname string
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
		RedisClient: rc,
		logger:      logger,
		settings:    settings,
		UserService: userService,
	}
}

func (w *WalletService) Balance(userID, walletID int) (*api.BalanceResponse, error) {

	wallet, err := w.getUserWalletByID(userID, walletID)
	if err != nil {
		return nil, err
	}

	w.logger.Debug("wallet grabbed", zap.Any("wallet", wallet))

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
	res := w.DBConn.Save(wallet)
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

	if creditReq.Amount.IsNegative() {
		return nil, errors.New("amount cannot be negative")
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

	return &api.CreditResponse{
		UserID:   creditReq.UserId,
		WalletID: creditReq.WalletId,
		Balance:  newBalance.Div(decimal.NewFromFloat(100)),
	}, nil
}

func (w *WalletService) Debit(debitReq *api.DebitRequest) (*api.DebitResponse, error) {
	if debitReq.Amount.IsNegative() {
		return nil, errors.New("amount cannot be negative")
	}

	wallet, err := w.getUserWalletByID(debitReq.UserId, debitReq.WalletId)
	if err != nil {
		return nil, err
	}

	amountToAdd := debitReq.Amount.Mul(decimal.NewFromFloat(100))

	walletFunds := decimal.NewFromInt(int64(wallet.Funds)) // should already be in 100s

	newBalance := walletFunds.Sub(amountToAdd)

	if newBalance.IsNegative() {
		return nil, errors.New("not enough balance")
	}

	wallet.Funds, _ = strconv.Atoi(newBalance.String())

	if err = w.updateUserWalletByID(wallet); err != nil {
		return nil, err
	}

	return &api.DebitResponse{
		UserID:   debitReq.UserId,
		WalletID: debitReq.WalletId,
		Balance:  newBalance.Div(decimal.NewFromFloat(100)),
	}, nil
}
