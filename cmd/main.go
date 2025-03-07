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
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

func main() {
	config, err := configs.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Cannot launch due to config error: %s", err)
	}	

	storage := user.MapUserStorage{}
	storage.Initialize()

	app := handler.AppHandler{
		Config: config,
		Storage: storage,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.CorsMiddleware(app.HealthCheckHandler, config))
	mux.HandleFunc("POST /api/v1/auth/login", handler.CorsMiddleware(app.LoginHandler, config))
	mux.HandleFunc("POST /api/v1/auth/registration", handler.CorsMiddleware(app.RegistrationHandler, config))
	mux.HandleFunc("POST /api/v1/auth/logout", handler.CorsMiddleware(app.LogoutHandler, config))
	mux.HandleFunc("GET /api/v1/auth/user", handler.CorsMiddleware(app.UserDataHandler, config))

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
	
	log.Println("Server has been gracefully shut down.")
}

