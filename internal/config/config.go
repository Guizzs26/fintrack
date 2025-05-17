package config

import (
	"net/http"
	"time"
)

type Config struct {
	App          AppConfig
	ServerConfig ServerConfig
}

type AppConfig struct {
	Env string
}

/*
- ServerConfig holds the global dependencies and configurations needed to start the server.

- This struct in injected with things like DB connections, loggers, config, etc...
*/
type ServerConfig struct {
	Addr              string
	Router            http.Handler
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func InitConfig() *Config {
	loadEnv()

	addr := mustGetString("ADDR", ":3333")
	env := mustGetString("ENV", "development")
	readTimeOut := mustGetInt("READ_TIMEOUT", 10)
	readHeaderTimeout := mustGetInt("READ_HEADER_TIMEOUT", 5)
	writeTimeout := mustGetInt("WRITE_TIMEOUT", 10)
	idleTimeout := mustGetInt("IDLE_TIMEOUT", 60)

	appCfg := AppConfig{
		Env: env,
	}

	srvCfg := ServerConfig{
		Addr:              addr,
		ReadTimeout:       time.Second * time.Duration(readTimeOut),
		ReadHeaderTimeout: time.Second * time.Duration(readHeaderTimeout),
		WriteTimeout:      time.Second * time.Duration(writeTimeout),
		IdleTimeout:       time.Second * time.Duration(idleTimeout),
	}

	cfg := &Config{
		App:          appCfg,
		ServerConfig: srvCfg,
	}

	return cfg
}
