package websocket

import (
    "github.com/gorilla/websocket"
	"github.com/make0x20/driplet/internal/jwt"
    "log/slog"
    "net/http"
    "sync"
)

// HubOptions holds the options for the hub
type HubOptions struct {
    Logger          *slog.Logger
    Upgrader        *websocket.Upgrader
    ReadBufferSize  int
    WriteBufferSize int
}

type Option func(*HubOptions)

// Hub is the main websocket hub.
type Hub struct {
    options    *HubOptions
    clients    map[*Client]bool
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

// NewHub creates a new websocket hub
func defaultHubOptions() *HubOptions {
    return &HubOptions{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    }
}

// NewHub creates a new websocket hub
func NewHub(logger *slog.Logger, options ...Option) *Hub {
    if logger == nil {
        panic("logger is required")
    }

    opts := defaultHubOptions()
    opts.Logger = logger

    for _, opt := range options {
        opt(opts)
    }

    if opts.Upgrader == nil {
        opts.Upgrader = &websocket.Upgrader{
            ReadBufferSize:  opts.ReadBufferSize,
            WriteBufferSize: opts.WriteBufferSize,
            CheckOrigin: func(r *http.Request) bool {
                return true
            },
        }
    }

    return &Hub{
        options:    opts,
        clients:    make(map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

// Run starts the hub
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
        }
    }
}

// HandleConnection handles websocket connections
func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request, endpoint string, claims *jwt.Claims) error {
    conn, err := h.options.Upgrader.Upgrade(w, r, nil)
    if err != nil {
        return err
    }

    client := NewClient(h, conn, endpoint, claims)
    h.options.Logger.Info("Created new client",
        "endpoint", endpoint,
        "claims", claims.Custom,
    )

    h.register <- client
    go client.WritePump()
    go client.ReadPump()
    return nil
}

// unregisterClient unregisters a client from the hub
func (h *Hub) unregisterClient(client *Client) {
    select {
    case h.unregister <- client:
        h.options.Logger.Debug("Client queued for unregistration")
    default:
        go func() {
            h.options.Logger.Debug("Unregister channel full, queuing in goroutine")
            h.unregister <- client
        }()
    }
}

// Helper functions
func contains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}

func removeString(slice []string, str string) []string {
    result := make([]string, 0, len(slice))
    for _, s := range slice {
        if s != str {
            result = append(result, s)
        }
    }
    return result
}
