package routes

import (
	"github.com/make0x20/driplet/handlers"
	"github.com/make0x20/driplet/middleware"
	"github.com/make0x20/driplet/internal/config"
	"github.com/make0x20/driplet/internal/jwt"
	"github.com/make0x20/driplet/internal/nonce"
	"github.com/make0x20/driplet/internal/websocket"
	"log/slog"
	"net/http"
)

func Setup(logger *slog.Logger, cfg *config.Config, hub *websocket.Hub) http.Handler {
	mux := http.NewServeMux()

	// Default middleware chain
	defaultChain := middleware.DefaultChain(logger)
	// JWT validator
	validator := jwt.NewValidator(nonce.NewMemoryStore())

	// Frontpage - returns json status ok
	mux.Handle("GET /", defaultChain(
		http.HandlerFunc(handlers.Front())),
	)

	// Websocket endpoint
	mux.Handle("GET /ws/{name}", defaultChain(
		http.HandlerFunc(handlers.WebSocket(logger, cfg, hub, validator))),
	)

	// Publish message endpoint
	mux.Handle("POST /api/{name}/message", defaultChain(
		http.HandlerFunc(handlers.PublishMessage(logger, cfg, hub, validator))),
	)

	// Ping endpoint
	mux.Handle("GET /api/{name}/ping", defaultChain(
		http.HandlerFunc(handlers.Ping(logger, cfg))),
	)

	return mux
}
