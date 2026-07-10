package api

import (
	"context"
	"net/http"
	"transactionhub/internal/application/commands"
	"transactionhub/internal/application/dto"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	accountRepo     account.Repository
	transferHandler *commands.TransferHandler
	reversalHandler *commands.ReversalHandler
	creditHandler   *commands.CreditHandler
	debitHandler    *commands.DebitHandler
	reserveHandler  *commands.ReserveHandler
	captureHandler  *commands.CaptureHandler
}

func NewTransactionController(
	ar account.Repository,
	th *commands.TransferHandler,
	rh *commands.ReversalHandler,
	ch *commands.CreditHandler,
	dh *commands.DebitHandler,
	resh *commands.ReserveHandler,
	caph *commands.CaptureHandler,
) *TransactionController {
	return &TransactionController{
		accountRepo:     ar,
		transferHandler: th,
		reversalHandler: rh,
		creditHandler:   ch,
		debitHandler:    dh,
		reserveHandler:  resh,
		captureHandler:  caph,
	}
}

type TransactionRequest struct {
	AccountID             string `json:"account_id" binding:"required"`
	DestinationAccountID  string `json:"destination_account_id"`
	OriginalTransactionID string `json:"original_transaction_id"`
	Operation             string `json:"operation" binding:"required"`
	Amount                int64  `json:"amount" binding:"required,gt=0"`
	Currency              string `json:"currency" binding:"required"`
	ReferenceID           string `json:"reference_id" binding:"required"`
}

func (tc *TransactionController) ProcessTransaction(c *gin.Context) {
	var req TransactionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload or mandatory fields missing", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()

	switch req.Operation {
	case "credit":
		cmd := commands.CreditCommand{
			AccountID:   req.AccountID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			ReferenceID: req.ReferenceID,
		}
		tx, err := tc.creditHandler.Execute(ctx, cmd)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "transaction": tc.buildResponse(ctx, tx)})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Credit processed successfully!", "transaction": tc.buildResponse(ctx, tx)})

	case "debit":
		cmd := commands.DebitCommand{
			AccountID:   req.AccountID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			ReferenceID: req.ReferenceID,
		}
		tx, err := tc.debitHandler.Execute(ctx, cmd)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "transaction": tc.buildResponse(ctx, tx)})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Debit processed successfully!", "transaction": tc.buildResponse(ctx, tx)})

	case "reserve":
		cmd := commands.ReserveCommand{
			AccountID:   req.AccountID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			ReferenceID: req.ReferenceID,
		}
		tx, err := tc.reserveHandler.Execute(ctx, cmd)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "transaction": tc.buildResponse(ctx, tx)})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Reserve processed successfully!", "transaction": tc.buildResponse(ctx, tx)})

	case "capture":
		cmd := commands.CaptureCommand{
			AccountID:   req.AccountID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			ReferenceID: req.ReferenceID,
		}
		tx, err := tc.captureHandler.Execute(ctx, cmd)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "transaction": tc.buildResponse(ctx, tx)})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Capture processed successfully!", "transaction": tc.buildResponse(ctx, tx)})

	case "transfer":
		if req.DestinationAccountID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "destination_account_id is required for transfers"})
			return
		}
		cmd := commands.TransferCommand{
			SourceAccountID:      req.AccountID,
			DestinationAccountID: req.DestinationAccountID,
			Amount:               req.Amount,
			Currency:             req.Currency,
			ReferenceID:          req.ReferenceID,
		}
		tx, err := tc.transferHandler.Execute(ctx, cmd)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "transaction": tc.buildResponse(ctx, tx)})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Transfer processed successfully!", "transaction": tc.buildResponse(ctx, tx)})

	case "reversal":
		if req.OriginalTransactionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "original_transaction_id is required for reversals"})
			return
		}
		cmd := commands.ReversalCommand{
			OriginalTransactionID: req.OriginalTransactionID,
			ReferenceID:           req.ReferenceID,
		}
		tx, err := tc.reversalHandler.Execute(ctx, cmd)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "transaction": tc.buildResponse(ctx, tx)})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Reversal processed successfully!", "transaction": tc.buildResponse(ctx, tx)})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation. Valid options: credit, debit, reserve, capture, transfer, reversal"})
	}
}

// Monta o DTO de resposta buscando o saldo atualizado da conta
func (tc *TransactionController) buildResponse(ctx context.Context, tx *transaction.Transaction) *dto.TransactionResponse {
	if tx == nil {
		return nil
	}

	resp := &dto.TransactionResponse{
		TransactionID: tx.ID,
		AccountID:     tx.AccountID,
		Operation:     string(tx.Operation),
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		ReferenceID:   tx.ReferenceID,
		Status:        string(tx.Status),
		ErrorMessage:  tx.ErrorMessage,
		Timestamp:     tx.Timestamp,
	}

	// Busca o saldo atual da conta para incluir na resposta
	if acc, err := tc.accountRepo.FindByID(ctx, tx.AccountID); err == nil {
		resp.Balance = acc.Balance
		resp.ReservedBalance = acc.ReservedBalance
		resp.AvailableBalance = acc.AvailableBalance()
	}

	return resp
}
