package providers

import (
	"spay/endpoints/api/middlewares"
	"spay/models"

	"github.com/labstack/echo/v4"
)

type ProviderApiRessource struct {
	*models.ProviderModel
}

func AttachAPI(server *echo.Group) {
	providerApi := &ProviderApiRessource{&models.ProviderModel{}}

	// Transaction
	serviceApiService := server.Group("/providers")
	{
		// Fetch
		serviceApiService.GET("/", providerApi.Fetch(), middlewares.GrantMid(), models.PaginationMid())

		// Add Provider
		serviceApiService.POST("/", providerApi.Add(), middlewares.GrantMid())

		// serviceOneApiService := serviceApiService.Group("/:id", middlewares.GrantMid(), transactionApi.GetOnMid())
		// {
		// 	// Get Transaction Info
		// 	serviceOneApiService.GET("/", transactionApi.GetInfo())

		// 	// Delete Transaction Info
		// 	serviceOneApiService.DELETE("/", transactionApi.Delete())
		// }
	}
}
