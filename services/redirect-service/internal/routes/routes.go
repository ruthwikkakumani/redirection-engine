package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/cache"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/config"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/handler"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/kafka"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/repository"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/service"
	"go.uber.org/zap"
	_ "github.com/ruthwikkakumani/redirection-engine/services/redirect-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *pgxpool.Pool, redisClient *cache.RedisClient, producer *kafka.Producer) {

	repo := repository.NewUrlRepo(logger, pool)
	urlService := service.NewUrlService(logger, repo, redisClient)
	urlHandler := handler.NewUrlHandler(logger, urlService, producer)
	
	// Swagger documentation - Only exposed in non-production environments
	if config.GetEnv("ENV", "development") != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	r.GET("/r/:code", urlHandler.RedirectURL)
}