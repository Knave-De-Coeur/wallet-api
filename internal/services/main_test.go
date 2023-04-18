package services

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"wallet-api/internal/config"
	"wallet-api/internal/utils"

	"gorm.io/driver/mysql"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mockDB  *sql.DB
	sqlMock sqlmock.Sqlmock

	gormDB *gorm.DB
	log    *zap.Logger

	err error

	userService   *UserService
	walletService *WalletService

	redisClientMock redismock.ClientMock

	rd *redis.Client
)

func TestMain(m *testing.M) {
	log, err = utils.SetUpLogger()
	if err != nil {
		panic(err)
	}

	defer log.Sync()

	mockDB, sqlMock, err = sqlmock.New()
	if err != nil {
		panic(fmt.Sprintf("an error '%s' was not expected when opening a stub database connection", err))
	}
	defer mockDB.Close()

	dialector := mysql.New(mysql.Config{
		DSN:                       "sqlmock_db_0",
		DriverName:                "mysql",
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		NowFunc: func() time.Time {
			return time.Unix(time.Now().Unix(), 0).UTC()
		},
	})
	if err != nil {
		panic(err)
	}

	rd, redisClientMock = redismock.NewClientMock()

	userService = NewUserService(gormDB, rd, log, UserServiceSettings{
		Port:      0,
		Hostname:  config.WalletConfigs.Host,
		JWTSecret: config.WalletConfigs.JWTSecret,
	})
	walletService = NewWalletService(gormDB, rd, log, WalletServiceSettings{
		Port:              0,
		Hostname:          config.WalletConfigs.Host,
		RedisCacheTimeout: config.WalletConfigs.RedisExpiry,
	}, userService)

	m.Run()
}
