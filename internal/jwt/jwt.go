package jwt

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/make0x20/driplet/internal/nonce"
	"strings"
	"time"
)

// Claims holds JWT token claims
type Claims struct {
	jwt.RegisteredClaims
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// MessageMetadata holds nonce and timestamp for validating API messages
type MessageMetadata struct {
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
}

// GetCustomClaim retrieves a custom claim from the claims
func (c *Claims) GetCustomClaim(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := interface{}(c.Custom)
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// Validator validates JWT and API tokens
type Validator struct {
	nonceStore nonce.Store
	timeNow    func() time.Time
}

// NewValidator creates a new Validator
func NewValidator(store nonce.Store) *Validator {
	if store == nil {
		store = nonce.NewMemoryStore()
	}

	return &Validator{
		nonceStore: store,
		timeNow:    time.Now,
	}
}

// ValidateClientToken validates a client JWT token
func (v *Validator) ValidateClientToken(tokenString string, jwtSecret string) (*Claims, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Validate claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// ValidateAPIToken validates an API token
func (v *Validator) ValidateAPIToken(signature string, payload []byte, endpoint string, apiSecret string) error {
	// Validate HMAC signature
	mac := hmac.New(jwt.SigningMethodHS256.Hash.New, []byte(apiSecret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)

	// Decode signature
	providedMAC, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature format")
	}

	// Compare signatures
	if !hmac.Equal(providedMAC, expectedMAC) {
		return fmt.Errorf("invalid signature")
	}

	// Parse and validate metadata
	var metadata MessageMetadata
	if err := json.Unmarshal(payload, &metadata); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	// Validate timestamp
	now := v.timeNow().Unix()
	if metadata.Timestamp < now-60 || metadata.Timestamp > now+60 {
		return fmt.Errorf("message timestamp outside acceptable range")
	}

	// Check and store nonce
	if v.nonceStore.Check(metadata.Nonce, endpoint) {
		return fmt.Errorf("nonce has been used before")
	}
	v.nonceStore.Store(metadata.Nonce, endpoint, v.timeNow().Add(time.Minute))

	return nil
}
