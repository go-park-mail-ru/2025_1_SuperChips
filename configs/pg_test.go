package configs

import (
	"fmt"
	"testing"
)

func TestLoadPgConfigFromEnv_Success(t *testing.T) {
    t.Setenv("POSTGRES_USER", "testuser")
    t.Setenv("POSTGRES_PASSWORD", "testpass")
    t.Setenv("POSTGRES_DB", "testdb")
    t.Setenv("POSTGRES_HOST", "localhost")

    var cfg PostgresConfig
    err := cfg.LoadConfigFromEnv()
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }

    if cfg.PgUser != "testuser" {
        t.Errorf("Expected PgUser 'testuser', got '%s'", cfg.PgUser)
    }
    if cfg.PgPassword != "testpass" {
        t.Errorf("Expected PgPassword 'testpass', got '%s'", cfg.PgPassword)
    }
    if cfg.PgDB != "testdb" {
        t.Errorf("Expected PgDB 'testdb', got '%s'", cfg.PgDB)
    }
    if cfg.PgHost != "localhost" {
        t.Errorf("Expected PgHost 'localhost', got '%s'", cfg.PgHost)
    }
}

var missingEnvTestCases = []struct {
    missingEnv string
}{
    {"POSTGRES_USER"},
    {"POSTGRES_PASSWORD"},
    {"POSTGRES_DB"},
    {"POSTGRES_HOST"},
}

func TestLoadConfigFromEnv_MissingEnvs(t *testing.T) {
    for _, tc := range missingEnvTestCases {
        t.Run(tc.missingEnv, func(t *testing.T) {
            envs := []string{
                "POSTGRES_USER",
                "POSTGRES_PASSWORD",
                "POSTGRES_DB",
                "POSTGRES_HOST",
            }
            for _, env := range envs {
                if env != tc.missingEnv {
                    t.Setenv(env, "dummy")
                }
            }

            var cfg PostgresConfig
            err := cfg.LoadConfigFromEnv()

            expectedErr := fmt.Sprintf("missing environment variable %s", tc.missingEnv)
            if err == nil || err.Error() != expectedErr {
                t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
            }
        })
    }
}