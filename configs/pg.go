package configs

import (
	"fmt"
	"os"
)

type PostgresConfig struct {
	PgUser     string // имя пользователя для входа
	PgPassword string // пароль
	PgDB       string // имя бдшки
	PgHost     string // хост
}

const missingVarText = "missing environment variable %s"

func (config *PostgresConfig) LoadConfigFromEnv() error {
	pgUser, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		return fmt.Errorf(missingVarText, "POSTGRES_USER")
	}

	pgPassword, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		return fmt.Errorf(missingVarText, "POSTGRES_PASSWORD")
	}

	pgDB, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		return fmt.Errorf(missingVarText, "POSTGRES_DB")
	}

	pgHost, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		return fmt.Errorf(missingVarText, "POSTGRES_HOST")
	}

	config.PgUser = pgUser
	config.PgPassword = pgPassword
	config.PgDB = pgDB
	config.PgHost = pgHost

	return nil
}
