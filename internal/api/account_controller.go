package api

import (
	"net/http"
	"time"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/infrastructure/repository"

	"github.com/gin-gonic/gin"
)

type AccountController struct {
	repo *repository.GormAccountRepository
}

func NewAccountController(repo *repository.GormAccountRepository) *AccountController {
	return &AccountController{repo: repo}
}

type CreateAccountRequest struct {
	ID              string `json:"id" binding:"required"`
	ClientID        string `json:"client_id" binding:"required"`
	Balance         int64  `json:"balance"`
	ReservedBalance int64  `json:"reserved_balance"`
	CreditLimit     int64  `json:"credit_limit"`
	Status          string `json:"status" binding:"required"`
}

func (ac *AccountController) CreateAccount(c *gin.Context) {
	var request CreateAccountRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Payload", "details": err.Error()})
		return
	}

	newAccount := &account.Account{
		ID:              request.ID,
		ClientID:        request.ClientID,
		Balance:         request.Balance,
		ReservedBalance: request.ReservedBalance,
		CreditLimit:     request.CreditLimit,
		Status:          account.Status(request.Status),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := ac.repo.Save(c.Request.Context(), newAccount); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save in the database", "details": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account succesful created!",
		"account": newAccount,
	})

}
