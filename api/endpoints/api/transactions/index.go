package transactions

import (
	"spay/endpoints/api/middlewares"
	"spay/models"

	"github.com/labstack/echo/v4"
)

type TransactionApiRessource struct {
	*models.TransactionModel
}

func AttachAPI(server *echo.Group) {
	transactionApi := &TransactionApiRessource{&models.TransactionModel{}}

	// Transaction
	serviceApiService := server.Group("/transactions")
	{
		// Fetch
		serviceApiService.GET("/", transactionApi.Fetch(), middlewares.GrantMid(), models.PaginationMid())

		// Add Transaction
		serviceApiService.POST("/", transactionApi.Add(), middlewares.GrantMid())

		// serviceOneApiService := serviceApiService.Group("/:id", middlewares.GrantMid(), transactionApi.GetOnMid())
		// {
		// 	// Get Transaction Info
		// 	serviceOneApiService.GET("/", transactionApi.GetInfo())

		// 	// Delete Transaction Info
		// 	serviceOneApiService.DELETE("/", transactionApi.Delete())
		// }
	}
}
