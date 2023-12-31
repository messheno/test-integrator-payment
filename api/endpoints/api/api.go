package api

import (
	"spay/endpoints/api/providers"
	"spay/endpoints/api/services"
	"spay/endpoints/api/transactions"
	"spay/endpoints/api/users"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func AttachAPI(server *echo.Echo, log *zerolog.Logger) {
	apiServer := server.Group("/api")
	{
		// Users Endpoints: /api/users
		users.AttachAPI(apiServer)

		// Services Endpoints: /api/services
		services.AttachAPI(apiServer)

		// Transaction Endpoints: /api/providers
		providers.AttachAPI(apiServer)

		// Transaction Endpoints: /api/transactions
		transactions.AttachAPI(apiServer)
	}
}
