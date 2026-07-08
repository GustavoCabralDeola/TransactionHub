package api

import (
	"net/http"
	"transactionhub/internal/application/commands"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	transferHandler *commands.TransferHandler
}

func NewTransactionController(th *commands.TransferHandler) *TransactionController {
	return &TransactionController{
		transferHandler: th,
	}
}

type TransactionRequest struct {
	AccountID            string `json:"accountid" binding:"required"`
	DestinationAccountID string `json:"destinationAccountId"`
	Operation            string `json:"operation" binding:"required"`
	Amount               int64  `json:"amount" binding:"required,gt=0"`
	Currency             string `json:"currency" binding:"required"`
	ReferenceID          string `json:"referenceid" binding:"required"`
}

func (tc *TransactionController) ProcessTransaction(c *gin.Context) {

	var request TransactionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Payload or mandatory fields missing",
			"details": err.Error(),
		})
		return
	}

	switch request.Operation {
	case "transfer":

		if request.DestinationAccountID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "destinationAccountId is required for transfers"})
			return
		}

		cmd := commands.TransferCommand{
			SourceAccountID:      request.AccountID,
			DestinationAccountID: request.DestinationAccountID,
			Amount:               request.Amount,
			Currency:             request.Currency,
			ReferenceID:          request.ReferenceID,
		}

		tx, err := tc.transferHandler.Execute(c.Request.Context(), cmd)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "transaction": tx})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Transfer processed successfully!", "transaction": tx})

	default:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid Operation"})
	}
}
