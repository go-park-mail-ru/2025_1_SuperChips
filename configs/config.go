package configs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port              string
	JWTSecret         []byte
	ExpirationTime    time.Duration
	CookieSecure      bool
	Environment       string
	IpAddress         string
	ImageBaseDir      string
	StaticBaseDir     string
	AvatarDir         string
	BaseUrl           string
	PageSize          int
	AllowedOrigins    []string
	ContextExpiration time.Duration
	VKClientID        string
}

var (
	errMissingJWT = errors.New("missing jwt token key")
)

func (config *Config) LoadConfigFromEnv() error {
	port, err := getEnvHelper("PORT", ":8080")
	if err != nil {
		log.Fatalln(err.Error())
	}

	config.Port = ":" + port

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

	var cookieSecureBool bool
	if cookieSecure == "true" {
		cookieSecureBool = true
	} else {
		cookieSecureBool = false
	}

	config.CookieSecure = cookieSecureBool

	env, err := getEnvHelper("ENVIRONMENT")
	if err != nil {
		log.Fatalln(err.Error())
	}

	config.Environment = env

	ipAddress, err := getEnvHelper("IP", "localhost")
	if err != nil {
		log.Fatalln(err.Error())
	}

	config.IpAddress = ipAddress

	imgDir, err := getEnvHelper("IMG_DIR", "./static/img")
	if err != nil {
		log.Fatalln(err.Error())
	}

	config.ImageBaseDir = imgDir

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
		log.Println("env variable pageSize not given, setting default value (20)")
	}

	allowedOrigins, err := getEnvHelper("ALLOWED_ORIGINS", "*")
	if err != nil {
		log.Fatalln(err.Error())
	}

	contextDuration, err := getEnvHelper("CONTEXT_DURATION", "3s")
	if err != nil {
		log.Fatalf("Couldn't parse CONTEXT_DURATION: %s", err.Error())
	}

	contextDurationTime, err := time.ParseDuration(contextDuration)
	if err != nil {
		log.Fatalf("Couldn't parse CONTEXT_DURATION: %s", err.Error())
	}

	config.ContextExpiration = contextDurationTime

	config.AllowedOrigins = strings.Split(allowedOrigins, ",")

	staticBaseDir, _ := getEnvHelper("STATIC_BASE_DIR", "/static/")
	config.StaticBaseDir = staticBaseDir

	avatarDir, _ := getEnvHelper("AVATAR_DIR", "avatars")
	config.AvatarDir = avatarDir

	baseUrl, _ := getEnvHelper("BASE_URL", "https://yourflow.ru")
	config.BaseUrl = baseUrl

	VKClientID, err := getEnvHelper("VK_CLIENT_ID", "")
	if err != nil {
		log.Fatalf("%v", err)
	}

	config.VKClientID = VKClientID

	config.printConfig()

	return nil
}

func (cfg Config) printConfig() {
	log.Println("-----------------------------------------------")
	log.Println("Resulting config: ")
	log.Printf("Port: %s\n", cfg.Port)
	log.Printf("ExpirationTime: %s\n", cfg.ExpirationTime.String())
	log.Printf("CookieSecure: %t\n", cfg.CookieSecure)
	log.Printf("Env: %s\n", cfg.Environment)
	log.Printf("IP: %s\n", cfg.IpAddress)
	log.Printf("Image dir: %s\n", cfg.ImageBaseDir)
	log.Printf("PageSize: %d\n", cfg.PageSize)
	log.Printf("Allowed origins: %s\n", strings.Join(cfg.AllowedOrigins, ", "))
	log.Printf("Static base dir: %s\n", cfg.StaticBaseDir)
	log.Printf("Avatar folder: %s\n", cfg.AvatarDir)
	log.Printf("Base URL: %s\n", cfg.BaseUrl)
	log.Println("-----------------------------------------------")
}

func getEnvHelper(key string, defaultValue ...string) (string, error) {
	value, ok := os.LookupEnv(key)
	if ok {
		return value, nil
	}

	if len(defaultValue) > 0 {
		log.Printf("Variable %s not found, using default value: %s", key, defaultValue[0])
		return defaultValue[0], nil
	}

	errMsg := fmt.Sprintf("Mandatory environment variable %s not set!", key)
	return "", errors.New(errMsg)
}
