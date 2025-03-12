package configs

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	JWTSecret      []byte
	ExpirationTime time.Duration
	CookieSecure   bool
	Environment    string
	IpAddress      string
	ImageBaseDir   string
	PageSize       int
}

var (
	errMissingJWT = errors.New("missing jwt token key")
	errNoEnv      = errors.New("missing env variable")
)

func printConfig(cfg Config) {
	log.Println("-----------------------------------------------")
	log.Println("Resulting config: ")
	log.Printf("Port: %s\n", cfg.Port)
	log.Printf("ExpirationTime: %s\n", cfg.ExpirationTime.String())
	log.Printf("CookieSecure: %t\n", cfg.CookieSecure)
	log.Printf("Env: %s\n", cfg.Environment)
	log.Printf("IP: %s\n", cfg.IpAddress)
	log.Printf("Image dir: %s\n", cfg.ImageBaseDir)
	log.Printf("PageSize: %d\n", cfg.PageSize)
	log.Println("-----------------------------------------------")
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
		log.Println("env variable cookieSecure not given, setting default value (false)")
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

	ipAddress, ok := os.LookupEnv("IP")
	if ok {
		config.IpAddress = ipAddress
	} else {
		log.Println("env variable IpAddress not given, setting default value (localhost)")
		config.IpAddress = "localhost"
	}

	imgDir, ok := os.LookupEnv("IMG_DIR")
	if ok {
		config.ImageBaseDir = imgDir
	} else {
		log.Println("env variable ImageBaseDir not given, setting default value (./static/img)")
		config.ImageBaseDir = "./static/img"
	}

	config.PageSize = 20
	pageSize, ok := os.LookupEnv("PAGE_SIZE")
	if ok {
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil {
			log.Println("error parsing env variable pageSize, assuming 20")
		} else {
			config.PageSize = pageSizeInt
		}
	} else {
		log.Println("env variable pageSize not given, setting defaul value (20)")
	}

	printConfig(config)

	return config, nil
}
