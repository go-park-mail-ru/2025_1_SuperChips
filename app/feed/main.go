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

	"github.com/go-park-mail-ru/2025_1_SuperChips/pin"
	microserviceGrpc "github.com/go-park-mail-ru/2025_1_SuperChips/internal/grpc"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/feed"
	repository "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8011")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config := configs.FeedConfig{}
	if err := config.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to config error: %s", err)
	}

	pgConfig := configs.PostgresConfig{}
	if err := pgConfig.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to pg config error: %s", err)
	}

	slog.Info("Waiting for database to start...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	// т.к. бд не сразу после запуска начинает принимать запросы
	// пробуем подключиться к бд в течение 10 секунд
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgConfig.PgHost, 5432, pgConfig.PgUser, pgConfig.PgPassword, pgConfig.PgDB)
	db, err := pg.ConnectDB(psqlconn, ctx)
	if err != nil {
		log.Fatalf("Cannot launch due to database connection error: %s", err)
	}

	defer db.Close()

	server := grpc.NewServer()

	feedRepo, err := repository.NewPGPinStorage(db, config.ImageBaseDir, config.BaseUrl)
	if err != nil {
		log.Fatalf("Error creating pg user storage: %v", err)
	}

	usecase := pin.NewPinService(feedRepo, config.BaseUrl, config.ImageBaseDir)

	feedServer := microserviceGrpc.NewGrpcFeedHandler(usecase)
	gen.RegisterFeedServer(server, feedServer)

	go func() {
		log.Println("Starting server on :8011")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
		log.Println("started on port :8011")
	}()

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	server.GracefulStop()
}

