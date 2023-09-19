package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"spay/endpoints/api"
	"spay/models"
	"spay/utils"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

const serviceName = "quantech-payment"

// @title Sikem Payment API
// @version 0.0.2
// @description API de paiement.
// @termsOfService http://www.sikem.ci/terms/

// @contact.name API Support
// @contact.url http://www.sikem.ci/support
// @contact.email support@sikem.ci

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host spay.sikem.ci
// @BasePath /

// @securityDefinitions.apikey
// @in header
// @name Authorization
func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	npLog := utils.NewLog(serviceName).With().Logger()

	config, err := models.LoadConfig()
	if err != nil {
		npLog.Fatal().Stack().Err(err).Str("service", serviceName).Msgf("cannot load config %s", serviceName)
	}

	errChan := make(chan error)
	stopChan := make(chan os.Signal, 1)

	// bind OS events to the signal channel
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	// HTTP
	go func(port string) {
		config, err := models.LoadConfig()
		if err == nil {
			if config.DBInit {
				db, err := models.GetDB()
				if err == nil {
					// models.DropTable(db)
					models.CreateUpdateTable(db)
				}
			}
		}

		// Echo instance
		server := echo.New()

		// Echo Banner
		server.HideBanner = true

		// Middleware
		server.Use(middleware.Recover())
		server.Use(middleware.CORS())
		server.Use(middleware.Logger())
		// server.Pre(middleware.AddTrailingSlash())

		server.Validator = models.NewCustomValidator()
		//server.Binder = &models.CustomBinder{}

		// Web site
		// server.Static("/", "./assets/web")

		api.AttachAPI(server, &npLog)

		npLog.Info().Msgf("Create HTTP server in port%v", port)
		go func() {
			if err := server.Start(port); err != nil {
				npLog.Error().Err(err).Msg("shutting down the server")
				server.Logger.Info("shutting down the server")
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), models.ConnectTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Logger.Fatal(err)
		}
	}(fmt.Sprintf(":%v", config.PortAPI))

	select {
	case err := <-errChan:
		npLog.Printf("Fatal error: %v\n", err)
	case <-stopChan:
	}
}
