package models

import (
	"os"

	"github.com/spf13/viper"
)

// Config strcuture de configuration
type Config struct {
	PortAPI string `mapstructure:"PORT_API"`

	// Keycloak client data
	KeyCloakHost         string `mapstructure:"KEYCLOAK_HOST"`
	KeyCloakClientID     string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	KeyCloakClientSecret string `mapstructure:"KEYCLOAK_CLIENT_SECRET"`
	KeyCloakClientRealm  string `mapstructure:"KEYCLOAK_CLIENT_REALM"`

	// DATABASE
	DBProvider string `mapstructure:"DB_PROVIDER"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_NAME"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPass     string `mapstructure:"DB_PASS"`
	DBInit     bool   `mapstructure:"DB_INIT"`

	// SMTP
	SmtpFrom     string `mapstructure:"SMTP_FROM"`
	SmtpHost     string `mapstructure:"SMTP_HOST"`
	SmtpPort     string `mapstructure:"SMTP_PORT"`
	SmtpStartTls bool   `mapstructure:"SMTP_START_TLS"`
	SmtpUser     string `mapstructure:"SMTP_USER"`
	SmtpPass     string `mapstructure:"SMTP_PASS"`

	// Uploads
	UploadDir string `mapstructure:"UPLOAD_DIR"`
}

// LoadConfig load config
func LoadConfig() (config Config, err error) {
	confFileName := "app.yaml"

	if _, err := os.Stat(confFileName); !os.IsNotExist(err) {
		viper.SetConfigFile(confFileName)
	}

	// Chargement automatiquement les variables d'environement
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	return
}
