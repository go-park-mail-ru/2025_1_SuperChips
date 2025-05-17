package configs

import (
	"fmt"
	"testing"
)

func TestLoadPgConfigFromEnv_Success(t *testing.T) {
    t.Setenv("PGUSER", "testuser")
    t.Setenv("PGPASSWORD", "testpass")
    t.Setenv("PGDATABASE", "testdb")
    t.Setenv("PGHOST", "localhost")
    t.Setenv("PGPORT", "5432")

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
    if cfg.PgPort != "5432" {
        t.Errorf("Expected PgPort '5432', got '%s'", cfg.PgPort)
    }
}

var missingEnvTestCases = []struct {
    missingEnv string
}{
    {"PGUSER"},
    {"PGPASSWORD"},
    {"PGDATABASE"},
    {"PGHOST"},
    {"PGPORT"},
}

func TestLoadConfigFromEnv_MissingEnvs(t *testing.T) {
    for _, tc := range missingEnvTestCases {
        t.Run(tc.missingEnv, func(t *testing.T) {
            envs := []string{
                "PGUSER",
                "PGPASSWORD",
                "PGDATABASE",
                "PGHOST",
                "PGPORT",
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