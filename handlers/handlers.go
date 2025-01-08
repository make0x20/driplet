package handlers

import (
	"github.com/make0x20/driplet/internal/config"
	"github.com/make0x20/driplet/internal/jwt"
	"github.com/make0x20/driplet/internal/websocket"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// Frontpage prints ok - used for health checks
func Front() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	}
}

// WebSocket handles WebSocket connections on the API endpoint
func WebSocket(logger *slog.Logger, cfg *config.Config, hub *websocket.Hub, validator *jwt.Validator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.PathValue("name")
		token := r.URL.Query().Get("token")

		// Check if endpoint exists - is valid
		if _, exists := cfg.Endpoints[endpoint]; !exists {
			logger.Debug("Invalid endpoint", "endpoint", endpoint)
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
			return
		}

		// Validate JWT token
		claims, err := validator.ValidateClientToken(token, cfg.Endpoints[endpoint].JWTSecret)
		if err != nil {
			logger.Debug("Invalid token", "endpoint", endpoint, "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Upgrade connection to WebSocket and handle it
		if err := hub.HandleConnection(w, r, endpoint, claims); err != nil {
			logger.Error("Could not upgrade connection", "error", err)
			http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
			return
		}
	}
}

// PublishMessage handles receiving push messages on the API endpoint
func PublishMessage(logger *slog.Logger, cfg *config.Config, hub *websocket.Hub, validator *jwt.Validator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get endpoint from URL parameter
		endpoint := r.PathValue("name")

		// Check if endpoint exists - is valid
		if _, exists := cfg.Endpoints[endpoint]; !exists {
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
			logger.Debug("Invalid endpoint", "endpoint", endpoint)
			return
		}

		// Read entire body for validation
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Error reading body", "error", err)
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}

		// Log message
		logger.Debug("Publishing message",
			"endpoint", endpoint,
			"body", string(body),
		)

		// Validate signature from header
		signature := r.Header.Get("X-Driplet-Signature")
		if signature == "" {
			logger.Debug("Missing signature", "endpoint", endpoint)
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}

		// Validate JWT token
		if err := validator.ValidateAPIToken(signature, body, endpoint, cfg.Endpoints[endpoint].APISecret); err != nil {
			logger.Debug("Invalid signature", "endpoint", endpoint, "error", err)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		// Parse message
		var msg websocket.BroadcastMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			logger.Debug("Invalid message format", "endpoint", endpoint, "error", err)
			http.Error(w, "Invalid message format", http.StatusBadRequest)
			return
		}

		// Set endpoint from URL parameter
		msg.Endpoint = endpoint

		// Broadcast the message
		if err := hub.Broadcast(msg); err != nil {
			logger.Error("Error broadcasting message", "error", err)
			http.Error(w, "Error broadcasting message", http.StatusInternalServerError)
			return
		}

		// Respond with OK
		w.WriteHeader(http.StatusOK)
	}
}

// Ping responds with OK on a valid endpoint
func Ping(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.PathValue("name")

		// Check if endpoint exists - is valid
		if _, exists := cfg.Endpoints[endpoint]; !exists {
			logger.Debug("Invalid endpoint", "endpoint", endpoint)
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
			return
		}

		// Respond with OK
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"endpoint": endpoint,
		})
	}
}
