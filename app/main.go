package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	pgStorage "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	middleware "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/middleware"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pin"
	"github.com/go-park-mail-ru/2025_1_SuperChips/user"
	"github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// @title flow API
// @version 1.0
// @description API for Flow.
func main() {
	config := configs.Config{}
	if err := config.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to config error: %s", err)
	}

	pgConfig := configs.PostgresConfig{}
	if err := pgConfig.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to pg config error: %s", err)
	}

	log.Println("Waiting for database to start...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)

	defer cancel()

	// т.к. бд не сразу после запуска начинает принимать запросы
	// пробуем подключиться к бд в течение 10 секунд
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgConfig.PgHost, 5432, pgConfig.PgUser, pgConfig.PgPassword, pgConfig.PgDB)
	db, err := pg.ConnectDB(psqlconn, ctx)
	if err != nil {
		log.Fatalf("Cannot launch due to database connection error: %s", err)
	}

	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize migration driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %s", err)
	}

	if err := m.Up(); err != nil {
		log.Fatalf("Failed to apply migrations: %s", err)
	}

	userStorage, err := pgStorage.NewPGUserStorage(db)
	if err != nil {
		log.Fatalf("Cannot launch due to user storage db error: %s", err)
	}

	pinStorage, err := pgStorage.NewPGPinStorage(db)
	if err != nil {
		log.Fatalf("Cannot launch due to pin storage db error: %s", err)
	}

	jwtManager := auth.NewJWTManager(config)

	userService := user.NewUserService(userStorage)
	pinService := pin.NewPinService(pinStorage)

	authHandler := rest.AuthHandler{
		Config:      config,
		UserService: userService,
		JWTManager:  *jwtManager,
	}

	pinsHandler := rest.PinsHandler{
		Config:     config,
		PinService: pinService,
	}

	allowedGetOptions := []string{http.MethodGet, http.MethodOptions}
	allowedPostOptions := []string{http.MethodPost, http.MethodOptions}

	fs := http.FileServer(http.Dir("." + config.StaticBaseDir))

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix(config.StaticBaseDir, fs))

	mux.HandleFunc("/health", middleware.CorsMiddleware(rest.HealthCheckHandler, config, allowedGetOptions))
	mux.HandleFunc("/api/v1/feed", middleware.CorsMiddleware(pinsHandler.FeedHandler, config, allowedGetOptions))
	mux.HandleFunc("/api/v1/auth/login", middleware.CorsMiddleware(authHandler.LoginHandler, config, allowedPostOptions))
	mux.HandleFunc("/api/v1/auth/registration", middleware.CorsMiddleware(authHandler.RegistrationHandler, config, allowedPostOptions))
	mux.HandleFunc("/api/v1/auth/logout", middleware.CorsMiddleware(authHandler.LogoutHandler, config, allowedPostOptions))
	mux.HandleFunc("/api/v1/auth/user", middleware.CorsMiddleware(authHandler.UserDataHandler, config, allowedGetOptions))

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

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown unsuccessful: %v", err)
	}

	log.Println("Server has been gracefully shut down.")
}
