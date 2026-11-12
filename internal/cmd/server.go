package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/repurpose-ai/re-purpose-be/config"
	controllerHttp "github.com/repurpose-ai/re-purpose-be/internal/controller/http"
	"github.com/repurpose-ai/re-purpose-be/internal/controller/middleware"
	"github.com/repurpose-ai/re-purpose-be/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var runServer = &cobra.Command{
	Use:   "server",
	Short: "run HTTP server",
	Long:  "Start the RePurpose.ai HTTP API server.",
	Run:   server,
}

func init() {
	RootCmd.AddCommand(runServer)
}

func server(_ *cobra.Command, _ []string) {
	startedAt := time.Now()

	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	db := connectDatabase()
	if db != nil {
		defer closeDatabase(db)
	}

	redisClient := connectRedis()
	if redisClient != nil {
		defer func() { _ = redisClient.Close() }()
	}

	container := InitContainer(db, redisClient, startedAt)

	// gin.New() keeps gin's default logger/recovery out of the way so all
	// request logs go through internal/logger with the same structure.
	app := gin.New()
	app.UseRawPath = true
	app.UnescapePathValues = true
	app.RemoveExtraSlash = true

	httpMiddleware := middleware.NewMiddleware()
	app.Use(httpMiddleware.RequestID())
	app.Use(httpMiddleware.Logger())
	app.Use(httpMiddleware.Recovery())
	app.Use(cors.New(buildCORSConfig()))

	controllerHttp.RegisterDocs(app)
	controllerHttp.RouteService(
		&app.RouterGroup,
		httpMiddleware,
		container.ContentService,
		container.SystemService,
		container.UserService,
		container.SponsorService,
	)

	httpServer := &http.Server{
		Addr:              ":" + config.HttpPort(),
		Handler:           app,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logrus.WithFields(logrus.Fields{
			"port": config.HttpPort(),
			"db":   config.DatabaseDriver(),
			"llm":  container.EngineName,
			"mode": config.EnvironmentMode(),
		}).Info("RePurpose.ai API listening 🚀")

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Error("http server error")
			stop()
		}
	}()

	<-ctx.Done()
	logrus.Warn("shutdown signal received, draining connections...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("graceful shutdown failed")
	}

	logrus.Info("server stopped ✅")
}

func connectDatabase() *gorm.DB {
	driver := strings.ToLower(config.DatabaseDriver())
	if driver != "postgres" && driver != "postgresql" {
		logrus.Infof("db.driver=%q is not postgres; using in-memory repositories", driver)
		return nil
	}

	db, err := database.InitializePostgresConnection()
	if err != nil {
		logrus.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	return db
}

func connectRedis() *redis.Client {
	if !config.EnableCaching() {
		logrus.Info("caching disabled; redis is not connected")
		return nil
	}

	client, err := database.InitializeRedis(config.RedisCacheHost())
	if err != nil {
		logrus.Fatalf("failed to connect to Redis [%s]: %v", config.RedisCacheHost(), err)
	}
	return client
}

func buildCORSConfig() cors.Config {
	origins := config.CORSAllowedOrigins()

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	for _, origin := range origins {
		if origin == "*" {
			corsConfig.AllowAllOrigins = true
			corsConfig.AllowCredentials = false
			return corsConfig
		}
	}

	corsConfig.AllowOrigins = origins
	return corsConfig
}

func closeDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		logrus.WithError(err).Error("failed to get sql db")
		return
	}
	if err := sqlDB.Close(); err != nil {
		logrus.WithError(err).Error("failed to close sql db")
	}
}
