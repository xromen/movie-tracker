package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/xromen/movietracker/internal/config"
	"github.com/xromen/movietracker/internal/handler"
	"github.com/xromen/movietracker/internal/platform/cache"
	"github.com/xromen/movietracker/internal/platform/database"
	"github.com/xromen/movietracker/internal/platform/hasher"
	"github.com/xromen/movietracker/internal/platform/jwt"
	"github.com/xromen/movietracker/internal/platform/refreshtoken"
	"github.com/xromen/movietracker/internal/platform/tmdb"
	"github.com/xromen/movietracker/internal/repository"
	"github.com/xromen/movietracker/internal/service"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Database.PoolConfig())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	m, err := migrate.New(
		"file:///migrations", // путь внутри Docker-образа
		cfg.Database.DSN(),
	)
	if err != nil {
		slog.Error("failed to initialize migrations", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("database migrations applied")

	bcryptHasher := hasher.NewBcrypt(bcrypt.DefaultCost)

	jwtManager := jwt.NewManager(jwt.Config{
		Secret:         cfg.JWT.Secret,
		AccessTokenTTL: cfg.JWT.AccessTokenTTL,
	})

	refreshTokenManager := refreshtoken.NewManager(refreshtoken.Config{
		RefreshTokenTTL: cfg.RefreshToken.RefreshTokenTTL,
	})

	redisCache, err := cache.NewRedisCache(cache.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Disabled: cfg.Redis.Disabled,
	})
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := redisCache.Close(); err != nil {
			slog.Warn("failed to close redis client", "error", err)
		}
	}()

	healthHandler := handler.NewHealthHandler(pool, redisCache, "1.0.0")

	userRepo := repository.NewUserRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	userService := service.NewUserService(userRepo, refreshTokenRepo, bcryptHasher, jwtManager, refreshTokenManager, logger)
	userHandler := handler.NewUserHandler(userService, logger)

	mediaRepo := repository.NewMediaRepository(pool)

	tmdbClient, err := tmdb.NewClient(tmdb.Config{
		BaseURL:       cfg.TMDB.BaseURL,
		BearerToken:   cfg.TMDB.BearerToken,
		ImagesBaseURL: cfg.TMDB.ImagesBaseURL,
		Timeout:       cfg.TMDB.Timeout,
	})
	if err != nil {
		slog.Error("failed to connect to TMDB", "error", err)
		os.Exit(1)
	}

	movieService := service.NewMovieService(mediaRepo, tmdbClient, redisCache, logger)
	movieHandler := handler.NewMovieHandler(movieService, logger)

	tvShowService := service.NewTVShowService(mediaRepo, tmdbClient, redisCache, logger)
	tvShowHandler := handler.NewTVShowHandler(tvShowService, logger)

	watchListService := service.NewWatchListService(mediaRepo, tmdbClient, redisCache, logger)
	watchListHandler := handler.NewWatchListHandler(watchListService, logger)

	collectionService := service.NewCollectionService(tmdbClient, redisCache, logger)
	collectionHandler := handler.NewCollectionHandler(collectionService, logger)

	searchService := service.NewSearchService(tmdbClient, redisCache, logger)
	searchHandler := handler.NewSearchHandler(searchService, logger)

	router := setupRouter(
		cfg,
		jwtManager,
		userService,
		healthHandler,
		userHandler,
		movieHandler,
		tvShowHandler,
		watchListHandler,
		collectionHandler,
		searchHandler,
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	g, gCtx := errgroup.WithContext(ctx)

	// HTTP сервер.
	g.Go(func() error {
		slog.Info("server starting", "port", cfg.HTTP.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	// Горутина которая следит за отменой контекста
	// и инициирует shutdown HTTP сервера.
	g.Go(func() error {
		// Ждём сигнала завершения (Ctrl+C или ошибки другой горутины).
		<-gCtx.Done()
		slog.Info("shutting down...")

		// Даём 30 секунд на завершение текущих запросов.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}

		slog.Info("server stopped")
		return nil
	})

	// Ждём завершения всех горутин.
	if err := g.Wait(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}

	slog.Info("application stopped gracefully")
}

func setupRouter(
	cfg *config.Config,
	jwtManager jwt.Manager,
	userService service.UserService,
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
	movieHandler *handler.MovieHandler,
	tvShowHandler *handler.TVShowHandler,
	watchListHandler *handler.WatchListHandler,
	collectionHandler *handler.CollectionHandler,
	searchHandler *handler.SearchHandler,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(handler.StructuredLogger(logger()))
	//router.Use(structuredLogger(logger()))

	router.Use(cors.New(cors.Config{
		// Список разрешённых origins.
		// В development можно ["*"], в production — конкретные домены.
		AllowOrigins: []string{
			"*",
			"http://localhost:3000", // React dev server
			"http://localhost:5173", // Vite dev server
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID", // чтобы фронт видел trace ID
		},
		// Разрешаем куки и Authorization заголовок.
		AllowCredentials: true,
		// Сколько секунд браузер кэширует preflight ответ.
		// 12 часов — не будет слать OPTIONS перед каждым запросом.
		MaxAge: 12 * time.Hour,
	}))

	api := router.Group("/api")

	adminRole := "admin"
	health := api.Group("/health")
	{
		health.GET("/live", healthHandler.Live)
		health.GET("/ready", handler.AuthMiddleware(jwtManager, userService, &adminRole), healthHandler.Ready)
	}

	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/logout", userHandler.Logout)
		auth.POST("/register", userHandler.Register)
		auth.POST("/refresh", userHandler.Refresh)
	}

	protected := v1.Group("/")
	protected.Use(handler.AuthMiddleware(jwtManager, userService, nil))

	{
		unprotectedMovies := v1.Group("/movie")
		{
			unprotectedMovies.GET("/:id", movieHandler.GetMovieDetail)
			unprotectedMovies.GET("/:id/recommendations", movieHandler.GetMovieRecommendations)
			unprotectedMovies.GET("/search", movieHandler.Search)
			unprotectedMovies.GET("/now-playing", movieHandler.NowPlaying)
			unprotectedMovies.GET("/popular", movieHandler.Popular)
			unprotectedMovies.GET("/top-rated", movieHandler.TopRated)
			unprotectedMovies.GET("/upcoming", movieHandler.Upcoming)
		}

		//protectedMovies := protected.Group("/movie")
		//{
		//	//protectedMovies.GET("", movieHandler.List)
		//	//protectedMovies.POST("/list", movieHandler.AddToList)
		//}
	}

	{
		unprotectedTVShows := v1.Group("/tv")
		{
			unprotectedTVShows.GET("/:id", tvShowHandler.GetTVShowDetail)
			unprotectedTVShows.GET("/:id/season/:season_number", handler.OptionalAuthMiddleware(jwtManager), tvShowHandler.GetSeasonEpisodes)
			unprotectedTVShows.GET("/:id/recommendations", tvShowHandler.GetRecommendations)
			unprotectedTVShows.GET("/search", tvShowHandler.Search)
			unprotectedTVShows.GET("/airing-today", tvShowHandler.GetAiringToday)
			unprotectedTVShows.GET("/popular", tvShowHandler.GetPopular)
			unprotectedTVShows.GET("/top-rated", tvShowHandler.GetTopRated)
			unprotectedTVShows.GET("/on-the-air", tvShowHandler.GetOnTheAir)
		}

		protectedTVShows := protected.Group("/tv")
		{
			//protectedTVShows.GET("", tvShowHandler.GetList)
			//protectedTVShows.POST("/list", tvShowHandler.AddToList)

			protectedTVShows.PUT("/:id/season/:season_number/episode/:episode_number/watched", tvShowHandler.MarkEpisodeWatched)
			protectedTVShows.DELETE("/:id/season/:season_number/episode/:episode_number/watched", tvShowHandler.UnmarkEpisodeWatched)

			protectedTVShows.PUT("/:id/season/:season_number/watched", tvShowHandler.MarkSeasonWatched)
			protectedTVShows.DELETE("/:id/season/:season_number/watched", tvShowHandler.UnmarkSeasonWatched)
		}
	}

	{
		protectedWatchList := protected.Group("/watch-list")
		{
			protectedWatchList.GET("", watchListHandler.GetUserWatchList)
			protectedWatchList.DELETE("/status", watchListHandler.DeleteUserStatus)
			protectedWatchList.GET("/status", watchListHandler.GetMediaUserStatus)
			protectedWatchList.POST("/status", watchListHandler.SetStatus)
		}
	}

	{
		unprotectedCollections := v1.Group("/collections")
		{
			unprotectedCollections.GET("/:id", collectionHandler.GetDetails)
		}
	}

	{
		search := v1.Group("/search")
		{
			search.GET("/multi", searchHandler.SearchMulti)
		}
	}

	return router
}

func structuredLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", statusCode,
			"latency_ms", latency.Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}

func logger() *slog.Logger {
	return slog.Default()
}
