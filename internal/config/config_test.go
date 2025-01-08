package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadCustomConfig(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "config.toml")

    // Test valid config
    content := `
[Global]
BindAddress = "127.0.0.1"
Port = 8080

[Endpoints.test]
Name = "test"
APISecret = "test-secret"
JWTSecret = "test-jwt-secret"
`
    if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
        t.Fatal(err)
    }

    cfg, err := NewWithPath(configPath)
    if err != nil {
        t.Fatal(err)
    }

    if cfg.Global.BindAddress != "127.0.0.1" {
        t.Errorf("expected BindAddress to be 127.0.0.1, got %s", cfg.Global.BindAddress)
    }

    endpoint, exists := cfg.Endpoints["test"]
    if !exists || endpoint.Name != "test" {
        t.Error("test endpoint not configured correctly")
    }

    // Test invalid TOML syntax
    invalidContent := `
[Global
BindAddress = 123
`
    if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
        t.Fatal(err)
    }

    if _, err := NewWithPath(configPath); err == nil {
        t.Error("expected error with invalid TOML syntax")
    }
}

func TestEnvironmentOverrides(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "config.toml")

    content := `[Global]
BindAddress = "0.0.0.0"
Port = 4719`
    if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
        t.Fatal(err)
    }

    os.Setenv("DRIPLET_GLOBAL_BINDADDRESS", "127.0.0.1")
    os.Setenv("DRIPLET_GLOBAL_PORT", "8080")
    defer func() {
        os.Unsetenv("DRIPLET_GLOBAL_BINDADDRESS")
        os.Unsetenv("DRIPLET_GLOBAL_PORT")
    }()

    cfg, err := NewWithPath(configPath)
    if err != nil {
        t.Fatal(err)
    }

    if cfg.Global.BindAddress != "127.0.0.1" || cfg.Global.Port != 8080 {
        t.Error("environment variables did not override config values")
    }
}

func TestMissingConfigCreatesDefault(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "config.toml")

    cfg, err := NewWithPath(configPath)
    if err != nil {
        t.Fatal(err)
    }

    if _, exists := cfg.Endpoints["default"]; !exists {
        t.Error("default endpoint not created")
    }
}

func TestMultipleEndpointsAndOverrides(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "config.toml")

    // Test multiple endpoints in config
    content := `
[Global]
BindAddress = "127.0.0.1"
Port = 8080

[Endpoints.web]
Name = "web"
APISecret = "web-secret"
JWTSecret = "web-jwt-secret"

[Endpoints.api]
Name = "api"
APISecret = "api-secret"
JWTSecret = "api-jwt-secret"
`
    if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
        t.Fatal(err)
    }

    // Set environment variables to override endpoint configs
    os.Setenv("DRIPLET_ENDPOINTS_WEB_APISECRET", "web-secret-override")
    os.Setenv("DRIPLET_ENDPOINTS_API_JWTSECRET", "api-jwt-override")
    defer func() {
        os.Unsetenv("DRIPLET_ENDPOINTS_WEB_APISECRET")
        os.Unsetenv("DRIPLET_ENDPOINTS_API_JWTSECRET")
    }()

    cfg, err := NewWithPath(configPath)
    if err != nil {
        t.Fatal(err)
    }

    // Check both endpoints exist
    if len(cfg.Endpoints) != 2 {
        t.Errorf("expected 2 endpoints, got %d", len(cfg.Endpoints))
    }

    // Check web endpoint and its override
    if web, exists := cfg.Endpoints["web"]; !exists {
        t.Error("web endpoint not found")
    } else if web.APISecret != "web-secret-override" {
        t.Errorf("web APISecret not overridden, got %s", web.APISecret)
    }

    // Check api endpoint and its override
    if api, exists := cfg.Endpoints["api"]; !exists {
        t.Error("api endpoint not found")
    } else if api.JWTSecret != "api-jwt-override" {
        t.Errorf("api JWTSecret not overridden, got %s", api.JWTSecret)
    }
}
