package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/config"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/service"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger *zap.Logger
	authService *service.AuthService
}

type registerReq struct {
	Name string `json:"name" binding:"required,min=3"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type forgotPasswordReq struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordReq struct {
	Email       string `json:"email" binding:"required,email"`
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func NewAuthHandler(logger *zap.Logger, authService *service.AuthService) (*AuthHandler) {
	return &AuthHandler{
		logger: logger,
		authService: authService,
	}
}

// RegisterHandler godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body registerReq true "Registration details"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /register [post]
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	
	var req registerReq
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid register request", 
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// Forward req to AuthService
	if err := h.authService.RegisterService(req.Name, req.Email, req.Password); err != nil {
		h.logger.Error("register failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		
		c.JSON(http.StatusConflict, gin.H{
			"error": "unable to register user",
		})
		
		return 
	}
	
	h.logger.Info("user registered successfully",
		zap.String("user", req.Name),
		zap.String("email", req.Email),
	)
	
	// Response
	c.JSON(http.StatusCreated, gin.H{
    	"message": "user registered successfully",
	})
}

// LoginHandler godoc
// @Summary Login a user
// @Description Authenticate user and return token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body loginReq true "Login credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	
	var req loginReq
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request", 
			zap.Error(err),
		)
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
		})
		return
	}
	
	token, err := h.authService.LoginService(req.Email, req.Password)
	if err != nil {
		h.logger.Warn("login failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error" : "invalid credentials",
		})
		
		return
	}
	
	// Set cookie
	isProd := config.GetEnv("ENV", "development") == "production"

	c.SetCookie(
	    "token",
	    token,
	    3600*24,
	    "/",
	    "",
	    isProd,
	    true,
	)
	
	// Temporary message
	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
	})
}

// ForgotPasswordHandler godoc
// @Summary Request password reset
// @Description Send a reset token to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body forgotPasswordReq true "Email for reset"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /forgot-password [post]
func (h *AuthHandler) ForgotPasswordHandler(c *gin.Context) {
	var req forgotPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	if err := h.authService.RequestPasswordReset(req.Email); err != nil {
		h.logger.Error("failed to request password reset", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset token has been generated"})
}

// ResetPasswordHandler godoc
// @Summary Reset user password
// @Description Reset password using a valid token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body resetPasswordReq true "New password details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /reset-password [post]
func (h *AuthHandler) ResetPasswordHandler(c *gin.Context) {
	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ResetPassword(req.Email, req.Token, req.NewPassword); err != nil {
		h.logger.Warn("failed to reset password", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}