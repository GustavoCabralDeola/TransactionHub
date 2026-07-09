package main

import (
	"log"
	"transactionhub/internal/boostrap"
	"transactionhub/internal/infrastructure/database"
)

func main() {
	//Conexão com o banco:
	db := database.NewConnection()
	log.Println("Database connected and tables verified successfully")

	//Builda as instancias para construir o sistema:
	app := boostrap.Build(db)

	//Roda o servidor junto com número da porta:
	app.Start(":8080")
}
