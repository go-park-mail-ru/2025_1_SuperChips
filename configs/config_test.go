package configs

import (
    "errors"
    "testing"
    "time"
)

func slicesEqual(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func TestLoadConfigFromEnv_Success(t *testing.T) {
    t.Setenv("JWT_SECRET", "testsecret")
    t.Setenv("ENVIRONMENT", "test")
    t.Setenv("PORT", "9090")
    t.Setenv("EXPIRATION_TIME", "30m")
    t.Setenv("COOKIE_SECURE", "true")
    t.Setenv("IP", "127.0.0.1")
    t.Setenv("IMG_DIR", "/test/img")
    t.Setenv("PAGE_SIZE", "10")
    t.Setenv("ALLOWED_ORIGINS", "http://example.com,http://test.com")
    t.Setenv("STATIC_BASE_DIR", "/test/static")
    t.Setenv("AVATAR_DIR", "test_avatars")
    t.Setenv("BASE_URL", "https://test.com")

    var cfg Config
    err := cfg.LoadConfigFromEnv()
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }

    if string(cfg.JWTSecret) != "testsecret" {
        t.Errorf("Expected JWTSecret 'testsecret', got '%s'", cfg.JWTSecret)
    }
    if cfg.Port != ":9090" {
        t.Errorf("Expected Port ':9090', got '%s'", cfg.Port)
    }
    if cfg.ExpirationTime != 30*time.Minute {
        t.Errorf("Expected ExpirationTime 30m, got '%v'", cfg.ExpirationTime)
    }
    if !cfg.CookieSecure {
        t.Error("Expected CookieSecure true")
    }
    if cfg.Environment != "test" {
        t.Errorf("Expected Environment 'test', got '%s'", cfg.Environment)
    }
    if cfg.IpAddress != "127.0.0.1" {
        t.Errorf("Expected IpAddress '127.0.0.1', got '%s'", cfg.IpAddress)
    }
    if cfg.ImageBaseDir != "/test/img" {
        t.Errorf("Expected ImageBaseDir '/test/img', got '%s'", cfg.ImageBaseDir)
    }
    if cfg.PageSize != 10 {
        t.Errorf("Expected PageSize 10, got '%d'", cfg.PageSize)
    }
    expectedOrigins := []string{"http://example.com", "http://test.com"}
    if !slicesEqual(cfg.AllowedOrigins, expectedOrigins) {
        t.Errorf("Expected AllowedOrigins %v, got %v", expectedOrigins, cfg.AllowedOrigins)
    }
    if cfg.StaticBaseDir != "/test/static" {
        t.Errorf("Expected StaticBaseDir '/test/static', got '%s'", cfg.StaticBaseDir)
    }
    if cfg.AvatarDir != "test_avatars" {
        t.Errorf("Expected AvatarDir 'test_avatars', got '%s'", cfg.AvatarDir)
    }
    if cfg.BaseUrl != "https://test.com" {
        t.Errorf("Expected BaseUrl 'https://test.com', got '%s'", cfg.BaseUrl)
    }
}

func TestLoadConfigFromEnv_MissingJWT(t *testing.T) {
    t.Setenv("ENVIRONMENT", "test")
    var cfg Config
    err := cfg.LoadConfigFromEnv()
    if !errors.Is(err, errMissingJWT) {
        t.Errorf("Expected errMissingJWT, got: %v", err)
    }
}


func TestLoadConfigFromEnv_ExpirationTime(t *testing.T) {
    t.Setenv("JWT_SECRET", "secret")
    t.Setenv("ENVIRONMENT", "test")

    tests := []struct {
        expirationTime string
        expected       time.Duration
    }{
        {"30m", 30 * time.Minute},
        {"2h", 2 * time.Hour},
        {"invalid", 15 * time.Minute},
        {"", 15 * time.Minute},
    }

    for _, tt := range tests {
        t.Run(tt.expirationTime, func(t *testing.T) {
            if tt.expirationTime != "" {
                t.Setenv("EXPIRATION_TIME", tt.expirationTime)
            } else {
                t.Setenv("EXPIRATION_TIME", "")
            }

            var cfg Config
            err := cfg.LoadConfigFromEnv()
            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }

            if cfg.ExpirationTime != tt.expected {
                t.Errorf("Expected ExpirationTime %v, got %v", tt.expected, cfg.ExpirationTime)
            }
        })
    }
}

func TestLoadConfigFromEnv_PageSize(t *testing.T) {
    t.Setenv("JWT_SECRET", "secret")
    t.Setenv("ENVIRONMENT", "test")

    tests := []struct {
        pageSize string
        expected int
    }{
        {"10", 10},
        {"invalid", 20},
        {"", 20},
    }

    for _, tt := range tests {
        t.Run(tt.pageSize, func(t *testing.T) {
            if tt.pageSize != "" {
                t.Setenv("PAGE_SIZE", tt.pageSize)
            } else {
                t.Setenv("PAGE_SIZE", "")
            }

            var cfg Config
            err := cfg.LoadConfigFromEnv()
            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }

            if cfg.PageSize != tt.expected {
                t.Errorf("Expected PageSize %d, got %d", tt.expected, cfg.PageSize)
            }
        })
    }
}

func TestLoadConfigFromEnv_AllowedOrigins(t *testing.T) {
    t.Setenv("JWT_SECRET", "secret")
    t.Setenv("ENVIRONMENT", "test")

    tests := []struct {
        allowedOrigins string
        expected       []string
    }{
        {"http://example.com,http://test.com", []string{"http://example.com", "http://test.com"}},
        {"*", []string{"*"}},
        {"", []string{""}},
    }

    for _, tt := range tests {
        t.Run(tt.allowedOrigins, func(t *testing.T) {
            t.Setenv("ALLOWED_ORIGINS", tt.allowedOrigins)

            var cfg Config
            err := cfg.LoadConfigFromEnv()
            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }

            if !slicesEqual(cfg.AllowedOrigins, tt.expected) {
                t.Errorf("Expected AllowedOrigins %v, got %v", tt.expected, cfg.AllowedOrigins)
            }
        })
    }
}