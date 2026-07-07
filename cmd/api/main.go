package main

import (
	"log"
	"transactionhub/internal/infrastructure/database"

	"github.com/gin-gonic/gin"
)

func main() {
	_ = database.NewConnection()
	log.Println("Database connected and tables created")

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "sucess",
			"message": "TransactionHub source is running",
		})
	})

	log.Println("Server started on port :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
