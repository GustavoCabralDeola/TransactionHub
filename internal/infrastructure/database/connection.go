package database

import (
	"fmt"
	"log"
	"os"
	"transactionhub/internal/infrastructure/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection() *gorm.DB {
	dsn := buildDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	err = db.AutoMigrate(&repository.AccountModel{}, &repository.TransactionModel{})
	if err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}

	return db
}

// Monta a string de conexão do PostgreSQL a partir de variáveis de ambiente
func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "transactionhub")
	password := getEnv("DB_PASSWORD", "transactionhub")
	dbname := getEnv("DB_NAME", "transactionhub")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, password, dbname,
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
