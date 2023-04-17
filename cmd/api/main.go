package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"wallet-api/internal/config"
	"wallet-api/internal/utils"
)

func main() {
	logger, err := utils.SetUpLogger()
	if err != nil {
		log.Fatalf("somethign went wrong setting up logger for api: %+v", err)
	}

	defer func(logger *zap.Logger) {
		_ = logger.Sync()
		// if err != nil {
		// 	fmt.Printf("something went wrong deferring the close to the logger: %v", err)
		// }
	}(logger)

	logger.Info("🚀 connecting to db")

	quizDBConn, err := utils.SetUpDBConnection(
		config.CurrentConfigs.DBUser,
		config.CurrentConfigs.DBPassword,
		config.CurrentConfigs.Host,
		config.CurrentConfigs.DBName,
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

	routes, err := setUpRoutes(quizDBConn, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if err = routes.Run(); err != nil {
		logger.Fatal("something went wrong setting up router")
	}
}

// setUpRoutes adds routes and returns gin engine
func setUpRoutes(quizDBConn *gorm.DB, logger *zap.Logger) (*gin.Engine, error) {

	_, err := strconv.Atoi(config.CurrentConfigs.Port)
	if err != nil {
		logger.Error(fmt.Sprintf("port config not int %d", err))
		return nil, err
	}

	// userService := services.NewUserService(quizDBConn, nc, logger, services.UserServiceSettings{
	// 	Port:     portNum,
	// 	Hostname: config.CurrentConfigs.Host,
	// })

	r := gin.New()

	r.Use(gin.Logger())

	// r.Use(gin.Middleware)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// handlers.NewUserHandler(userService, nc).UserRoutes(r.Group("/"))

	return r, nil
}
