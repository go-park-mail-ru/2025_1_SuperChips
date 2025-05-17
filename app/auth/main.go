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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	connConfig := configs.ConnConfig{}
	if err := connConfig.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to connection config error: %s", err)
	}

	pgConfig := configs.PostgresConfig{}
	if err := pgConfig.LoadConfigFromEnv(); err != nil {
		log.Fatalf("Cannot launch due to pg config error: %s", err)
	}

	lis, err := net.Listen("tcp", connConfig.Port)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Waiting for database to start...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	// т.к. бд не сразу после запуска начинает принимать запросы
	// пробуем подключиться к бд в течение 10 секунд
	db, err := pg.ConnectDB(pgConfig, ctx)
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
		log.Println("Starting gRPC server on " + connConfig.Port)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Error starting gRPC server: %v", err)
		}
		log.Println("gRPC server started on port " + connConfig.Port)
	}()

	// HTTP сервер для prometheus.
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/api/v1/metrics", promhttp.Handler())
		
		log.Println("Starting HTTP server on " + connConfig.PrometheusPort)
		if err := http.ListenAndServe(connConfig.PrometheusPort, mux); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
		log.Println("HTTP server started on port " + connConfig.PrometheusPort)
	}()

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	server.GracefulStop()
}

