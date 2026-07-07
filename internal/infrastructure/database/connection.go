package database

import (
	"log"
	"transactionhub/internal/infrastructure/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewConnection() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("transactionhub.db"), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	err = db.AutoMigrate(&repository.AccountModel{}, &repository.TransactionModel{})

	if err != nil {
		log.Fatal("Failed to create the tables: ", err)
	}

	return db
}
