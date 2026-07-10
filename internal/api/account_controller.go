package api

import (
	"net/http"
	"time"
	"transactionhub/internal/api/dto"
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

// Representa o payload para criação de conta
type CreateAccountRequest struct {
	ID          string `json:"id" binding:"required" example:"ACC-010"`
	ClientID    string `json:"client_id" binding:"required" example:"CLI-001"`
	Balance     int64  `json:"balance" example:"1000"`
	CreditLimit int64  `json:"credit_limit" example:"50000"`
	Status      string `json:"status" binding:"required" example:"active"`
}

// CreateAccount godoc
//
//	@Summary		Create a new account
//	@Description	Creates a bank account with initial balance, credit limit, and status
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			account	body		CreateAccountRequest	true	"Account data"
//	@Success		201		{object}	map[string]interface{}	"Account created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid payload"
//	@Failure		500		{object}	map[string]interface{}	"Internal error while saving to database"
//	@Router			/accounts [post]
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
		ReservedBalance: 0,
		CreditLimit:     request.CreditLimit,
		Status:          account.Status(request.Status),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := ac.repo.Save(c.Request.Context(), newAccount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save in the database", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully!",
		"account": dto.AccountResponse{
			ID:               newAccount.ID,
			ClientID:         newAccount.ClientID,
			Balance:          newAccount.Balance,
			ReservedBalance:  newAccount.ReservedBalance,
			AvailableBalance: newAccount.AvailableBalance(),
			CreditLimit:      newAccount.CreditLimit,
			Status:           string(newAccount.Status),
		},
	})
}

// GetAccount godoc
//
//	@Summary		Get an account
//	@Description	Returns the full data and current balances of an account by ID
//	@Tags			accounts
//	@Produce		json
//	@Param			id	path		string					true	"Account ID (e.g. ACC-001)"
//	@Success		200	{object}	dto.AccountResponse		"Account data"
//	@Failure		404	{object}	map[string]interface{}	"Account not found"
//	@Router			/accounts/{id} [get]
func (ac *AccountController) GetAccount(c *gin.Context) {
	id := c.Param("id")

	acc, err := ac.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, dto.AccountResponse{
		ID:               acc.ID,
		ClientID:         acc.ClientID,
		Balance:          acc.Balance,
		ReservedBalance:  acc.ReservedBalance,
		AvailableBalance: acc.AvailableBalance(),
		CreditLimit:      acc.CreditLimit,
		Status:           string(acc.Status),
	})
}
