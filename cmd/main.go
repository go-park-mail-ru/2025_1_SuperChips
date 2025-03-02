package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthCheckHandler)
	mux.HandleFunc("POST /api/v1/auth/login", handler.LoginHandler)
	mux.HandleFunc("POST /api/v1/auth/registration", handler.RegistrationHandler)
	mux.HandleFunc("POST /api/v1/auth/logout", handler.LogoutHandler)
	mux.HandleFunc("GET /api/v1/auth/user", handler.UserDataHandler)

	config := configs.LoadConfigFromEnv()

	server := http.Server{
		Addr: config.Port,
		Handler: mux,
	}

	errorChan := make(chan error, 1)

	go func() {
		log.Printf("Server listening on port %s", config.Port)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errorChan <- err
		}		
	}()

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <- errorChan:
		log.Printf("Error initializing the server: %v Terminating.", err)
	case <- shutdown:
		log.Println("Termination signal detected, shutting down gracefully.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown unsuccessful: %v", err)
	}
	
	log.Println("Server was gracefully shut down.")
}
