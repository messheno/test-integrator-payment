package main

import (
	"context"
	"fmt"
	"integrator/endpoints"
	"integrator/models"
	"integrator/utils"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	_ "integrator/docs" // Document Swagger

	echoSwagger "github.com/swaggo/echo-swagger"
)

const serviceName = "integrator-core"

// @title Integrator Core API
// @version 0.0.1
// @description API Core.
// @termsOfService https://www.integrator.com/terms

// @contact.name API Support
// @contact.url https://www.Integrator.com/support
// @contact.email support@Integrator.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host core-api.Integrator.com
// @BasePath /

// @securityDefinitions.apikey
// @in header
// @name Authorization
func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	coreLog := utils.NewLog(serviceName).With().Logger()

	config, err := models.LoadConfig()
	if err != nil {
		coreLog.Fatal().Stack().Err(err).Str("service", serviceName).Msgf("cannot load config %s", serviceName)
	}

	errChan := make(chan error)
	stopChan := make(chan os.Signal, 1)

	// bind OS events to the signal channel
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	// HTTP
	go func(port string) {
		// Test de connexion
		dsn := ""
		var db *gorm.DB
		var err error
		debug := true

		// Chargement du fichier de configuration crypté
		config, err := models.LoadConfig()
		if err != nil {
			log.Fatal("Impossible de charger le fichier de configuration")
			return
		}

		switch config.DBType {
		case int32(models.DB_TYPE_POSTGRESQL.EnumIndex()):
			dsn = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", config.DBHost, config.DBUser, config.DBPass, config.DBName, config.DBPort)
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: debug,
			})
		case int32(models.DB_TYPE_MYSQL.EnumIndex()):
			dsn = fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: debug,
			})
		case int32(models.DB_TYPE_SQLSERVER.EnumIndex()):
			dsn = fmt.Sprintf("sqlserver://%v:%v@%v:%v?database=%v", config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
			db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: debug,
			})
		case int32(models.DB_TYPE_SQLITE.EnumIndex()):
			dsn = config.DBPath

			os.MkdirAll(path.Dir(dsn), os.ModePerm)

			db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: debug,
			})

		default:
			err = fmt.Errorf("drive inconnu %d", config.DBType)
		}

		if err == nil {
			// Migrate the schema
			err = models.CreateUpdateTable(db)
			if err != nil {
				coreLog.Error().Err(err).Msg(err.Error())
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

		// Documentation
		server.GET("/docs/*", echoSwagger.WrapHandler)

		// Web site
		// server.Static("/", "./assets/web")

		endpoints.AttachAPI(server, db)

		coreLog.Info().Msgf("Create HTTP server in port%v", port)
		go func() {
			if err := server.Start(port); err != nil {
				coreLog.Error().Err(err).Msg("shutting down the server")
				server.Logger.Info("shutting down the server")
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), models.CONNECTION_TIMEOUT)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Logger.Fatal(err)
		}
	}(fmt.Sprintf(":%v", config.PortAPI))

	select {
	case err := <-errChan:
		coreLog.Printf("Fatal error: %v\n", err)
	case <-stopChan:
	}
}
