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

type WalletHandler struct {
	WalletService services.WalletServices
	Validator     *validator.Validate
}

func NewWalletHandler(service *services.WalletService) *WalletHandler {
	return &WalletHandler{
		WalletService: service,
		Validator:     validator.New(),
	}
}

// WalletRoutes sets up user routes with accompanying methods for processing
func (handler *WalletHandler) WalletRoutes(r *gin.RouterGroup) {

	r.Group("wallet").
		// GET("", handler.getUserWallets).
		// POST("new", handler.newWallet).
		GET(":walletid/balance", handler.getWalletBalance).
		POST(":walletid/credit", handler.creditWallet).
		POST(":walletid/debit", handler.debitWallet)

	return
}

func (handler *WalletHandler) getWalletBalance(c *gin.Context) {

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

	// todo get user id from redis token
	user, err := handler.WalletService.Balance(1, wID)
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
