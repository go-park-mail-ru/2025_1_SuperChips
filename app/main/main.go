package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/comment"
	boardshrService "github.com/go-park-mail-ru/2025_1_SuperChips/boardshr"
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	_ "github.com/go-park-mail-ru/2025_1_SuperChips/docs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/pg"
	osStorage "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/os/pincrud"
	pgStorage "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	boardshrDelivery "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/boardshr"
	middleware "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/middleware"
	pincrudDelivery "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/pincrud"
	"github.com/go-park-mail-ru/2025_1_SuperChips/like"
	"github.com/go-park-mail-ru/2025_1_SuperChips/metrics"
	"github.com/go-park-mail-ru/2025_1_SuperChips/notification"
	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/go-park-mail-ru/2025_1_SuperChips/profile"
	genAuth "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/auth"
	genChat "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
	genFeed "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/feed"
	genWebsocket "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/websocket"
	"github.com/go-park-mail-ru/2025_1_SuperChips/search"
	"github.com/go-park-mail-ru/2025_1_SuperChips/subscription"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
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

func ToProtoWebMessage(msg domain.WebMessage) (*genWebsocket.WebMessage, error) {
    var contentMap map[string]interface{}

    switch content := msg.Content.(type) {
    case map[string]interface{}:
        contentMap = content
    default:
        jsonData, err := json.Marshal(content)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal content: %w", err)
        }
        
        if err := json.Unmarshal(jsonData, &contentMap); err != nil {
            return nil, fmt.Errorf("failed to convert content to map: %w", err)
        }
    }

    structContent, err := structpb.NewStruct(contentMap)
    if err != nil {
        return nil, err
    }

    return &genWebsocket.WebMessage{
        Type:    msg.Type,
        Content: structContent,
    }, nil
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
	boardShrStorage := pgStorage.NewBoardShrStorage(db)
	searchStorage := pgStorage.NewSearchRepository(db)
	chatStorage := pgStorage.NewChatRepository(db)
	commentStorage := pgStorage.NewCommentRepository(db)
	notificationStorage := pgStorage.NewNotificationRepository(db)

	jwtManager := auth.NewJWTManager(config)

	subscriptionService := subscription.NewSubscriptionUsecase(subscriptionStorage, chatStorage, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)
	pinCRUDService := pincrudService.NewPinCRUDService(pinStorage, boardStorage, imageStorage)
	profileService := profile.NewProfileService(profileStorage, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)
	boardService := board.NewBoardService(boardStorage, pinStorage, boardShrStorage, config.BaseUrl, config.ImageBaseDir)
	boardShrService := boardshrService.NewBoardShrService(boardShrStorage, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)
	likeService := like.NewLikeService(likeStorage, pinStorage)
	searchService := search.NewSearchService(searchStorage, config.BaseUrl, config.ImageBaseDir, config.StaticBaseDir, config.AvatarDir)
	commentService := comment.NewCommentService(commentStorage, pinStorage, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)
	notificationService := notification.NewNotificationService(notificationStorage, config.BaseUrl, config.StaticBaseDir, config.AvatarDir)

	metricsService := metrics.NewMetricsService()
	metricsService.RegisterMetrics()

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

	grpcConnWebsocket, err := grpc.NewClient(
		"websocket:8020",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcConnWebsocket.Close()

	authClient := genAuth.NewAuthClient(grpcConnAuth)
	feedClient := genFeed.NewFeedClient(grpcConnFeed)
	chatClient := genChat.NewChatServiceClient(grpcConnChat)
	websocketClient := genWebsocket.NewWebsocketClient(grpcConnWebsocket)

	notificationChan := make(chan domain.WebMessage)
	defer close(notificationChan)

	go func() {
		for domainMsg := range notificationChan {
			protoMsg, err := ToProtoWebMessage(domainMsg)
			if err != nil {
				log.Printf("conversion error: %v", err)
				continue
			}
	
			_, err = websocketClient.SendWebMessage(context.Background(), &genWebsocket.SendWebMessageRequest{
				WebMessage: protoMsg,
			})
			if err != nil {
				log.Printf("gRPC send error: %v", err)
			}
		}
	}()

	authHandler := rest.AuthHandler{
		Config:      config,
		UserService: authClient,
		JWTManager:  *jwtManager,
		ContextDuration: config.ContextExpiration,
	}

	subscriptionHandler := rest.SubscriptionHandler{
		ContextExpiration: config.ContextExpiration,
		SubscriptionService: subscriptionService,
		NotificationChan: notificationChan,
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
		NotificationChan: notificationChan,
	}

	boardHandler := rest.BoardHandler{
		BoardService:    boardService,
		ContextDeadline: config.ContextExpiration,
	}

	boardShrHandler := boardshrDelivery.BoardShrHandler{
		BoardShrService: boardShrService,
		ContextDeadline: config.ContextExpiration,
	}
	
	searchHander := rest.SearchHandler{
		Service: searchService,
		ContextTimeout: config.ContextExpiration,
	}
	
	notificationHandler := rest.NotificationHandler{
		NotificationService: notificationService,
		ContextExpiration: config.ContextExpiration,
	}

	commentHandler := rest.CommentHandler{
		Service: commentService,
		ContextExpiration: config.ContextExpiration,
	}

	fs := http.FileServer(http.Dir("." + config.StaticBaseDir))
	fsHandler := func(w http.ResponseWriter, r *http.Request) {
        fs.ServeHTTP(w, r)
    }

	mux := http.NewServeMux()

	if config.Environment == "test" {
		mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	}

	// prometheus
	mux.Handle("/api/v1/metrics", promhttp.Handler())

	mux.HandleFunc("/api/v1/dummy/{ms}",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
			waitTime, err := strconv.Atoi(r.PathValue("ms"))
			if err == nil {
				time.Sleep(time.Duration(waitTime) * time.Millisecond)
			}
		}, 
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))

	fsContext := context.Background()

	// static
	mux.Handle("/static/", http.StripPrefix(config.StaticBaseDir, middleware.ChainMiddleware(
		fsHandler,
		middleware.AuthMiddleware(jwtManager, false),
		middleware.Fileserver(fsContext, authClient),
		middleware.CorsMiddleware(config, allowedGetOptionsHead),
	)))

	// health
	mux.HandleFunc("/health",
		middleware.ChainMiddleware(rest.HealthCheckHandler, middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))

	// feed
	mux.HandleFunc("/api/v1/feed",
		middleware.ChainMiddleware(pinsHandler.FeedHandler, middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))

	// auth
	mux.HandleFunc("/api/v1/auth/login",
		middleware.ChainMiddleware(authHandler.LoginHandler, middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))
	mux.HandleFunc("/api/v1/auth/registration",
		middleware.ChainMiddleware(authHandler.RegistrationHandler, middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))
	mux.HandleFunc("/api/v1/auth/logout",
		middleware.ChainMiddleware(authHandler.LogoutHandler, middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))

	// profile
	mux.HandleFunc("/api/v1/profile",
		middleware.ChainMiddleware(profileHandler.CurrentUserProfileHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("/api/v1/users/{username}",
		middleware.ChainMiddleware(profileHandler.PublicProfileHandler,
			middleware.CorsMiddleware(config, allowedGetOptionsHead),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("/api/v1/profile/update",
		middleware.ChainMiddleware(profileHandler.PatchUserProfileHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPatchOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("/api/v1/profile/avatar",
		middleware.ChainMiddleware(profileHandler.UserAvatarHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("/api/v1/profile/password",
		middleware.ChainMiddleware(profileHandler.ChangeUserPasswordHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	// flows
	mux.HandleFunc("OPTIONS /api/v1/flows",
		middleware.ChainMiddleware(func(http.ResponseWriter, *http.Request) {},
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("GET /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.ReadHandler,
			middleware.AuthMiddleware(jwtManager, false),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("DELETE /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.DeleteHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("PUT /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.UpdateHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPutOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	mux.HandleFunc("POST /api/v1/flows",
		middleware.ChainMiddleware(pinCRUDHandler.CreateHandler,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	// likes
	mux.HandleFunc("POST /api/v1/like",
		middleware.ChainMiddleware(likeHandler.LikeFlow, 
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/like", middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, 
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))
	
	
	// boards
	mux.HandleFunc("OPTIONS /api/v1/boards/{board_id}/flows",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			}, 
			middleware.CorsMiddleware(config, allowedOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("POST /api/v1/boards/{id}/flows",
		middleware.ChainMiddleware(boardHandler.AddToBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("GET /api/v1/boards/{board_id}/flows",
		middleware.ChainMiddleware(boardHandler.GetBoardFlows,
			middleware.AuthMiddleware(jwtManager, false),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/boards/{board_id}/flows/{id}",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			},
			middleware.CorsMiddleware(config, allowedOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	mux.HandleFunc("GET /api/v1/boards/{board_id}/flows/{id}",
		middleware.ChainMiddleware(boardHandler.GetFromBoard,
			middleware.AuthMiddleware(jwtManager, false),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("DELETE /api/v1/boards/{board_id}/flows/{id}",
		middleware.ChainMiddleware(boardHandler.DeleteFromBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	mux.HandleFunc("OPTIONS /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}, 
			middleware.CorsMiddleware(config, allowedOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("DELETE /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(boardHandler.DeleteBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("PUT /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(boardHandler.UpdateBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPutOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
		
	mux.HandleFunc("GET /api/v1/boards/{board_id}",
		middleware.ChainMiddleware(boardHandler.GetBoard,
			middleware.AuthMiddleware(jwtManager, false),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	mux.HandleFunc("OPTIONS /api/v1/users/{username}/boards",
		middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}, 
			middleware.CorsMiddleware(config, allowedOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
				
	mux.HandleFunc("GET /api/v1/users/{username}/boards",
		middleware.ChainMiddleware(boardHandler.GetUserPublic,
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("POST /api/v1/users/{username}/boards",
		middleware.ChainMiddleware(boardHandler.CreateBoard,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("/api/v1/profile/boards",
		middleware.ChainMiddleware(boardHandler.GetUserAllBoards,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	

	// board sharing (author)
	mux.HandleFunc("GET /api/v1/boards/{board_id}/invites",
		middleware.ChainMiddleware(boardShrHandler.GetInvitationLinks,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))		

	mux.HandleFunc("POST /api/v1/boards/{board_id}/invites",
		middleware.ChainMiddleware(boardShrHandler.CreateInvitation,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	mux.HandleFunc("DELETE /api/v1/boards/{board_id}/invites/{link}",
		middleware.ChainMiddleware(boardShrHandler.DeleteInvitation,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	mux.HandleFunc("GET /api/v1/boards/{board_id}/coauthors",
		middleware.ChainMiddleware(boardShrHandler.GetCoauthors,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	mux.HandleFunc("DELETE /api/v1/boards/{board_id}/coauthors",
		middleware.ChainMiddleware(boardShrHandler.DeleteCoauthor,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	

	// board sharing (coauthor)
	mux.HandleFunc("POST /api/v1/join/{link}",
		middleware.ChainMiddleware(boardShrHandler.UseInvitationLink,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	mux.HandleFunc("DELETE /api/v1/boards/{board_id}/coauthoring",
		middleware.ChainMiddleware(boardShrHandler.RefuseCoauthoring,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedDeleteOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))
	
	
	// search
	mux.HandleFunc("/api/v1/search/flows", 
		middleware.ChainMiddleware(searchHander.SearchPins,
			middleware.CorsMiddleware(config, allowedGetOptionsHead),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log(),
			middleware.Recovery()))

	mux.HandleFunc("/api/v1/search/boards", 
	middleware.ChainMiddleware(searchHander.SearchBoards,
		middleware.CorsMiddleware(config, allowedGetOptionsHead),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log(),
		middleware.Recovery()))

	mux.HandleFunc("/api/v1/search/users", 
	middleware.ChainMiddleware(searchHander.SearchUsers,
		middleware.CorsMiddleware(config, allowedGetOptionsHead),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log(),
		middleware.Recovery()))

	
	// external id
	mux.HandleFunc("/api/v1/auth/vkid/login",
		middleware.ChainMiddleware(authHandler.ExternalLogin,
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("/api/v1/auth/vkid/register",
		middleware.ChainMiddleware(authHandler.ExternalRegister,
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))


	// subscription
	mux.HandleFunc("GET /api/v1/profile/followers",
		middleware.ChainMiddleware(subscriptionHandler.GetUserFollowers,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("GET /api/v1/profile/following",
		middleware.ChainMiddleware(subscriptionHandler.GetUserFollowing, 
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CorsMiddleware(config, allowedGetOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("POST /api/v1/subscription",
		middleware.ChainMiddleware(subscriptionHandler.CreateSubscription,
			middleware.AuthMiddleware(jwtManager, true),
			middleware.CSRFMiddleware(),
			middleware.CorsMiddleware(config, allowedPostOptions),
			middleware.MetricsMiddleware(metricsService),
			middleware.Log()))

	mux.HandleFunc("DELETE /api/v1/subscription",
	middleware.ChainMiddleware(subscriptionHandler.DeleteSubscription,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedDeleteOptions),
		middleware.MetricsMiddleware(metricsService),
		middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/subscription", 	
	middleware.ChainMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, middleware.CorsMiddleware(config, allowedOptions),
	middleware.MetricsMiddleware(metricsService),
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

	// notifications
	mux.HandleFunc("/api/v1/notifications", middleware.ChainMiddleware(notificationHandler.GetNotifications,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.Log()))
	// comments
	mux.HandleFunc("/api/v1/flows/{flow_id}/comments", middleware.ChainMiddleware(commentHandler.GetComments,
		middleware.AuthMiddleware(jwtManager, false),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	mux.HandleFunc("POST /api/v1/flows/{flow_id}/comments", middleware.ChainMiddleware(commentHandler.AddComment,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/flows/{flow_id}/comments", middleware.ChainMiddleware(commentHandler.GetComments,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	mux.HandleFunc("DELETE /api/v1/flows/{flow_id}/comments/{comment_id}", middleware.ChainMiddleware(commentHandler.DeleteComment,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedDeleteOptions),
		middleware.Log()))

	mux.HandleFunc("OPTIONS /api/v1/flows/{flow_id}/comments/{comment_id}", middleware.ChainMiddleware(commentHandler.DeleteComment,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

	mux.HandleFunc("POST /api/v1/flows/{flow_id}/comments/{comment_id}/like", middleware.ChainMiddleware(commentHandler.LikeComment,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedPostOptions),
		middleware.Log()))
	
	mux.HandleFunc("OPTIONS /api/v1/flows/{flow_id}/comments/{comment_id}/like", middleware.ChainMiddleware(commentHandler.LikeComment,
		middleware.AuthMiddleware(jwtManager, true),
		middleware.CSRFMiddleware(),
		middleware.CorsMiddleware(config, allowedGetOptions),
		middleware.Log()))

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
