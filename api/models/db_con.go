package models

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	defaultLog "log"
)

func GetDB() (*gorm.DB, error) {
	dsn := ""
	var db *gorm.DB

	config, err := LoadConfig()
	if err != nil {
		return db, err
	}

	newLogger := logger.New(
		defaultLog.New(os.Stdout, "\r\n", defaultLog.LstdFlags), // io writer
		logger.Config{
			LogLevel:             logger.Error,
			Colorful:             true,
			ParameterizedQueries: true,
		},
	)

	switch config.DBProvider {
	case "pg":
		dsn = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", config.DBHost, config.DBUser, config.DBPass, config.DBName, config.DBPort)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			TranslateError: true,
			Logger:         newLogger,
		})
	case "mysql":
		dsn = fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			TranslateError: true,
			Logger:         newLogger,
		})
	case "sqlserver":
		dsn = fmt.Sprintf("sqlserver://%v:%v@%v:%v?database=%v", config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{
			TranslateError: true,
			Logger:         newLogger,
		})

	default:
		err = fmt.Errorf("drive inconnu %v", config.DBProvider)
	}

	// if config.DBInit {
	// 	err := CreateUpdateTable(db)
	// 	if err != nil {
	// 		return db, err
	// 	}
	// }

	return db, err
}

func CreateUpdateTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&UserModel{},
		&ShopModel{},
		&ShopPermissionModel{},
		&ProviderModel{},
		&TransactionModel{},
	)
}

func DropTable(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&UserModel{},
		&ShopModel{},
		&ShopPermissionModel{},
		&ProviderModel{},
		&TransactionModel{},
	)
}
