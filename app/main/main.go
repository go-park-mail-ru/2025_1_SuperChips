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
	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/go-park-mail-ru/2025_1_SuperChips/profile"
	genAuth "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/auth"
	genChat "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
	genFeed "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/feed"
	"github.com/go-park-mail-ru/2025_1_SuperChips/search"
	"github.com/go-park-mail-ru/2025_1_SuperChips/subscription"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/gorilla/websocket"
	"github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func handleWebSocketProxy(w http.ResponseWriter, r *http.Request) {
    clientConn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("Failed to upgrade client connection", "error", err)
        return
    }
    defer clientConn.Close()

    headers := http.Header{}
    if cookies := r.Header.Get("Cookie"); cookies != "" {
        headers.Add("Cookie", cookies)
    }

    dialCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    microserviceURL := "ws://websocket_chat:8013/ws"
    dialer := websocket.Dialer{
        HandshakeTimeout: 5 * time.Second,
    }
    microserviceConn, _, err := dialer.DialContext(dialCtx, microserviceURL, headers)
    if err != nil {
        slog.Error("Failed to connect to microservice", "error", err)
        return
    }
    defer microserviceConn.Close()

    // 4. Configure ping/pong for BOTH connections
    configureKeepalive(clientConn, 30*time.Second)
    configureKeepalive(microserviceConn, 30*time.Second)

    // 5. Proper bidirectional proxy with error handling
    errChan := make(chan error, 2)
    
    go proxyMessages(clientConn, microserviceConn, errChan, "client->microservice")
    go proxyMessages(microserviceConn, clientConn, errChan, "microservice->client")

    // Wait for first error
    select {
    case err := <-errChan:
        slog.Info("WebSocket proxy terminating", "error", err)
    case <-r.Context().Done():
        slog.Info("WebSocket proxy terminating", "reason", "request context done")
    }
}

func configureKeepalive(conn *websocket.Conn, interval time.Duration) {
    conn.SetReadDeadline(time.Now().Add(interval * 2))
    
    // Create a channel to signal when the connection closes
    done := make(chan struct{})
    
    // Handle incoming pings
    conn.SetPingHandler(func(appData string) error {
        conn.SetReadDeadline(time.Now().Add(interval * 2))
        err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
        if err != nil {
            slog.Error("Failed to send pong", "error", err)
        }
        return err
    })

    // Set up close handler to signal the done channel
    originalCloseHandler := conn.CloseHandler()
    conn.SetCloseHandler(func(code int, text string) error {
        close(done)
        if originalCloseHandler != nil {
            return originalCloseHandler(code, text)
        }
        return nil
    })

    // Send periodic pings
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
                    slog.Error("Failed to send ping", "error", err)
                    return
                }
            case <-done:
                return
            }
        }
    }()
}

func proxyMessages(src, dst *websocket.Conn, errChan chan<- error, direction string) {
    for {
        msgType, msg, err := src.ReadMessage()
        if err != nil {
            errChan <- fmt.Errorf("%s: %w", direction, err)
            return
        }
        
        if err := dst.WriteMessage(msgType, msg); err != nil {
            errChan <- fmt.Errorf("%s: %w", direction, err)
            return
        }
    }
}

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

	subscriptionStorage := pgStorage.NewSubscriptionStorage(db)
	likeStorage := pgStorage.NewPgLikeStorage(db)
	boardStorage := pgStorage.NewBoardStorage(db)
	searchStorage := pgStorage.NewSearchRepository(db)

	jwtManager := auth.NewJWTManager(config)

	subscriptionService := subscription.NewSubscriptionUsecase(subscriptionStorage)
	pinCRUDService := pincrudService.NewPinCRUDService(pinStorage, boardStorage, imageStorage)
	profileService := profile.NewProfileService(profileStorage, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)
	boardService := board.NewBoardService(boardStorage, config.BaseUrl, config.ImageBaseDir)
	likeService := like.NewLikeService(likeStorage)
	searchService := search.NewSearchService(searchStorage, config.BaseUrl, config.ImageBaseDir, config.StaticBaseDir, config.AvatarDir)

	grpcConnAuth, err := grpc.NewClient(
		"auth:8010",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcConnAuth.Close()

	grpcConnFeed, err := grpc.NewClient(
		"feed:8011",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcConnFeed.Close()

	grpcConnChat, err := grpc.NewClient(
		"chat:8012",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcConnChat.Close()

	authClient := genAuth.NewAuthClient(grpcConnAuth)
	feedClient := genFeed.NewFeedClient(grpcConnFeed)
	chatClient := genChat.NewChatServiceClient(grpcConnChat)

	authHandler := rest.AuthHandler{
		Config:      config,
		UserService: authClient,
		JWTManager:  *jwtManager,
		ContextDuration: config.ContextExpiration,
	}

	subscriptionHandler := rest.SubscriptionHandler{
		ContextExpiration: config.ContextExpiration,
		SubscriptionService: subscriptionService,
	}

	chatHandler := rest.ChatHandler{
		ContextExpiration: config.ContextExpiration,
		ChatService: chatClient,
	}

	pinsHandler := rest.PinsHandler{
		Config:     config,
		FeedClient: feedClient,
		ContextExpiration: config.ContextExpiration,
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
			middleware.CorsMiddleware(config, allowedGetOptionsHead),
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
			middleware.CorsMiddleware(config, allowedGetOptionsHead),
			middleware.Log(),
			middleware.Recovery()))

	mux.HandleFunc("/api/v1/search/boards", 
	middleware.ChainMiddleware(searchHander.SearchBoards,
		middleware.CorsMiddleware(config, allowedGetOptionsHead),
		middleware.Log(),
		middleware.Recovery()))

	mux.HandleFunc("/api/v1/search/users", 
	middleware.ChainMiddleware(searchHander.SearchUsers,
		middleware.CorsMiddleware(config, allowedGetOptionsHead),
		middleware.Log(),
		middleware.Recovery()))

	// external id
	mux.HandleFunc("/api/v1/auth/vkid/login",
		middleware.ChainMiddleware(authHandler.ExternalLogin,
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))

	mux.HandleFunc("/api/v1/auth/vkid/register",
		middleware.ChainMiddleware(authHandler.ExternalRegister,
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))


	// subscription
	mux.HandleFunc("GET /api/v1/profile/followers",
		middleware.ChainMiddleware(subscriptionHandler.GetUserFollowers,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))

	mux.HandleFunc("GET /api/v1/profile/following",
		middleware.ChainMiddleware(subscriptionHandler.GetUserFollowing, 
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.Log()))

	mux.HandleFunc("POST /api/v1/subscription",
		middleware.ChainMiddleware(subscriptionHandler.CreateSubscription,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.Log()))

	mux.HandleFunc("DELETE /api/v1/subscription",
	middleware.ChainMiddleware(subscriptionHandler.DeleteSubscription,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedDeleteOptions),
		middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/subscription", 	
	middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, middleware.CorsMiddleware(config, allowedOptions),
	middleware.Log()))

	// chat
	mux.HandleFunc("GET /api/v1/chats", middleware.ChainMiddleware(chatHandler.GetChats,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	mux.HandleFunc("POST /api/v1/chats", middleware.ChainMiddleware(chatHandler.NewChat,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/chats", middleware.ChainMiddleware(chatHandler.GetChats,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	// contacts
	mux.HandleFunc("GET /api/v1/contacts", middleware.ChainMiddleware(chatHandler.GetContacts, 
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/contacts", middleware.ChainMiddleware(chatHandler.GetContacts, 
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))
	
	mux.HandleFunc("POST /api/v1/contacts", middleware.ChainMiddleware(chatHandler.CreateContact,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.Log()))

	// ws
	mux.HandleFunc("/api/v1/ws", middleware.ChainMiddleware(handleWebSocketProxy,
		middleware.AuthMiddleware(jwtManager, true)))

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
