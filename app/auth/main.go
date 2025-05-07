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

	"github.com/go-park-mail-ru/2025_1_SuperChips/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	microserviceGrpc "github.com/go-park-mail-ru/2025_1_SuperChips/internal/grpc"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	repository "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/metrics"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/auth"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8010")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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

	metricsService := metrics.NewMetricsService()
	metricsService.RegisterMetrics()

	server := grpc.NewServer(grpc.UnaryInterceptor(metricsService.ServerMetricsInterceptor))

	authRepo, err := repository.NewPGUserStorage(db)
	if err != nil {
		log.Fatalf("Error creating pg user storage: %v", err)
	}

	boardRepo := repository.NewBoardStorage(db)

	usecase := auth.NewUserService(authRepo, boardRepo)

	authServer := microserviceGrpc.NewGrpcAuthHandler(usecase)
	gen.RegisterAuthServer(server, authServer)

	go func() {
		log.Println("Starting gRPC server on :8010")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Error starting gRPC server: %v", err)
		}
		log.Println("gRPC server started on port :8010")
	}()

	// HTTP сервер для prometheus.
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/api/v1/metrics", promhttp.Handler())
		
		log.Println("Starting HTTP server on :2112")
		if err := http.ListenAndServe(":2112", mux); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
		log.Println("HTTP server started on port :2112")
	}()

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	server.GracefulStop()
}

