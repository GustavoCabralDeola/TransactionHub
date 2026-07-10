package boostrap

import (
	"log"
	"transactionhub/internal/api"
	"transactionhub/internal/api/routes"
	"transactionhub/internal/application/commands"
	"transactionhub/internal/infrastructure/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Application struct {
	Router *gin.Engine
}

func Build(db *gorm.DB) *Application {
	accountRepo := repository.NewGormAccountRepository(db)
	transactionRepo := repository.NewGormTransactionRepository(db)

	transferHandler := commands.NewTransferHandler(accountRepo, transactionRepo)
	reversalHandler := commands.NewReversalHandler(accountRepo, transactionRepo)
	creditHandler := commands.NewCreditHandler(accountRepo, transactionRepo)
	debitHandler := commands.NewDebitHandler(accountRepo, transactionRepo)
	reserveHandler := commands.NewReserveHandler(accountRepo, transactionRepo)
	captureHandler := commands.NewCaptureHandler(accountRepo, transactionRepo)

	accountController := api.NewAccountController(accountRepo)
	transactionController := api.NewTransactionController(
		accountRepo,
		transferHandler,
		reversalHandler,
		creditHandler,
		debitHandler,
		reserveHandler,
		captureHandler,
	)

	router := routes.SetupRouter(accountController, transactionController)

	return &Application{
		Router: router,
	}
}

func (app *Application) Start(port string) {
	log.Printf("Server started on port %s...", port)
	if err := app.Router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
