package endpoints

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"integrator/models"
)

func UserAttachAPI(server *echo.Group, db *gorm.DB) {
	user := models.UserModel{}

	// User
	userApiService := server.Group("/users")
	{
		// GetAll
		userApiService.GET("/", user.APIGetAll(db), GrantMid(models.USER_MANAGER), models.PaginationMid())

		// Read
		userApiService.GET("/:id", user.APIRead(db), GrantMid(models.USER_MANAGER), user.GetOnMid(db))

		// Create
		userApiService.POST("/", user.APICreate(db), GrantMid(models.USER_MANAGER))

		// Update
		userApiService.PUT("/:id", user.APIUpdate(db), GrantMid(models.USER_MANAGER), user.GetOnMid(db))

		// Delete
		userApiService.DELETE("/:id", user.APIDelete(db), GrantMid(models.USER_MANAGER), user.GetOnMid(db))

		// Login
		userApiService.POST("/login", user.APILogin(db))
	}
}
