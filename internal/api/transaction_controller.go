package api

import (
	"context"
	"net/http"
	"transactionhub/internal/application/commands"
	"transactionhub/internal/application/dto"
	"transactionhub/internal/domain/account"
	"transactionhub/internal/domain/transaction"
	"transactionhub/internal/infrastructure/event"

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
	publisher       event.EventPublisher
}

func NewTransactionController(
	ar account.Repository,
	th *commands.TransferHandler,
	rh *commands.ReversalHandler,
	ch *commands.CreditHandler,
	dh *commands.DebitHandler,
	resh *commands.ReserveHandler,
	caph *commands.CaptureHandler,
	pub event.EventPublisher,
) *TransactionController {
	return &TransactionController{
		accountRepo:     ar,
		transferHandler: th,
		reversalHandler: rh,
		creditHandler:   ch,
		debitHandler:    dh,
		reserveHandler:  resh,
		captureHandler:  caph,
		publisher:       pub,
	}
}

// Representa o payload de uma operação financeira
type TransactionRequest struct {
	AccountID             string `json:"account_id" binding:"required" example:"ACC-001"`
	DestinationAccountID  string `json:"destination_account_id" example:"ACC-002"`
	OriginalTransactionID string `json:"original_transaction_id" example:"a1b2c3d4-..."`
	Operation             string `json:"operation" binding:"required" example:"credit"`
	Amount                int64  `json:"amount" binding:"required,gt=0" example:"100000"`
	Currency              string `json:"currency" binding:"required" example:"BRL"`
	ReferenceID           string `json:"reference_id" binding:"required" example:"TXN-001"`
}

// ProcessTransaction godoc
//
//	@Summary		Process a financial transaction
//	@Description	Executes a financial operation on an account. Supported operations: credit, debit, reserve, capture, transfer and reversal.
//	@Description
//	@Description	**credit**: Adds funds to the account balance.
//	@Description	**debit**: Removes funds from the balance (uses available balance + credit limit).
//	@Description	**reserve**: Moves funds from available balance to reserved balance.
//	@Description	**capture**: Confirms a reservation, removing it from the reserved balance.
//	@Description	**transfer**: Moves funds between two accounts (requires destination_account_id).
//	@Description	**reversal**: Reverses a previous transaction (requires original_transaction_id).
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			transaction	body		TransactionRequest		true	"Transaction data"
//	@Success		201			{object}	map[string]interface{}	"Transaction processed successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid payload or missing required fields"
//	@Failure		422			{object}	map[string]interface{}	"Transaction could not be processed (e.g. insufficient balance)"
//	@Router			/transactions [post]
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
		tc.publishEvent(tx)
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
		tc.publishEvent(tx)
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
		tc.publishEvent(tx)
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
		tc.publishEvent(tx)
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
		tc.publishEvent(tx)
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
		tc.publishEvent(tx)
		c.JSON(http.StatusCreated, gin.H{"message": "Reversal processed successfully!", "transaction": tc.buildResponse(ctx, tx)})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation. Valid options: credit, debit, reserve, capture, transfer, reversal"})
	}
}

func (tc *TransactionController) publishEvent(tx *transaction.Transaction) {
	if tx != nil && tc.publisher != nil {
		go tc.publisher.Publish(context.Background(), "transaction_events", tx)
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

	if acc, err := tc.accountRepo.FindByID(ctx, tx.AccountID); err == nil {
		resp.Balance = acc.Balance
		resp.ReservedBalance = acc.ReservedBalance
		resp.AvailableBalance = acc.AvailableBalance()
	}

	return resp
}
