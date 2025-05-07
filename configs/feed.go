package configs

import (
	"log"
	"os"
	"strconv"
	"time"
)

type FeedConfig struct {
	ImageBaseDir      string
	StaticBaseDir     string
	AvatarDir         string
	BaseUrl           string
	PageSize          int
	ContextExpiration time.Duration
}

func (config *FeedConfig) LoadConfigFromEnv() error {
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

	contextDuration, err := getEnvHelper("CONTEXT_DURATION", "3s")
	if err != nil {
		log.Fatalf("Couldn't parse CONTEXT_DURATION: %s", err.Error())
	}

	contextDurationTime, err := time.ParseDuration(contextDuration)
	if err != nil {
		log.Fatalf("Couldn't parse CONTEXT_DURATION: %s", err.Error())
	}

	config.ContextExpiration = contextDurationTime

	staticBaseDir, _ := getEnvHelper("STATIC_BASE_DIR", "/static/")
	config.StaticBaseDir = staticBaseDir

	avatarDir, _ := getEnvHelper("AVATAR_DIR", "avatars")
	config.AvatarDir = avatarDir

	baseUrl, _ := getEnvHelper("BASE_URL", "https://yourflow.ru")
	config.BaseUrl = baseUrl

	config.printConfig()

	return nil
}

func (cfg FeedConfig) printConfig() {
	log.Println("-----------------------------------------------")
	log.Println("Resulting feed config: ")
	log.Printf("Image dir: %s\n", cfg.ImageBaseDir)
	log.Printf("PageSize: %d\n", cfg.PageSize)
	log.Printf("Static base dir: %s\n", cfg.StaticBaseDir)
	log.Printf("Avatar folder: %s\n", cfg.AvatarDir)
	log.Printf("Base URL: %s\n", cfg.BaseUrl)
	log.Println("-----------------------------------------------")
}

