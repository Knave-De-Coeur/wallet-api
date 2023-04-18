package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"wallet-api/internal/config"
	"wallet-api/internal/handlers"
	"wallet-api/internal/services"
	"wallet-api/internal/utils"
)

func main() {
	logger, err := utils.SetUpLogger()
	if err != nil {
		log.Fatalf("somethign went wrong setting up logger for api: %+v", err)
	}

	defer logger.Sync()

	logger.Info("🚀 connecting to db")

	quizDBConn, err := utils.SetUpDBConnection(
		config.WalletConfigs.DBUser,
		config.WalletConfigs.DBPassword,
		config.WalletConfigs.Host,
		config.WalletConfigs.DBName,
		logger,
	)
	if err != nil {
		logger.Fatal("exiting application...", zap.Error(err))
	}

	logger.Info(fmt.Sprintf("✅ Setup connection to %s db.", quizDBConn.Migrator().CurrentDatabase()))

	logger.Info("🚀 Running migrations")

	if err = utils.SetUpSchema(quizDBConn, logger); err != nil {
		logger.Fatal(err.Error())
	}

	db, err := quizDBConn.DB()
	if err != nil {
		logger.Fatal("something went wrong getting the database conn from gorm", zap.Error(err))
	}

	if err = utils.RunUpMigrations(db, logger); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info(fmt.Sprintf("✅ Applied migrations to %s db.", quizDBConn.Migrator().CurrentDatabase()))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.WalletConfigs.RedisAddress,
		Password: config.WalletConfigs.RedisPassword,
		DB:       config.WalletConfigs.RedisDB,
	})

	routes, err := setUpRoutes(quizDBConn, redisClient, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if err = routes.Run(); err != nil {
		logger.Fatal("something went wrong setting up router")
	}
}

// setUpRoutes adds routes and returns gin engine
func setUpRoutes(quizDBConn *gorm.DB, rc *redis.Client, logger *zap.Logger) (*gin.Engine, error) {

	portNum, err := strconv.Atoi(config.WalletConfigs.Port)
	if err != nil {
		logger.Error(fmt.Sprintf("port config not int %d", err))
		return nil, err
	}

	userService := services.NewUserService(quizDBConn, rc, logger, services.UserServiceSettings{
		Port:      portNum,
		Hostname:  config.WalletConfigs.Host,
		JWTSecret: config.WalletConfigs.JWTSecret,
	})

	walletService := services.NewWalletService(quizDBConn, rc, logger, services.WalletServiceSettings{
		Port:     portNum,
		Hostname: config.WalletConfigs.Host,
	}, userService)

	r := gin.New()

	r.Use(gin.Logger())

	// r.Use(gin.Middleware)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	handlers.NewUserHandler(userService).UserRoutes(r.Group("/"))
	handlers.NewWalletHandler(walletService, config.WalletConfigs.JWTSecret).WalletRoutes(r.Group("/"))

	return r, nil
}
