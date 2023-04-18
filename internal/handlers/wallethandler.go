package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"wallet-api/internal/api"
	"wallet-api/internal/middleware"
	"wallet-api/internal/services"
)

type WalletHandler struct {
	WalletService services.WalletServices
	Validator     *validator.Validate
	JwtSecret     string
}

func NewWalletHandler(service *services.WalletService, jwtSecret string) *WalletHandler {
	return &WalletHandler{
		WalletService: service,
		Validator:     validator.New(),
		JwtSecret:     jwtSecret,
	}
}

// WalletRoutes sets up user routes with accompanying methods for processing
func (handler *WalletHandler) WalletRoutes(r *gin.RouterGroup) {

	r.Group("wallet", middleware.RequireAuth(handler.JwtSecret)).
		// GET("", handler.getUserWallets).
		// POST("new", handler.newWallet).
		GET(":walletid/balance", handler.getWalletBalance).
		POST(":walletid/credit", handler.creditWallet).
		POST(":walletid/debit", handler.debitWallet)

	return
}

func (handler *WalletHandler) getWalletBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("no user id", nil, fmt.Errorf("no user id saved from token")))
		return
	}

	walletID := c.Param("walletid")
	if walletID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get walletID from url", nil, fmt.Errorf("missing url")))
		return
	}

	wID, err := strconv.Atoi(walletID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get walletID from url", nil, fmt.Errorf("wrong id format")))
		return
	}

	user, err := handler.WalletService.Balance(userID.(int), wID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to get user by walletID", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("successfully grabbed wallet balance", user, nil))
	return
}

func (handler *WalletHandler) creditWallet(c *gin.Context) {
	walletID := c.Param("walletid")
	if walletID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get walletID from url", nil, fmt.Errorf("missing url")))
		return
	}

	var creditRequest api.CreditRequest

	if err := c.ShouldBindJSON(&creditRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get json body", nil, err))
		return
	}

	creditRequest.WalletId, _ = strconv.Atoi(walletID)
	uid, _ := c.Get("user_id")

	creditRequest.UserId = uid.(int)

	user, err := handler.WalletService.Credit(&creditRequest)
	if err != nil && err == gorm.ErrRecordNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, api.GenerateMessageResponse("failed to credit wallet", nil, err))
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to credit wallet", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("credit wallet successful", user, nil))
	return
}

func (handler *WalletHandler) debitWallet(c *gin.Context) {
	walletID := c.Param("walletid")
	if walletID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to get walletID from url", nil, fmt.Errorf("missing url")))
		return
	}

	var debitRequest api.DebitRequest
	if err := c.ShouldBindJSON(&debitRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.GenerateMessageResponse("failed to parse json body", nil, err))
		return
	}

	debitRequest.WalletId, _ = strconv.Atoi(walletID)
	uid, _ := c.Get("user_id")

	debitRequest.UserId = uid.(int)
	user, err := handler.WalletService.Debit(&debitRequest)
	if err != nil && err == gorm.ErrRecordNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, api.GenerateMessageResponse("failed to debit wallet", nil, err))
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.GenerateMessageResponse("failed to debit wallet", nil, err))
		return
	}

	c.JSON(http.StatusOK, api.GenerateMessageResponse("debit wallet successful", user, nil))
	return
}
