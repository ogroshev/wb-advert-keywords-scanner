package config

import (
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Database struct {
		Host   	  string
		Name	  string
		User	  string
		Password  string
		Port	  int
	}
	ScanIntervalSec int
	LogLevel string
}

const (
	kDefaultPostgresPort = 5432
)

func LoadConfig(path string) (config Config, err error) {
    viper.AddConfigPath(path)
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")

    viper.AutomaticEnv()
	err = viper.BindEnv("Database.Host", "DATABASE_HOST")
	viper.BindEnv("Database.Name", "DATABASE_NAME")
	viper.BindEnv("Database.User", "DATABASE_USER")
	viper.BindEnv("Database.Password", "DATABASE_PASSWORD")
	viper.BindEnv("Database.Port", "DATABASE_PORT")
	err = viper.BindEnv("ScanIntervalSec", "SCAN_INTERVAL_SEC")
	if err != nil {
		log.Warnf("could not bind ScanIntervalSec: %s", err)
	}
	err = viper.BindEnv("LogLevel", "LOG_LEVEL")
	if err != nil {
		log.Warnf("could not bind LogLevel: %s", err)
	}
	
	viper.SetDefault("Database.Port", kDefaultPostgresPort)
	viper.SetDefault("ScanIntervalSec", 60)
	viper.SetDefault("LogLevel", "info")

    viper.ReadInConfig()
	
    err = viper.Unmarshal(&config)
    return
}