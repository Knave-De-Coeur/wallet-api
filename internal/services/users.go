package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"wallet-api/internal/api"
	"wallet-api/internal/pkg"
)

type UserService struct {
	DBConn   *gorm.DB
	logger   *zap.Logger
	settings UserServiceSettings
}

// UserServiceSettings used to affect code flow
type UserServiceSettings struct {
	Port     int
	Hostname string
}

type UserServices interface {
	InsertUser(user *api.User) (*api.User, error)
	GetUsers() ([]api.User, error)
	GetUserByUsername(username string) (*pkg.User, error)
	GetUserByID(uID uint) (*api.User, error)
	Login(request api.LoginRequest) (*api.User, error)
}

func NewUserService(dbConn *gorm.DB, logger *zap.Logger, settings UserServiceSettings) *UserService {
	return &UserService{
		DBConn:   dbConn,
		logger:   logger,
		settings: settings,
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
func (service *UserService) Login(request api.LoginRequest) (*api.User, error) {

	user, err := service.GetUserByUsername(request.Username)
	if err != nil {
		return nil, err
	}

	if user.Password != request.Password {
		return nil, fmt.Errorf("invalid passord for user")
	}

	unixCT := service.DBConn.NowFunc()

	fieldsToUpdate := map[string]interface{}{"last_login_time_stamp": unixCT, "updated_at": unixCT}

	// update record with login timestamp
	res := service.DBConn.
		Table("users").
		Where("id = ?", user.ID).
		Updates(fieldsToUpdate)
	if res.Error != nil {
		service.logger.Error("something went wrong updating a player", zap.Error(res.Error))
		return nil, res.Error
	}

	// TODO: generate jwt

	return &api.User{
		ID:                 strconv.Itoa(int(user.ID)),
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Email:              user.Email,
		Username:           user.Username,
		Age:                user.Age,
		CreatedAT:          user.CreatedAt.Format(time.RFC3339),
		UpdatedAT:          user.UpdatedAt.Format(time.RFC3339),
		LastLoginTimeStamp: user.LastLoginTimeStamp.Time.Format(time.RFC3339),
	}, nil
}
