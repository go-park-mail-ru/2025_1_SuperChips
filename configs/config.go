package configs

import "os"

type Config struct {
	Port string
}

func LoadConfigFromEnv() Config {

	config := Config{}

	port, ok := os.LookupEnv("PORT")
	if ok {
		config.Port = port
	} else {
		config.Port = ":8080"
	}

	return config
}
