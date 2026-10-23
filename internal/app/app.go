package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"shieldgate/internal/config"
	"shieldgate/internal/controller/restapi/middleware"
	v1 "shieldgate/internal/controller/restapi/v1"
	"shieldgate/internal/repository/cache"
	repository "shieldgate/internal/repository/postgres"
	"shieldgate/internal/service"
	"shieldgate/pkg/jwt"
	"shieldgate/pkg/limiter"
	"shieldgate/pkg/metrics"
	"shieldgate/pkg/postgres"
	"shieldgate/pkg/proxy"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type Closer interface {
	Close()
}

type App struct {
	cfg        config.Config
	httpServer *http.Server
	db         *postgres.DB
	rdb        *redis.Client
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	pgConn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := postgres.New(ctx, pgConn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisHost,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	userRepo := repository.NewUserRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)

	cachedServiceRepo := cache.NewCachedServiceRepository(serviceRepo, rdb, 10*time.Minute)
	cachedAPIKeyRepo := cache.NewCachedAPIKeyRepository(apiKeyRepo, rdb, 10*time.Minute)

	collectorManager := metrics.NewCollectorManager()

	jwtDuration := 24 * time.Hour
	jwtManager := jwt.New(cfg.JWTSecret, jwtDuration)

	redisLimiter := limiter.NewRedisLimiter(rdb)

	proxyManager := proxy.NewProxyManager()

	authService := service.NewAuthService(userRepo, jwtManager)
	routeService := service.NewRouteService(serviceRepo, apiKeyRepo)
	gatewayService := service.NewGatewayService(cachedServiceRepo, cachedAPIKeyRepo, redisLimiter)

	authCtrl := v1.NewAuthController(authService)
	serviceCtrl := v1.NewServiceController(routeService)
	apiKeyCtrl := v1.NewAPIKeyController(routeService)
	gatewayCtrl := v1.NewGatewayController(gatewayService, proxyManager)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("POST /api/v1/auth/register", middleware.Logger(http.HandlerFunc(authCtrl.Register)))
	mux.Handle("POST /api/v1/auth/login", middleware.Logger(http.HandlerFunc(authCtrl.Login)))

	mux.Handle("POST /api/v1/services", middleware.Logger(middleware.Auth(jwtManager, http.HandlerFunc(serviceCtrl.Create))))
	mux.Handle("GET /api/v1/services", middleware.Logger(middleware.Auth(jwtManager, http.HandlerFunc(serviceCtrl.GetAll))))
	mux.Handle("POST /api/v1/services/{id}/keys", middleware.Logger(middleware.Auth(jwtManager, http.HandlerFunc(apiKeyCtrl.Generate))))

	mux.Handle("/", middleware.Logger(middleware.Metrics(collectorManager, gatewayCtrl)))
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return &App{
		cfg:        cfg,
		httpServer: server,
		db:         db,
		rdb:        rdb,
	}, nil
}

func (a *App) Run() error {
	go func() {
		slog.Info("Starting server on port " + a.cfg.AppPort)
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start server: " + err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	a.db.Close()
	if err := a.rdb.Close(); err != nil {
		slog.Error("failed to close redis", "error", err)
	}
	slog.Info("Server stopped cleanly")
	return nil
}
