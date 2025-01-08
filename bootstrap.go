package main

import (
	"flag"
	"fmt"
	"github.com/make0x20/driplet/internal/config"
	"github.com/make0x20/driplet/logger"
	"log"
	"log/slog"
	"os"
	"strings"
)

// loadConfig loads the config from the given path.
func loadConfig() *config.Config {
	configPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	fmt.Println(splash())

	cfg, err := config.NewWithPath(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	return cfg
}

// setupLogger creates a new logger with the given config.
func setupLogger(cfg *config.Config) *slog.Logger {
	var file *os.File
	if cfg.Global.LogFile != "" {
		var err error
		file, err = os.OpenFile(cfg.Global.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
	}

	logLevel := slog.LevelInfo
	if cfg.Global.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}

	return logger.New(logLevel, file)
}

// configEndpoints returns a string with the names of the endpoints in the config.
func configEndpoints(c *config.Config) string {
	var endpoints []string
	for _, e := range c.Endpoints {
		endpoints = append(endpoints, e.Name)
	}

	output := "Loaded Driplet config endpoints:"
	output += "  " + strings.Join(endpoints, ", ")

	return output
}

// splash returns the splash screen for Driplet.
func splash() string {
	return "\x1b[1;34m" + `
▓█████▄  ██▀███   ██▓ ██▓███   ██▓    ▓█████▄▄▄█████▓
▒██▀ ██▌▓██ ▒ ██▒▓██▒▓██░  ██▒▓██▒    ▓█   ▀▓  ██▒ ▓▒
░██   █▌▓██ ░▄█ ▒▒██▒▓██░ ██▓▒▒██░    ▒███  ▒ ▓██░ ▒░
░▓█▄   ▌▒██▀▀█▄  ░██░▒██▄█▓▒ ▒▒██░    ▒▓█  ▄░ ▓██▓ ░ 
░▒████▓ ░██▓ ▒██▒░██░▒██▒ ░  ░░██████▒░▒████▒ ▒██▒ ░ 
▒▒▓  ▒ ░ ▒▓ ░▒▓░░▓  ▒▓▒░ ░  ░░ ▒░▓  ░░░ ▒░ ░ ▒ ░░   
░ ▒  ▒   ░▒ ░ ▒░ ▒ ░░▒ ░     ░ ░ ▒  ░ ░ ░  ░   ░    
░ ░  ░   ░░   ░  ▒ ░░░         ░ ░      ░    ░      
░       ░      ░               ░  ░   ░  ░        
░                                                   
Starting Driplet..
` + "\x1b[0m"
}
