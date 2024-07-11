package config

import (
	"fmt"
	"log"
	"os"
)

const Requestlimit = "1"

type DBConnectionCfg struct {
	HostAddress string
	HostPort    string
	User        string
	Password    string
	DBName      string
	SSLMode     string
}

func NewDBConnectionCfg() DBConnectionCfg {
	cfg := DBConnectionCfg{}
	cfg.HostAddress = os.Getenv("POSTGRES_HOST_ADDRESS")
	cfg.HostPort = os.Getenv("POSTGRES_HOST_PORT")
	cfg.User = os.Getenv("POSTGRES_USER")
	cfg.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.DBName = os.Getenv("POSTGRES_DB")
	cfg.SSLMode = os.Getenv("SSL_MODE")

	someIsEmpty := false

	if cfg.HostAddress == "" {
		log.Println("POSTGRES_HOST_ADDRESS env variable is empty")
		someIsEmpty = true
	}

	if cfg.HostPort == "" {
		log.Println("POSTGRES_HOST_PORT env variable is empty")
		someIsEmpty = true
	}

	if cfg.User == "" {
		log.Println("POSTGRES_USER env variable is empty")
		someIsEmpty = true
	}

	if cfg.Password == "" {
		log.Println("POSTGRES_PASSWORD env variable is empty")
		someIsEmpty = true
	}

	if cfg.DBName == "" {
		log.Println("POSTGRES_DB env variable is empty")
		someIsEmpty = true
	}

	if cfg.SSLMode == "" {
		log.Println("SSL_MODE env variable is empty")
		someIsEmpty = true
	}

	if someIsEmpty {
		log.Fatalln("some desktop environments is empty")
	}

	return cfg
}

type OpenWeatherAPIID struct {
	APIID string
}

func NewOpenWeatherAPIID() OpenWeatherAPIID {
	cfg := OpenWeatherAPIID{}
	cfg.APIID = os.Getenv("OPENWEATHERAPI_ID")

	someIsEmpty := false

	if cfg.APIID == "" {
		log.Println("OPENWEATHERAPI_ID env variable is empty")
		someIsEmpty = true
	}

	if someIsEmpty {
		log.Fatalln("some desktop environments is empty")
	}

	return cfg
}

type ServerCfg struct {
	Port string
}

func NewServerCfg() ServerCfg {
	cfg := ServerCfg{}

	cfg.Port = os.Getenv("SERVER_PORT")

	someIsEmpty := false

	if cfg.Port == "" {
		log.Println("SERVER_PORT env variable is empty")
		someIsEmpty = true
	}

	if someIsEmpty {
		log.Fatalln("some desktop environments is empty")
	}
	cfg.Port = fmt.Sprintf(":%s", cfg.Port)

	return cfg
}
