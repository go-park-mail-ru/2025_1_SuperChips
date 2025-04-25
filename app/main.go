package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	_ "github.com/go-park-mail-ru/2025_1_SuperChips/docs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	osStorage "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/os/pincrud"
	pgStorage "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	middleware "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/middleware"
	pincrudDelivery "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/pincrud"
	"github.com/go-park-mail-ru/2025_1_SuperChips/like"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pin"
	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/go-park-mail-ru/2025_1_SuperChips/profile"
	"github.com/go-park-mail-ru/2025_1_SuperChips/search"
	"github.com/go-park-mail-ru/2025_1_SuperChips/user"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/swaggo/http-swagger"
)

var (
	allowedGetOptions     = []string{http.MethodGet, http.MethodOptions}
	allowedPostOptions    = []string{http.MethodPost, http.MethodOptions}
	allowedPatchOptions   = []string{http.MethodPatch, http.MethodOptions}
	allowedDeleteOptions  = []string{http.MethodDelete, http.MethodOptions}
	allowedPutOptions     = []string{http.MethodPut, http.MethodOptions}
	allowedOptions        = []string{http.MethodOptions}
	allowedGetOptionsHead = []string{http.MethodGet, http.MethodOptions, http.MethodHead}
)

// @title flow API
// @version 1.0
// @description API for Flow.
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config := configs.Config{}
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

	userStorage, err := pgStorage.NewPGUserStorage(db)
	if err != nil {
		log.Fatalf("Cannot launch due to user storage db error: %s", err)
	}

	pinStorage, err := pgStorage.NewPGPinStorage(db, config.ImageBaseDir, config.BaseUrl)
	if err != nil {
		log.Fatalf("Cannot launch due to pin storage db error: %s", err)
	}

	imageStorage, err := osStorage.NewOSImageStorage(config.ImageBaseDir)
	if err != nil {
		log.Fatalf("Cannot launch due to pin storage db error: %s", err)
	}

	profileStorage, err := pgStorage.NewPGProfileStorage(db)
	if err != nil {
		log.Fatalf("Cannot launch due to profile storage db error: %s", err)
	}

	likeStorage := pgStorage.NewPgLikeStorage(db)
	boardStorage := pgStorage.NewBoardStorage(db)
	searchStorage := pgStorage.NewSearchRepository(db)

	jwtManager := auth.NewJWTManager(config)

	userService := user.NewUserService(userStorage)
	pinCRUDService := pincrudService.NewPinCRUDService(pinStorage, imageStorage)
	pinService := pin.NewPinService(pinStorage, config.BaseUrl, config.ImageBaseDir)
	profileService := profile.NewProfileService(profileStorage, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)
	boardService := board.NewBoardService(boardStorage, config.BaseUrl, config.ImageBaseDir)
	likeService := like.NewLikeService(likeStorage)
	searchService := search.NewSearchService(searchStorage, config.BaseUrl, config.ImageBaseDir, config.StaticBaseDir, config.AvatarDir)

	authHandler := rest.AuthHandler{
		Config:      config,
		UserService: userService,
		JWTManager:  *jwtManager,
	}

	pinsHandler := rest.PinsHandler{
		Config:     config,
		PinService: pinService,
	}

	profileHandler := rest.ProfileHandler{
		ProfileService: profileService,
		JwtManager:     *jwtManager,
		StaticFolder:   config.StaticBaseDir,
		AvatarFolder:   config.AvatarDir,
		BaseUrl:        config.BaseUrl,
		ExpirationTime: config.ExpirationTime,
		CookieSecure:   config.CookieSecure,
	}

	pinCRUDHandler := pincrudDelivery.PinCRUDHandler{
		Config:     config,
		PinService: pinCRUDService,
	}

	likeHandler := rest.LikeHandler{
		LikeService: likeService,
		ContextTimeout: config.ContextExpiration,
	}

	boardHandler := rest.BoardHandler{
		BoardService:    boardService,
		ContextDeadline: config.ContextExpiration,
	}
	
	searchHander := rest.SearchHandler{
		Service: searchService,
		ContextTimeout: config.ContextExpiration,
	}

	fs := http.FileServer(http.Dir("." + config.StaticBaseDir))
	fsHandler := func(w http.ResponseWriter, r *http.Request) {
        fs.ServeHTTP(w, r)
    }

	mux := http.NewServeMux()

	if config.Environment == "test" {
		mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	}

	// static
	mux.Handle("/static/", http.StripPrefix(config.StaticBaseDir, middleware.ChainMiddleware(
		fsHandler,
		middleware.CorsMiddleware(config, allowedGetOptionsHead),
	)))

	// health
	mux.HandleFunc("/health",
		middleware.ChainMiddleware(rest.HealthCheckHandler, middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	// feed
	mux.HandleFunc("/api/v1/feed",
		middleware.ChainMiddleware(pinsHandler.FeedHandler, middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	// auth
	mux.HandleFunc("/api/v1/auth/login",
		middleware.ChainMiddleware(authHandler.LoginHandler, middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.Log()))
	mux.HandleFunc("/api/v1/auth/registration",
		middleware.ChainMiddleware(authHandler.RegistrationHandler, middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.Log()))
	mux.HandleFunc("/api/v1/auth/logout",
		middleware.ChainMiddleware(authHandler.LogoutHandler, middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.Log()))

	// profile
	mux.HandleFunc("/api/v1/profile",
		middleware.ChainMiddleware(profileHandler.CurrentUserProfileHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))
	mux.HandleFunc("/api/v1/users/{username}",
		middleware.ChainMiddleware(profileHandler.PublicProfileHandler,
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))
	mux.HandleFunc("/api/v1/profile/update",
		middleware.ChainMiddleware(profileHandler.PatchUserProfileHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPatchOptions),
			middleware.Log()))
	mux.HandleFunc("/api/v1/profile/avatar",
		middleware.ChainMiddleware(profileHandler.UserAvatarHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))
	mux.HandleFunc("/api/v1/profile/password",
		middleware.ChainMiddleware(profileHandler.ChangeUserPasswordHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))

	// flows
	mux.HandleFunc("OPTIONS /api/v1/flows",
		middleware.ChainMiddleware(func(http.ResponseWriter, *http.Request) {},
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))
	mux.HandleFunc("GET /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.ReadHandler,
			middleware.AuthMiddleware(jwtManager, false),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))
	mux.HandleFunc("DELETE /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.DeleteHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.Log()))
	mux.HandleFunc("PUT /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.UpdateHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPutOptions),
			middleware.Log()))
	mux.HandleFunc("POST /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.CreateHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))

	// likes
	mux.HandleFunc("POST /api/v1/like",
		middleware.ChainMiddleware(likeHandler.LikeFlow, 
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/like", middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, 
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	
	// boards
	mux.HandleFunc("POST /api/v1/boards/{id}/flows",
		middleware.ChainMiddleware(boardHandler.AddToBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))

	mux.HandleFunc("GET /api/v1/boards/{board_id}/flows",
		middleware.ChainMiddleware(boardHandler.GetBoardFlows,
			middleware.AuthMiddleware(jwtManager, false),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/boards/{board_id}/flows",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, middleware.CorsMiddleware(config, allowedOptions),
		middleware.Log()))

	mux.HandleFunc("/api/v1/boards/{board_id}/flows/{id}",
		middleware.ChainMiddleware(boardHandler.DeleteFromBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.Log()))

	mux.HandleFunc("DELETE /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(boardHandler.DeleteBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.Log()))

	mux.HandleFunc("PUT /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(boardHandler.UpdateBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPutOptions),
			middleware.Log()))

	mux.HandleFunc("GET /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(boardHandler.GetBoard,
			middleware.AuthMiddleware(jwtManager, false),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, middleware.CorsMiddleware(config, allowedOptions),
		middleware.Log()))

	mux.HandleFunc("GET /api/v1/users/{username}/boards",
		middleware.ChainMiddleware(boardHandler.GetUserPublic,
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))

	mux.HandleFunc("POST /api/v1/users/{username}/boards",
		middleware.ChainMiddleware(boardHandler.CreateBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/users/{username}/boards",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, middleware.CorsMiddleware(config, allowedOptions),
		middleware.Log()))

	mux.HandleFunc("/api/v1/profile/boards",
		middleware.ChainMiddleware(boardHandler.GetUserAllBoards,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))

	// search
	mux.HandleFunc("/api/v1/search/flows", 
		middleware.ChainMiddleware(searchHander.SearchPins,
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log(),
			middleware.Recovery()))

	mux.HandleFunc("/api/v1/search/boards", 
	middleware.ChainMiddleware(searchHander.SearchBoards,
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log(),
		middleware.Recovery()))

	mux.HandleFunc("/api/v1/search/users", 
	middleware.ChainMiddleware(searchHander.SearchUsers,
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log(),
		middleware.Recovery()))

	// server
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
