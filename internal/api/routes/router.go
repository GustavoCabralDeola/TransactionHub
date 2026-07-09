package routes

import (
	"transactionhub/internal/api"

	"github.com/gin-gonic/gin"
)

func SetupRouter(accountController *api.AccountController, transactionController *api.TransactionController) *gin.Engine {

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "sucess",
			"message": "TransactionHub API source is running",
		})
	})

	router.POST("/accounts", accountController.CreateAccount)
	router.POST("/transactions", transactionController.ProcessTransaction)

	return router
}
