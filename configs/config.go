package configs

import "os"

type Config struct {
	Port      string
	JWTSecret []byte
}

func LoadConfigFromEnv() Config {

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
		config.JWTSecret = []byte("default_key")
	}

	return config
}

