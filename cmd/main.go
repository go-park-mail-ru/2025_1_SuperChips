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
	userMap "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/map"
	pinSlice "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/slice"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	middleware "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/middleware"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pin"
	"github.com/go-park-mail-ru/2025_1_SuperChips/user"
)

// @title flow API
// @version 1.0
// @description API for Flow.
func main() {
	config, err := configs.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Cannot launch due to config error: %s", err)
	}

	userStorage := userMap.NewMapUserStorage()
	pinStorage := pinSlice.NewPinSliceStorage(config)

	jwtManager := auth.NewJWTManager(config)

	userService := user.NewUserService(&userStorage)
	pinService := pin.NewPinService(&pinStorage)

	app := rest.AppHandler{
		Config:      config,
		UserService: *userService,
		PinService:  *pinService,
		JWTManager:  *jwtManager,
	}

	allowedGetOptions := []string{http.MethodGet, http.MethodOptions}
	allowedPostOptions := []string{http.MethodPost, http.MethodOptions}

	fs := http.FileServer(http.Dir("./static/"))

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/health", middleware.CorsMiddleware(app.HealthCheckHandler, config, allowedGetOptions))
	mux.HandleFunc("/api/v1/feed", middleware.CorsMiddleware(app.FeedHandler, config, allowedGetOptions))
	mux.HandleFunc("/api/v1/auth/login", middleware.CorsMiddleware(app.LoginHandler, config, allowedPostOptions))
	mux.HandleFunc("/api/v1/auth/registration", middleware.CorsMiddleware(app.RegistrationHandler, config, allowedPostOptions))
	mux.HandleFunc("/api/v1/auth/logout", middleware.CorsMiddleware(app.LogoutHandler, config, allowedPostOptions))
	mux.HandleFunc("/api/v1/auth/user", middleware.CorsMiddleware(app.UserDataHandler, config, allowedGetOptions))

	server := http.Server{
		Addr:    config.Port,
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
	case err := <-errorChan:
		log.Printf("Error initializing the server: %v Terminating.", err)
	case <-shutdown:
		log.Println("Termination signal detected, shutting down gracefully.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown unsuccessful: %v", err)
	}

	log.Println("Server has been gracefully shut down.")
}
