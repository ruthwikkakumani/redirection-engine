package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/kafka"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/service"
	"go.uber.org/zap"
)

type UrlHandler struct {
	logger   *zap.Logger
	service  *service.UrlService
	producer *kafka.Producer // nil if Kafka is disabled
}

// NewUrlHandler initialises the handler.  producer may be nil (Kafka disabled).
func NewUrlHandler(logger *zap.Logger, svc *service.UrlService, producer *kafka.Producer) *UrlHandler {
	return &UrlHandler{
		logger:   logger,
		service:  svc,
		producer: producer,
	}
}

func (h *UrlHandler) RedirectURL(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
		return
	}

	originalURL, err := h.service.GetOriginalURL(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("URL not found or expired", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "url not found or expired"})
		return
	}

	// Fire-and-forget: publish the click event to Kafka before redirecting.
	if h.producer != nil {
		h.producer.PublishClick(kafka.ClickEvent{
			ShortCode:   code,
			OriginalURL: originalURL,
			IP:          c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Referer:     c.Request.Referer(),
			ClickedAt:   time.Now().UTC(),
		})
	}

	c.Redirect(http.StatusFound, originalURL)
}