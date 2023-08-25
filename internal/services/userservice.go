package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"wallet-api/internal/api"
	"wallet-api/internal/pkg"
)

type UserService struct {
	DBConn      *gorm.DB
	RedisClient *redis.Client
	logger      *zap.Logger
	settings    UserServiceSettings
}

// UserServiceSettings used to affect code flow
type UserServiceSettings struct {
	Port      int
	Hostname  string
	JWTSecret string
}

type UserServices interface {
	InsertUser(user *api.User) (*api.User, error)
	GetUsers() ([]api.User, error)
	GetUserByUsername(username string) (*pkg.User, error)
	GetUserByID(uID uint) (*api.User, error)
	Login(request api.LoginRequest) (*api.LoginResponse, error)
}

func NewUserService(dbConn *gorm.DB, rc *redis.Client, logger *zap.Logger, settings UserServiceSettings) *UserService {
	return &UserService{
		DBConn:      dbConn,
		RedisClient: rc,
		logger:      logger,
		settings:    settings,
	}
}

// InsertUser inserts new user in users table from data passed in arg.
func (service *UserService) InsertUser(req *api.User) (*api.User, error) {

	users, err := service.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Email == req.Email {
			return nil, errors.New("email already registered")
		}
		if user.Username == req.Username {
			return nil, errors.New("username taken")
		}
	}

	// TODO: hash and salt password

	user := &pkg.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Email:     req.Email,
		Age:       req.Age,
		Password:  req.Password,
	}

	res := service.DBConn.
		Select("first_name", "last_name", "email", "age", "username", "password").
		Create(user)
	if res.Error != nil {
		service.logger.Error("something went wrong inserting user", zap.Any("user", user), zap.Error(res.Error))
		return nil, res.Error
	}

	service.logger.Debug("rows inserted", zap.Int64("rowsAffected", res.RowsAffected))

	return req, nil
}

// GetUsers returns list of users in db.
func (service *UserService) GetUsers() ([]api.User, error) {

	var users []api.User

	// Get all records
	res := service.DBConn.
		Select("first_name", "last_name", "email", "age", "username", "created_at", "updated_at", "id").
		Find(&users)
	if res.Error != nil {
		service.logger.Error("something went wrong getting all players", zap.Error(res.Error))
		return nil, res.Error
	}

	service.logger.Debug("users grabbed", zap.Int64("number", res.RowsAffected))

	return users, nil
}

// GetUserByUsername attempts to retrieve a single row from the users table.
func (service *UserService) GetUserByUsername(username string) (*pkg.User, error) {

	var user pkg.User
	// Get all records
	res := service.DBConn.
		Select("id", "first_name", "last_name", "email", "age", "username", "password", "created_at", "updated_at", "last_login_time_stamp").
		Where("username = ?", username).
		First(&user)
	if res.Error != nil {
		service.logger.Error("something went wrong getting player by username", zap.Error(res.Error), zap.String("username", username))
		return nil, res.Error
	}

	service.logger.Debug("user grabbed", zap.Any("user", user))

	return &user, nil
}

// GetUserByID grabs from table by id
func (service *UserService) GetUserByID(uID uint) (*api.User, error) {

	var user api.User
	// Get all records
	res := service.DBConn.
		Select("first_name", "last_name", "email", "age", "username", "password", "last_login_time_stamp").
		Where("id = ?", uID).
		First(&user)
	if res.Error != nil {
		service.logger.Error("something went wrong getting player by ID", zap.Error(res.Error))
		return nil, res.Error
	}

	service.logger.Debug("user grabbed", zap.Any("user", user))

	return &user, nil
}

// Login is a wrapper for the GetUserByUsername that also validates the password
func (service *UserService) Login(request api.LoginRequest) (*api.LoginResponse, error) {

	user, err := service.GetUserByUsername(request.Username)
	if err != nil {
		return nil, err
	}

	if user.Password != request.Password {
		return nil, fmt.Errorf("invalid passord for user")
	}

	// save userID in jwt token for requests
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(service.settings.JWTSecret))
	if err != nil {
		service.logger.Error("failed to create token", zap.Error(err))
		return nil, err
	}

	unixCT := service.DBConn.NowFunc()

	// update record with login timestamp
	res := service.DBConn.
		Table("users").
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"last_login_time_stamp": unixCT,
			"updated_at":            unixCT,
		})
	if res.Error != nil {
		service.logger.Error("something went wrong updating a player", zap.Error(res.Error))
		return nil, res.Error
	}

	return &api.LoginResponse{Token: tokenString}, nil
}
