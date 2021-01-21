package main

import (
	"github.com/aukaskavalci/IKEA_assesment/api"
	"github.com/aukaskavalci/IKEA_assesment/db"
	"github.com/aukaskavalci/IKEA_assesment/postgres"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type configuration struct {
	LogLevel       string `mapstructure:"LOGLEVEL" default:"info"`
	Version        string `mapstructure:"VERSION" required:"true"`
	Environment    string `mapstructure:"ENVIRONMENT" required:"true"`
	BackendTimeout string `mapstructure:"BACKENDTIMEOUT" default:"25s"`
	ListenAddress  string `mapstructure:"LISTENADDRESS" default:":8080"`
	DBDriver       string `mapstructure:"DBDRIVER" required:"true"`
	DBHost         string `mapstructure:"DBHOST" required:"true"`
	DBPort         string `mapstructure:"DBPORT" required:"true"`
	DBUser         string `mapstructure:"DBUSER" required:"true"`
	DBPassword     string `mapstructure:"DBPASSWORD" required:"true"`
	DBName         string `mapstructure:"DBDBNAME" required:"true"`
}

func main() {
	logger := initializeLogger()
	config := setConfig(logger)

	//logger related settings
	lvl, err := logrus.ParseLevel(config.LogLevel)
	if err == nil {
		logger.SetLevel(lvl)
	}
	loggerEntry := logger.WithFields(logrus.Fields{
		"release": config.Version,
		"service": "inventory",
	})

	var inventory db.Inventory

	if config.DBDriver == "postgres" {
		config := postgres.Config{Logger: loggerEntry, Driver: config.DBDriver, Host: config.DBHost, Port: config.DBPort, User: config.DBUser, Password: config.DBPassword, Dbname: config.DBName}
		inventory = postgres.NewPInventory(config)
	}

	server := api.NewServer(inventory)
	server.Config.ListenAddress = config.ListenAddress
	server.Config.BackendTimeout = config.BackendTimeout
	server.Logger = loggerEntry

	err = server.Start()
	if err != nil {
		server.Logger.Fatal("cannot start server:", err)
	}
}

//setConfig gets the required config from env and fills the config
func setConfig(logger *logrus.Logger) configuration {
	var config configuration
	err := envconfig.Process("ISC", &config) //Inventory service config
	if err != nil {
		logger.WithField("err", err).Error("Could not load required config")
	}
	return config
}

//initializeLogger initialize the logger with formatter and caller settings
func initializeLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetReportCaller(true)
	return log
}
