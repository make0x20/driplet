package main

import (
    "github.com/make0x20/driplet/internal/websocket"
	"github.com/make0x20/driplet/routes"
	"net/http"
	"os"
    "fmt"
)

func main() {
	// Load the config
    cfg := loadConfig()

	// Setup the logger
    logger := setupLogger(cfg)
    
	// Debug log config
    logger.Debug("Loaded Driplet config",
        "bind_address", fmt.Sprintf("%s:%d", cfg.Global.BindAddress, cfg.Global.Port),
        "endpoint_count", len(cfg.Endpoints),
        "endpoints", fmt.Sprintf("%+v", cfg.Endpoints),
    )

	// Log the endpoints
    logger.Info(configEndpoints(cfg))

	// Create a websocket hub
    hub := websocket.NewHub(logger)
    go hub.Run()

	// Setup routes
    r := routes.Setup(logger, cfg, hub)
	addr := fmt.Sprintf("%s:%d", cfg.Global.BindAddress, cfg.Global.Port)

	// Start the server
	logger.Info("Starting Driplet server", "address", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("error starting server", "error", err)
		os.Exit(1)
	}
}


