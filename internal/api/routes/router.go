package routes

import (
	"transactionhub/internal/api"

	_ "transactionhub/docs" // importa os docs gerados pelo swag

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(accountController *api.AccountController, transactionController *api.TransactionController) *gin.Engine {

	router := gin.Default()

	// Health check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "sucess",
			"message": "TransactionHub API source is running",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/accounts", accountController.CreateAccount)
	router.GET("/accounts/:id", accountController.GetAccount)

	router.POST("/transactions", transactionController.ProcessTransaction)

	return router
}
