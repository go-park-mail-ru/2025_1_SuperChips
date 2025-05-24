package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/chat"
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	microserviceGrpc "github.com/go-park-mail-ru/2025_1_SuperChips/internal/grpc"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	repository "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	lis, err := net.Listen("tcp", ":8012")
	if err != nil {
		log.Fatal(err)
	}

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

	config := configs.FeedConfig{}
	if err := config.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to config error: %s", err)
	}

	server := grpc.NewServer()

	chatRepo := repository.NewChatRepository(db)
	chatService := chat.NewChatService(chatRepo, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)

	// hubCtx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	chatServer := microserviceGrpc.NewGrpcChatHandler(chatService)
	gen.RegisterChatServiceServer(server, chatServer)

	go func() {
		log.Println("Starting server on :8012")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		log.Println("started on port :8012")
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	
	server.GracefulStop()
}

