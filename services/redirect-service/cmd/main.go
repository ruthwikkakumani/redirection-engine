package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/ruthwikkakumani/url-shortener/pkg/logger"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/cache"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/db"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/kafka"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/middleware"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/routes"
	"go.uber.org/zap"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment.")
	}
}

func newServer(logger *zap.Logger, pool *pgxpool.Pool, redisClient *cache.RedisClient, producer *kafka.Producer) *gin.Engine {
	server := gin.New()

	server.Use(gin.Recovery())
	server.Use(middleware.ZapMiddleware(logger))

	routes.RegisterRoutes(server, logger, pool, redisClient, producer)

	return server
}

func startServer(server *gin.Engine, logger *zap.Logger) {
	port := config.GetEnv("PORT", "8083")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      server,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("Server starting",
			zap.String("port", port),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start",
				zap.Error(err),
			)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("shutting down server....")

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced Shutdown",
			zap.Error(err),
		)
	}

	logger.Info("server exited cleanly")
}

func main() {

	// Load env
	LoadEnv()

	env := config.GetEnv("ENV", "development")

	// Initialize Logger
	logger, err := logger.InitLogger(env)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = logger.Sync()
	}()

	// Initialize Redis
	redisClient := cache.NewRedisClient(logger)

	if err := redisClient.Init(context.Background()); err != nil {
		logger.Fatal("failed to initialize redis",
			zap.Error(err),
		)
	}

	defer redisClient.Close()

	// Initialize DB
	dbService := db.NewDB(logger)
	if err := dbService.InitDB(context.Background()); err != nil {
		logger.Fatal("failed to initialize db",
			zap.Error(err),
		)
	}
	defer dbService.Close()

	pool, err := dbService.GetPool()
	if err != nil {
		logger.Error("db not initialized",
			zap.Error(err),
		)
	}

	// Initialize Kafka Producer (optional — degrades gracefully if not configured)
	var producer *kafka.Producer
	kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "")
	if kafkaBrokers != "" {
		brokers := strings.Split(kafkaBrokers, ",")
		producer, err = kafka.NewProducer(brokers, logger)
		if err != nil {
			logger.Warn("kafka: failed to init producer (" + err.Error() + ") — analytics disabled")
			producer = nil
		} else {
			defer producer.Close()
			logger.Info("kafka: producer initialised", zap.Strings("brokers", brokers))
		}
	} else {
		logger.Warn("KAFKA_BROKERS not set — analytics events will not be published")
	}

	// server setup
	server := newServer(logger, pool, redisClient, producer)

	// start server
	startServer(server, logger)
}
