package api

import (
	"spay/endpoints/api/shops"
	"spay/endpoints/api/users"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func AttachAPI(server *echo.Echo, log *zerolog.Logger) {
	apiServer := server.Group("/api")
	{
		// Users Endpoints: /api/users
		users.AttachAPI(apiServer)

		// Shops Endpoints: /api/shops
		shops.AttachAPI(apiServer)

		// Shop Permissions Endpoints: /api/shop-permissions
		// Providers Endpoints: /api/providers
		// Transactions Endpoints: /api/transactions
	}
}
