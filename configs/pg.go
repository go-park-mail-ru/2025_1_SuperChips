package configs

import (
	"errors"
	"os"
)

type PostgresConfig struct {
	PgUser     string // имя пользователя для входа
	PgPassword string // пароль
	PgDB       string // имя бдшки
}

func (config *PostgresConfig) LoadConfigFromEnv() error {
	pgUser, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		return errors.New("missing environment variable POSTGRES_USER")
	}

	pgPassword, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		return errors.New("missing environment variable POSTGRES_PASSWORD")
	}

	pgDB, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		return errors.New("missing environment variable POSTGRES_DB")
	}

	config.PgUser = pgUser
	config.PgPassword = pgPassword
	config.PgDB = pgDB

	return nil
}