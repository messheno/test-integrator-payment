package models

import "github.com/spf13/viper"

// Config strcuture de configuration
type Config struct {
	PortAPI string `mapstructure:"PORT_API"`
	DBType  int32  `mapstructure:"DB_TYPE"`
	DBHost  string `mapstructure:"DB_HOST"`
	DBPort  string `mapstructure:"DB_PORT"`
	DBName  string `mapstructure:"DB_NAME"`
	DBUser  string `mapstructure:"DB_USER"`
	DBPass  string `mapstructure:"DB_PASSWORD"`
	DBPath  string `mapstructure:"DB_PATH"`
}

// LoadConfig load config
func LoadConfig() (config Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigFile("app.yaml")

	// Chargement automatiquement les variables d'environement
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	return
}
