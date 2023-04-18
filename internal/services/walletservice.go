package services

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"wallet-api/internal/api"
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
	// fmt.Printf("user id: %d and wallet id: %d \n", userID, walletID)
	return nil, nil
}

func (w *WalletService) Credit(creditReq *api.CreditRequest) (*api.CreditResponse, error) {
	return nil, nil
}

func (w *WalletService) Debit(debitReq *api.DebitRequest) (*api.DebitResponse, error) {
	return nil, nil
}
