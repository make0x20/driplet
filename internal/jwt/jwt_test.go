package jwt

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

// TestGetCustomClaim verifies the nested claim access functionality:
// - Access to nested fields using dot notation (e.g. "user.id")
// - Proper handling of non-existent paths
// - Type assertions for nested maps
func TestGetCustomClaim(t *testing.T) {
	claims := &Claims{
		Custom: map[string]interface{}{
			"user": map[string]interface{}{
				"id":   123,
				"role": "admin",
				"meta": map[string]interface{}{
					"verified": true,
				},
			},
		},
	}

	tests := []struct {
		name   string
		path   string
		want   interface{}
		exists bool
	}{
		{"simple path", "user.id", 123, true},
		{"nested path", "user.meta.verified", true, true},
		{"non-existent path", "user.notfound", nil, false},
		{"invalid path", "invalid.path", nil, false},
		{"empty path", "", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := claims.GetCustomClaim(tt.path)
			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}
			if exists && got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockStore implements nonce.Store for testing
type mockStore struct {
	stored map[string]bool
}

func newMockStore() *mockStore {
	return &mockStore{stored: make(map[string]bool)}
}

func (m *mockStore) Check(nonce, endpoint string) bool {
	return m.stored[nonce+endpoint]
}

func (m *mockStore) Store(nonce, endpoint string, expiresAt time.Time) {
	m.stored[nonce+endpoint] = true
}

// TestValidateAPIToken tests the API token validation:
// - HMAC signature verification
// - Timestamp validation (within 60 second window)
// - Nonce uniqueness check
// - Message format validation
func TestValidateAPIToken(t *testing.T) {
	store := newMockStore()
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	v := &Validator{
		nonceStore: store,
		timeNow: func() time.Time {
			return fixedTime
		},
	}

	apiSecret := "test-secret"
	endpoint := "test-endpoint"

	// Test 1: Valid token with correct signature and timestamp
	metadata := MessageMetadata{
		Nonce:     "test-nonce",
		Timestamp: fixedTime.Unix(),
	}
	payload, _ := json.Marshal(metadata)

	mac := hmac.New(jwt.SigningMethodHS256.Hash.New, []byte(apiSecret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	err := v.ValidateAPIToken(signature, payload, endpoint, apiSecret)
	if err != nil {
		t.Errorf("expected valid token to pass: %v", err)
	}

	// Test 2: Expired timestamp (older than 60 seconds)
	oldMetadata := MessageMetadata{
		Nonce:     "old-nonce",
		Timestamp: fixedTime.Add(-2 * time.Minute).Unix(),
	}
	oldPayload, _ := json.Marshal(oldMetadata)
	mac.Reset()
	mac.Write(oldPayload)
	oldSignature := hex.EncodeToString(mac.Sum(nil))

	err = v.ValidateAPIToken(oldSignature, oldPayload, endpoint, apiSecret)
	if err == nil {
		t.Error("expected old timestamp to fail")
	}

	// Test 3: Nonce reuse should fail
	err = v.ValidateAPIToken(signature, payload, endpoint, apiSecret)
	if err == nil {
		t.Error("expected nonce reuse to fail")
	}
}

// TestValidateClientToken tests JWT token validation:
// - Token signature verification
// - Claims extraction and parsing
// - Invalid token handling
func TestValidateClientToken(t *testing.T) {
	v := NewValidator(nil)
	secret := "test-secret"

	// Test 1: Create and validate a valid token
	claims := &Claims{
		Custom: map[string]interface{}{"user": "test"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	gotClaims, err := v.ValidateClientToken(tokenString, secret)
	if err != nil {
		t.Errorf("expected valid token to pass: %v", err)
	}
	if gotClaims.Custom["user"] != "test" {
		t.Errorf("got claims = %v, want user=test", gotClaims.Custom)
	}

	// Test 2: Invalid signature should fail
	_, err = v.ValidateClientToken(tokenString, "wrong-secret")
	if err == nil {
		t.Error("expected invalid signature to fail")
	}

	// Test 3: Malformed token should fail
	_, err = v.ValidateClientToken("invalid-token", secret)
	if err == nil {
		t.Error("expected invalid token format to fail")
	}
}
