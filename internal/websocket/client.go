package websocket

import (
    "github.com/make0x20/driplet/internal/jwt"
    "encoding/json"
    "github.com/gorilla/websocket"
    "sync"
)

// Client holds information about a websocket client.
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    endpoint string
    claims   *jwt.Claims
    topics   []string
    topicsMu sync.RWMutex
}

// NewClient creates a new client.
func NewClient(hub *Hub, conn *websocket.Conn, endpoint string, claims *jwt.Claims) *Client {
    return &Client{
        hub:      hub,
        conn:     conn,
        send:     make(chan []byte, 256),
        endpoint: endpoint,
        claims:   claims,
        topics:   make([]string, 0),
    }
}

// ReadPump reads messages from the client.
func (c *Client) ReadPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

	// Loops indefinitely to read messages from the client until connection is closed
    for {
		// Read the message from the client
        messageType, message, err := c.conn.ReadMessage()
        if err != nil {
            return
        }

		// Handle the message type
        if messageType == websocket.PingMessage {
            if err := c.conn.WriteMessage(websocket.PongMessage, nil); err != nil {
                return
            }
            continue
        }

		// Unmarshal the message
        var subMsg SubscriptionMessage
        if err := json.Unmarshal(message, &subMsg); err != nil {
            continue
        }

		// Handle the message type
        switch subMsg.Type {
        case MessageTypeSubscribe:
            c.topicsMu.Lock()
            if !contains(c.topics, subMsg.Topic) {
                c.topics = append(c.topics, subMsg.Topic)
                c.hub.options.Logger.Debug("Client subscribed to topic",
                    "topic", subMsg.Topic,
                    "client_topics", c.topics,
                )
            }
            c.topicsMu.Unlock()

        case MessageTypeUnsubscribe:
            c.topicsMu.Lock()
            c.topics = removeString(c.topics, subMsg.Topic)
            c.hub.options.Logger.Debug("Client unsubscribed from topic",
                "topic", subMsg.Topic,
                "client_topics", c.topics,
            )
            c.topicsMu.Unlock()
        }
    }
}

// WritePump writes messages to the client.
func (c *Client) WritePump() {
    defer func() {
        c.conn.Close()
    }()

	// Loops indefinitely to write messages to the client until connection is closed
    for {
        select {
		// Wait for a message to be sent
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }

            w.Write(message)

            if err := w.Close(); err != nil {
                return
            }
        }
    }
}
