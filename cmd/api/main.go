package main

import (
	"log"
	"transactionhub/internal/api"
	"transactionhub/internal/application/commands"
	"transactionhub/internal/infrastructure/database"
	"transactionhub/internal/infrastructure/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	db := database.NewConnection()
	log.Println("Database connected and tables verified successfully")

	accountRepo := repository.NewGormAccountRepository(db)
	transactionRepo := repository.NewGormTransactionRepository(db)

	transferHandler := commands.NewTransferHandler(accountRepo, transactionRepo)

	accountController := api.NewAccountController(accountRepo)
	transactionController := api.NewTransactionController(transferHandler)

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "sucess",
			"message": "TransactionHub API source is running",
		})
	})

	router.POST("/accounts", accountController.CreateAccount)
	router.POST("/transactions", transactionController.ProcessTransaction)

	log.Println("Server started on port :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
