package configs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port           string
	JWTSecret      []byte
	ExpirationTime time.Duration
	CookieSecure   bool
	Environment    string
}

var (
	errMissingJWT = errors.New("missing jwt token key")
	errNoEnv      = errors.New("missing env variable")
)

func printConfig(cfg Config) {
	fmt.Println("-----------------------------------------------")
	fmt.Println("Resulting config: ")
	fmt.Printf("Port: %s\n", cfg.Port)
	fmt.Printf("ExpirationTime: %s\n", cfg.ExpirationTime.String())
	fmt.Printf("CookieSecure: %t\n", cfg.CookieSecure)
	fmt.Printf("Env: %s\n", cfg.Environment)
	fmt.Println("-----------------------------------------------")
}

func LoadConfigFromEnv() (Config, error) {
	config := Config{}

	port, ok := os.LookupEnv("PORT")
	if ok {
		config.Port = ":" + port
	} else {
		config.Port = ":8080"
	}

	jwtSecret, ok := os.LookupEnv("JWT_SECRET")
	if ok {
		config.JWTSecret = []byte(jwtSecret)
	} else {
		return config, errMissingJWT
	}

	expirationTimeStr, ok := os.LookupEnv("EXPIRATION_TIME")
	if ok {
		expirationTime, err := time.ParseDuration(expirationTimeStr)
		if err != nil {
			log.Println("could not parse expiration time, setting default value (15 minutes)")
			config.ExpirationTime = 15 * time.Minute
		} else {
			config.ExpirationTime = expirationTime
		}
	} else {
		log.Println("env variable expirationTime not given, setting default value (15 minutes)")
		config.ExpirationTime = 15 * time.Minute
	}

	cookieSecure, ok := os.LookupEnv("COOKIE_SECURE")
	if ok {
		cookieSecureLower := strings.ToLower(cookieSecure)
		if cookieSecureLower == "true" {
			config.CookieSecure = true
		} else if cookieSecureLower == "false" {
			config.CookieSecure = false
		} else {
			log.Println("Error parsing cookieSecure, assuming false")
		}
	} else {
		log.Println("env variable cookieSecure not give, setting default value (false)")
	}

	env, ok := os.LookupEnv("ENVIRONMENT")
	if ok {
		envLower := strings.ToLower(env)
		if envLower == "prod" {
			config.Environment = envLower
		} else if envLower == "test" {
			config.Environment = envLower
		} else {
			log.Println("could not parse environment variable")
			return config, errNoEnv
		}
	} else {
		return config, errNoEnv
	}

	printConfig(config)

	return config, nil
}
