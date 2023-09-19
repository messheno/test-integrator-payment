package services

import (
	"spay/endpoints/api/middlewares"
	"spay/models"

	"github.com/labstack/echo/v4"
)

type ServiceApiRessource struct {
	*models.ServiceModel
}

func AttachAPI(server *echo.Group) {
	serviceApi := &ServiceApiRessource{&models.ServiceModel{}}

	// Service
	serviceApiService := server.Group("/services")
	{
		// Fetch
		serviceApiService.GET("/", serviceApi.Fetch(), middlewares.GrantMid(), models.PaginationMid())

		// Add Service
		serviceApiService.POST("/", serviceApi.Add(), middlewares.GrantMid())

		serviceOneApiService := serviceApiService.Group("/:id", middlewares.GrantMid(), serviceApi.GetOnMid())
		{
			// Get Service Info
			serviceOneApiService.GET("/", serviceApi.GetInfo())

			// Get Client Service Info
			serviceOneApiService.GET("/show-client", serviceApi.GetClient())

			// Regenerate Client Service Info
			serviceOneApiService.POST("/regenerate-client", serviceApi.GenClient())

			// Update Service Info
			serviceOneApiService.PUT("/", serviceApi.UpdateInfo())

			// Delete Service Info
			serviceOneApiService.DELETE("/", serviceApi.Delete())

			servicePermissionsApiService := serviceOneApiService.Group("/permissions")
			{
				// Fetch service permissions
				servicePermissionsApiService.GET("/", serviceApi.FetchPermission(), models.PaginationMid())

				// Add user to service
				servicePermissionsApiService.POST("/add", serviceApi.AddUserToService())
			}
		}
	}
}
