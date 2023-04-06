package main

import (
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/config"
)

// Config for application
type Config struct {
	DBConnection   string `json:"db_connection"`
	BindAddress    string `json:"bind_address"`
	MetricsAddress string `json:"metrics_address"`
}

var conf *Config

func main() {
	logger := hclog.Default()

	// Create a new config watcher
	c, err := config.New(
		"./config.json",
		1*time.Second,
		logger.StandardLogger(&hclog.StandardLoggerOptions{}),
		func(c *Config) {
			conf = c
			logger.Info("Config file updated", "config", conf)
		},
	)

	conf = c.Read()

	if err != nil {
		logger.Error("Unable to load config file", "error", err)
		os.Exit(1)
	}

	// shutdown the config watcher when the application exits
	defer c.Close()

	logger.Info("Config loaded", "config", conf)

	// block
	for {
		time.Sleep(100 * time.Millisecond)
	}
}
