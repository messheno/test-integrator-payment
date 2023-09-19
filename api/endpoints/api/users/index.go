package users

import (
	"spay/endpoints/api/middlewares"
	"spay/models"

	"github.com/labstack/echo/v4"
)

type UserApiRessource struct {
	*models.UserModel
}

func AttachAPI(server *echo.Group) {
	userApi := &UserApiRessource{&models.UserModel{}}

	// User
	userApiService := server.Group("/users")
	{
		// Fetch
		userApiService.GET("/", userApi.Fetch(), middlewares.GrantMid(), models.PaginationMid())

		// Add
		userApiService.POST("/", userApi.Add())

		// Change Role
		userApiService.POST("/change-role", userApi.ChangeRole(), middlewares.GrantMid())

		// Login
		userApiService.POST("/login", userApi.Login())

		// Get Info
		userApiService.GET("/:id", userApi.GetInfo(), middlewares.GrantMid(), userApi.GetOnMid())

		// Update User Info
		userApiService.PUT("/:id", userApi.UpdateInfo(), middlewares.GrantMid(), userApi.GetOnMid())
	}
}
