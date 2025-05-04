package configs

import (
	"log"
	"os"
	"time"
)

type AuthConfig struct {
	JWTSecret         []byte
	ExpirationTime    time.Duration
	CookieSecure      bool
}

func (config *AuthConfig) LoadConfigFromEnv() error {
	jwtSecret, ok := os.LookupEnv("JWT_SECRET")
	if ok {
		config.JWTSecret = []byte(jwtSecret)
	} else {
		return errMissingJWT
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

	cookieSecure, err := getEnvHelper("COOKIE_SECURE", "false")
	if err != nil {
		log.Fatalln(err.Error())
	} else if cookieSecure != "false" && cookieSecure != "true" {
		log.Fatalln("error parsing cookie_secure variable")
	}

	config.printConfig()

	return nil
}

func (cfg AuthConfig) printConfig() {
	log.Println("-----------------------------------------------")
	log.Println("Resulting config: ")
	log.Printf("ExpirationTime: %s\n", cfg.ExpirationTime.String())
	log.Printf("CookieSecure: %t\n", cfg.CookieSecure)
	log.Println("-----------------------------------------------")
}
