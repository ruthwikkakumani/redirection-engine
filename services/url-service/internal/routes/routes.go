package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/cache"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/config"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/handler"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/middleware"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/repository"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/service"
	"go.uber.org/zap"
	_ "github.com/ruthwikkakumani/redirection-engine/services/url-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, db *pgxpool.Pool, cache *cache.RedisClient) {

	repo := repository.NewUrlRepo(logger, db)
	urlService := service.NewUrlService(logger, repo, cache)
	urlHandler := handler.NewUrlHandler(logger, urlService)
	
	// Swagger documentation - Only exposed in non-production environments
	if config.GetEnv("ENV", "development") != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Protected routes
	urls := r.Group("/")
	protected := urls.Group("")
	protected.Use(middleware.AuthMiddleware())

	// Shorten Original URL
	protected.POST("", urlHandler.ShortenURL)

	// List registered urls
	protected.GET("/urls", urlHandler.ListURLS)

	// Update shorten URL
	protected.PATCH("/:shortCode", urlHandler.UpdateURL)

	// Delete URL
	protected.DELETE("/:shortCode", urlHandler.DeleteURL)

}
