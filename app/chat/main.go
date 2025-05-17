package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/chat"
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	microserviceGrpc "github.com/go-park-mail-ru/2025_1_SuperChips/internal/grpc"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	repository "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/middleware"
	chatWebsocket "github.com/go-park-mail-ru/2025_1_SuperChips/internal/websocket"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	connConfig := configs.ConnConfig{}
	if err := connConfig.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to connection config error: %s", err)
	}

	pgConfig := configs.PostgresConfig{}
	if err := pgConfig.LoadConfigFromEnv(); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", connConfig.Port)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Waiting for database to start...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	
	db, err := pg.ConnectDB(pgConfig, ctx)
	if err != nil {
		log.Fatalf("Cannot launch due to database connection error: %s", err)
	}

	defer db.Close()

	config := configs.FeedConfig{}
	if err := config.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to config error: %s", err)
	}

	server := grpc.NewServer()

	chatRepo := repository.NewChatRepository(db)
	chatService := chat.NewChatService(chatRepo, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)

	// hubCtx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	hub := chatWebsocket.CreateHub(chatRepo)
	
	chatWebsocketHandler := rest.ChatWebsocketHandler{
		Hub: hub,
		ContextExpiration: time.Second * 5,
	}

	// go hub.Run(hubCtx)

	http.HandleFunc("/ws", middleware.ChainMiddleware(chatWebsocketHandler.WebSocketUpgrader, 
		middleware.Log()))

	chatServer := microserviceGrpc.NewGrpcChatHandler(chatService)
	gen.RegisterChatServiceServer(server, chatServer)

	go func() {
		log.Println("Starting server on " + connConfig.Port)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		log.Println("started on port " + connConfig.Port)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	
	server.GracefulStop()
}

