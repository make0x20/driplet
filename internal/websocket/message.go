package websocket

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	MessageTypeSubscribe   = "subscribe"
	MessageTypeUnsubscribe = "unsubscribe"
)

// Message is the main message struct
type Message struct {
	Type    string          `json:"type"`
	Content string          `json:"content,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// SubscriptionMessage is a subscription message
type SubscriptionMessage struct {
	Type  string `json:"type"`
	Topic string `json:"topic"`
}

// Target is the target struct
type Target struct {
	Include map[string]interface{} `json:"include,omitempty"`
	Exclude map[string]interface{} `json:"exclude,omitempty"`
}

// BroadcastMessage is a broadcast message
type BroadcastMessage struct {
	Message  json.RawMessage `json:"message"`
	Target   Target          `json:"target"`
	Endpoint string          `json:"endpoint"`
	Topic    string          `json:"topic,omitempty"`
}

// Broadcast sends a message to all clients subscribed to the given topic.
func (h *Hub) Broadcast(msg BroadcastMessage) error {
	if msg.Topic == "" {
		return fmt.Errorf("topic is required for broadcasting messages")
	}

	// Validate the target
	if err := validateTarget(msg.Target); err != nil {
		return fmt.Errorf("invalid target structure: %w", err)
	}

	// Marshal the message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %w", err)
	}

	var unregisterClients []*Client

	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.clients) == 0 {
		h.options.Logger.Debug("No clients connected, skipping broadcast")
		return nil
	}

	// loop through all clients and send the message
	for client := range h.clients {
		// skip if endpoints do not match
		if client.endpoint != msg.Endpoint {
			continue
		}

		// Check if the client is subscribed to the topic
		client.topicsMu.RLock()
		subscribed := contains(client.topics, msg.Topic)
		client.topicsMu.RUnlock()
		if !subscribed {
			continue
		}

		// Check if the client should receive the message based targets
		if h.shouldReceiveMessage(client, msg.Target) {
			select {
			case client.send <- msgBytes:
				h.options.Logger.Debug("Message sent to client",
					"endpoint", client.endpoint,
					"topic", msg.Topic,
				)
			default:
				h.options.Logger.Debug("Client send buffer full, marking for unregistration",
					"endpoint", client.endpoint,
				)
				unregisterClients = append(unregisterClients, client)
			}
		}
	}

	for _, client := range unregisterClients {
		h.unregisterClient(client)
	}

	if len(unregisterClients) > 0 {
		h.options.Logger.Info("Unregistered disconnected clients",
			"count", len(unregisterClients),
		)
	}

	return nil
}

// shouldReceiveMessage checks if a client should receive a message based on the target.
func (h *Hub) shouldReceiveMessage(client *Client, target Target) bool {
	h.options.Logger.Debug("Checking message targeting",
		"client_claims", client.claims.Custom,
		"target_include", target.Include,
		"target_exclude", target.Exclude,
	)

	// No targeting = all clients receive
	if len(target.Include) == 0 && len(target.Exclude) == 0 {
		h.options.Logger.Debug("No targeting specified, all clients receive")
		return true
	}

	// Check if excluded
	for path, targetValue := range target.Exclude {
		claimValue, exists := client.claims.GetCustomClaim(path)
		h.options.Logger.Debug("Checking exclude rule",
			"path", path,
			"target_value", targetValue,
			"claim_exists", exists,
			"claim_value", claimValue,
		)
		if exists && matchValue(claimValue, targetValue) {
			h.options.Logger.Debug("Client excluded by matching exclude rule",
				"path", path,
				"target_value", targetValue,
				"claim_value", claimValue,
			)
			return false
		}
	}

	// No includes = included after passing exclusions
	if len(target.Include) == 0 {
		h.options.Logger.Debug("No inclusion rules and passed exclusions, client included")
		return true
	}

	// Check inclusions
	for path, targetValue := range target.Include {
		claimValue, exists := client.claims.GetCustomClaim(path)
		h.options.Logger.Debug("Checking include rule",
			"path", path,
			"target_value", targetValue,
			"claim_exists", exists,
			"claim_value", claimValue,
		)
		if exists && matchValue(claimValue, targetValue) {
			h.options.Logger.Debug("Client included by matching include rule",
				"path", path,
				"target_value", targetValue,
				"claim_value", claimValue,
			)
			return true
		}
	}

	h.options.Logger.Debug("Client did not match any inclusion rules")
	return false
}

// validateTarget validates the target structure.
func validateTarget(target Target) error {
	for path, value := range target.Include {
		if err := validateTargetValue(path, value); err != nil {
			return fmt.Errorf("invalid include target: %w", err)
		}
	}

	for path, value := range target.Exclude {
		if err := validateTargetValue(path, value); err != nil {
			return fmt.Errorf("invalid exclude target: %w", err)
		}
	}

	return nil
}

// validateTargetValue validates the target value.
func validateTargetValue(path string, value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string, float64, bool, []string:
		return nil
	case []interface{}:
		for _, elem := range v {
			if err := validateTargetValue(path, elem); err != nil {
				return err
			}
		}
		return nil
	case map[string]interface{}:
		return fmt.Errorf("nested objects not supported in target value for path: %s", path)
	default:
		return fmt.Errorf("unsupported type %T for path: %s", value, path)
	}
}

// matchValue checks if the claim value matches the target value.
func matchValue(claimValue, targetValue interface{}) bool {
	if claimValue == nil || targetValue == nil {
		return claimValue == targetValue
	}

	// Check if the value is a supported type
	if tv, ok := targetValue.(float64); ok {
		if cv, ok := claimValue.(int); ok {
			return tv == float64(cv)
		}
		if cv, ok := claimValue.(float64); ok {
			return tv == cv
		}
	}

	_, targetIsSlice := targetValue.([]interface{})
	_, targetIsStrings := targetValue.([]string)
	if targetIsSlice || targetIsStrings {
		var targetSlice []interface{}
		if ss, ok := targetValue.([]string); ok {
			targetSlice = make([]interface{}, len(ss))
			for i, s := range ss {
				targetSlice[i] = s
			}
		} else {
			targetSlice = targetValue.([]interface{})
		}

		if len(targetSlice) == 0 {
			if cs, ok := claimValue.([]interface{}); ok && len(cs) == 0 {
				return true
			}
			if cs, ok := claimValue.([]string); ok && len(cs) == 0 {
				return true
			}
			return false
		}

		_, claimIsSlice := claimValue.([]interface{})
		_, claimIsStrings := claimValue.([]string)
		if claimIsSlice || claimIsStrings {
			var claimSlice []interface{}
			if ss, ok := claimValue.([]string); ok {
				claimSlice = make([]interface{}, len(ss))
				for i, s := range ss {
					claimSlice[i] = s
				}
			} else {
				claimSlice = claimValue.([]interface{})
			}

			for _, tv := range targetSlice {
				for _, cv := range claimSlice {
					if reflect.DeepEqual(tv, cv) {
						return true
					}
				}
			}
		} else {
			for _, v := range targetSlice {
				if reflect.DeepEqual(claimValue, v) {
					return true
				}
			}
		}
		return false
	}

	return reflect.DeepEqual(claimValue, targetValue)
}
