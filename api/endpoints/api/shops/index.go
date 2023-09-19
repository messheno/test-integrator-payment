package shops

import (
	"spay/endpoints/api/middlewares"
	"spay/models"

	"github.com/labstack/echo/v4"
)

type ShopApiRessource struct {
	*models.ShopModel
}

func AttachAPI(server *echo.Group) {
	shopApi := &ShopApiRessource{&models.ShopModel{}}

	// User
	shopApiService := server.Group("/shops")
	{
		// Fetch
		shopApiService.GET("/", shopApi.Fetch(), middlewares.GrantMid(), models.PaginationMid())

		// Add Shop
		shopApiService.POST("/", shopApi.Add(), middlewares.GrantMid())

		shopOneApiService := shopApiService.Group("/:id", middlewares.GrantMid(), shopApi.GetOnMid())
		{
			// Get Shop Info
			shopOneApiService.GET("/", shopApi.GetInfo())

			// Get Client Shop Info
			shopOneApiService.GET("/show-client", shopApi.GetClient())

			// Regenerate Client Shop Info
			shopOneApiService.POST("/regenerate-client", shopApi.GenClient())

			// Update Shop Info
			shopOneApiService.PUT("/", shopApi.UpdateInfo())

			// Delete Shop Info
			shopOneApiService.DELETE("/", shopApi.Delete())

			shopPermissionsApiService := shopOneApiService.Group("/permissions")
			{
				// Fetch shop permissions
				shopPermissionsApiService.GET("/", shopApi.FetchPermission(), models.PaginationMid())

				// Add user to shop
				shopPermissionsApiService.POST("/add", shopApi.AddUserToShop())
			}
		}
	}
}
