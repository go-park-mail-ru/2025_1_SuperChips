package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	repository "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/middleware"
	chatWebsocket "github.com/go-park-mail-ru/2025_1_SuperChips/internal/websocket"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/websocket"
	microserviceGrpc "github.com/go-park-mail-ru/2025_1_SuperChips/internal/grpc"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8020")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pgConfig := configs.PostgresConfig{}
	if err := pgConfig.LoadConfigFromEnv(); err != nil {
		log.Fatal(err)
	}

	slog.Info("Waiting for database to start...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgConfig.PgHost, 5432, pgConfig.PgUser, pgConfig.PgPassword, pgConfig.PgDB)
	db, err := pg.ConnectDB(psqlconn, ctx)
	if err != nil {
		log.Fatalf("Cannot launch due to database connection error: %s", err)
	}

	defer db.Close()

	authConfig := configs.AuthConfig{}
	if err := authConfig.LoadConfigFromEnv(); err != nil {
		log.Fatalf("%s", err.Error())
	}

	jwtManager := auth.NewJWTManager(configs.Config{
		JWTSecret: authConfig.JWTSecret,
		ExpirationTime: authConfig.ExpirationTime,
	})

	chatRepo := repository.NewChatRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	hubCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := chatWebsocket.CreateHub(chatRepo, notificationRepo)
	
	chatWebsocketHandler := rest.ChatWebsocketHandler{
		Hub: hub,
		ContextExpiration: time.Second * 5,
	}

	go hub.Run(hubCtx)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", middleware.ChainMiddleware(chatWebsocketHandler.WebSocketUpgrader,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.Log()))

	server := http.Server{
		Addr:    ":8013",
		Handler: mux,
	}

	grpcServer := grpc.NewServer()
	websocketServer := microserviceGrpc.NewGrpcWebsocketHandler(hub)
	gen.RegisterWebsocketServer(grpcServer, websocketServer)

	errorChan := make(chan error, 1)

	go func() {
		log.Printf("Server listening on port %s", ":8013")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errorChan <- err
		}
	}()

	go func() {
		log.Println("Starting server on :8020")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		log.Println("started on port :8020")
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

	grpcServer.GracefulStop()

	log.Println("Server has been gracefully shut down.")
}

