package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"wallet-api/internal/api"
	"wallet-api/internal/services"
)

type UserHandler struct {
	UserService services.UserServices
	Validator   *validator.Validate
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		UserService: service,
		Validator:   validator.New(),
	}
}

// UserRoutes sets up user routes with accompanying methods for processing
func (handler *UserHandler) UserRoutes(r *gin.RouterGroup) {

	r.POST("login", handler.login)

	// TODO add validation for correct token and return 401 accordingly
	// r.Group("users").
	// 	GET("", handler.getUsers).
	// 	GET("username/:username", handler.getUserByUsername).
	// 	GET("id/:uID", handler.getUserByID).
	// 	POST("new", handler.newUser)

	return
}

func (handler *UserHandler) getUsers(c *gin.Context) {

	users, err := handler.UserService.GetUsers()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to get users", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("successfully grabbed all users", users, nil))
	return
}

func (handler *UserHandler) getUserByUsername(c *gin.Context) {

	username := c.Param("username")

	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get username from url", nil, fmt.Errorf("missing url")))
		return
	}

	user, err := handler.UserService.GetUserByUsername(username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to get user by username", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("successfully got user", user, nil))
	return
}

func (handler *UserHandler) getUserByID(c *gin.Context) {

	userID := c.Param("uID")
	userIDint, err := strconv.Atoi(userID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get userID", nil, err))
		return
	} else if userIDint < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get user", nil, fmt.Errorf("invalid user id")))
		return
	}

	user, err := handler.UserService.GetUserByID(uint(userIDint))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to get user", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("successfully got user", user, nil))
	return
}

func (handler *UserHandler) newUser(c *gin.Context) {

	var user api.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to parse new user request", nil, err))
		return
	}

	if err := handler.Validator.Struct(user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("missing or incorrect data received", nil, err))
		return
	}

	res, err := handler.UserService.InsertUser(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to add user", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("successfully inserted user", res, nil))
	return
}

// Login endpoint function that checks username and password and sets user appropriately
func (handler *UserHandler) login(c *gin.Context) {
	var loginReq api.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to login", nil, err))
		return
	}

	user, err := handler.UserService.Login(loginReq)
	if err != nil && err == gorm.ErrRecordNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, api.GenerateMessageResponse("failed to login requested user", nil, err))
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to login requested user", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("login successful", user, nil))
	return
}
