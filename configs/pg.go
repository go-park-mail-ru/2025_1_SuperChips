package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type PostgresConfig struct {
	PgUser     string // имя пользователя для входа
	PgPassword string // пароль
	PgDB       string // имя бдшки
	PgHost     string // хост
	PgPort     string // порт

	MaxOpenConns    int           // Максимальное количество открытых соединений.
	MaxIdleConns    int           // Максимальное количество постоянных соединений.
	ConnMaxIdleTime time.Duration // Максимальное время, которое постоянное соединение будет удерживаться в пуле.
}

const missingVarText = "missing environment variable %s"

func (config *PostgresConfig) LoadConfigFromEnv() error {
	pgUser, ok := os.LookupEnv("PGUSER")
	if !ok {
		return fmt.Errorf(missingVarText, "PGUSER")
	}

	pgPassword, ok := os.LookupEnv("PGPASSWORD")
	if !ok {
		return fmt.Errorf(missingVarText, "PGPASSWORD")
	}

	pgDB, ok := os.LookupEnv("PGDATABASE")
	if !ok {
		return fmt.Errorf(missingVarText, "PGDATABASE")
	}

	pgHost, ok := os.LookupEnv("PGHOST")
	if !ok {
		return fmt.Errorf(missingVarText, "PGHOST")
	}

	pgPort, ok := os.LookupEnv("PGPORT")
	if !ok {
		return fmt.Errorf(missingVarText, "PGPORT")
	}

	config.MaxOpenConns = 0
	maxOpenConns, ok := os.LookupEnv("DB_MAX_OPEN_CONNS")
	if ok {
		maxOpenConnsInt, err := strconv.Atoi(maxOpenConns)
		if err != nil {
			log.Println("error parsing env variable maxOpenConns, setting default value (0)")
		} else {
			config.MaxOpenConns = maxOpenConnsInt
		}
	} else {
		log.Println("error parsing env variable maxOpenConns, setting default value (0)")
	}

	config.MaxIdleConns = 0
	maxIdleConns, ok := os.LookupEnv("DB_MAX_OPEN_CONNS")
	if ok {
		maxIdleConnsInt, err := strconv.Atoi(maxIdleConns)
		if err != nil {
			log.Println("error parsing env variable maxIdleConns, setting default value (0)")
		} else {
			config.MaxIdleConns = maxIdleConnsInt
		}
	} else {
		log.Println("error parsing env variable maxIdleConns, setting default value (0)")
	}

	config.ConnMaxIdleTime = 0
	connMaxIdleTimeStr, ok := os.LookupEnv("DB_CONN_MAX_IDLE_TIME")
	if ok {
		connMaxIdleTime, err := time.ParseDuration(connMaxIdleTimeStr)
		if err != nil {
			log.Println("error parsing env variable connMaxIdleTime, setting default value (0)")
		} else {
			config.ConnMaxIdleTime = connMaxIdleTime
		}
	} else {
		log.Println("env variable connMaxIdleTime not given, setting default value (0)")
	}

	config.PgUser = pgUser
	config.PgPassword = pgPassword
	config.PgDB = pgDB
	config.PgHost = pgHost
	config.PgPort = pgPort

	return nil
}
