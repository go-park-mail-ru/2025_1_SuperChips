package configs

import (
	"errors"
	"log"
	"os"
)

type ConnConfig struct {
	Port           string
	PrometheusPort string
}

var (
	errMissingPort = errors.New("missing port")
)

func (config *ConnConfig) LoadConfigFromEnv() error {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		return errMissingPort
	}
	config.Port = ":" + port

	prometheusPort, err := getEnvHelper("PROMETHEUS_PORT", "2112")
	if err != nil {
		log.Fatalln(err.Error())
	}
	config.PrometheusPort = ":" + prometheusPort

	return nil
}

func (cfg ConnConfig) Print() {
	log.Println("-----------------------------------------------")
	log.Println("Resulting conn config: ")
	log.Printf("Port: %s\n", cfg.Port)
	log.Printf("Prometheus port: %s\n", cfg.PrometheusPort)
	log.Println("-----------------------------------------------")
}
