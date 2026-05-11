package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/cache"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/handler"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/kafka"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/service"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *pgxpool.Pool, redisClient *cache.RedisClient, producer *kafka.Producer) {

	repo := repository.NewUrlRepo(logger, pool)
	urlService := service.NewUrlService(logger, repo, redisClient)
	urlHandler := handler.NewUrlHandler(logger, urlService, producer)

	r.GET("/r/:code", urlHandler.RedirectURL)
}