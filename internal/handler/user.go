package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=$GOFILE -destination=user_mock_test.go -package=handler UserService

type UserService interface {
	CreateUser(ctx context.Context, user user.User, fingerprint string) (string, string, error)
	Authenticate(ctx context.Context, email, password, fingerprint string) (string, string, error)
	Refresh(ctx context.Context, refreshToken, fingerprint string) (string, string, error)
	Logout(ctx context.Context, refreshToken, fingerprint string) error
	GetUserByID(ctx context.Context, id uint) (entity.User, error)
}

type userHandler struct {
	userService UserService
	refreshLife int
	basePath    string
	domain      string
}

func NewUserHandler(
	userService UserService,
	refreshLife time.Duration,
	basePath string,
	domain string,
) userHandler {
	return userHandler{
		userService: userService,
		refreshLife: int(refreshLife.Seconds()),
		basePath:    basePath,
		domain:      domain,
	}
}

func (h *userHandler) Registration(c *gin.Context) {
	var regUserDTO dto.RegistrationUserReq
	if err := c.ShouldBindJSON(&regUserDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid user data",
			"details": err.Error(),
		})
		return
	}

	accessToken, refreshToken, err := h.userService.CreateUser(
		c.Request.Context(),
		regUserDTO.ConvertToSvc(),
		regUserDTO.Fingerprint,
	)
	if err != nil {
		c.Error(err)
		return
	}
	response := dto.RegistrationUserResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	setRefreshCookie(c, refreshToken, h.basePath, h.domain, h.refreshLife)

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) Login(c *gin.Context) {
	var loginUserDTO dto.LoginUserReq
	if err := c.ShouldBindJSON(&loginUserDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid user data",
			"details": err.Error(),
		})
		return
	}

	accessToken, refreshToken, err := h.userService.Authenticate(
		c.Request.Context(),
		loginUserDTO.Email,
		loginUserDTO.Password,
		loginUserDTO.Fingerprint,
	)

	if err != nil {
		c.Error(err)
		return
	}

	response := dto.LoginUserResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	setRefreshCookie(c, refreshToken, h.basePath, h.domain, h.refreshLife)

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) Refresh(c *gin.Context) {
	var refreshDTO dto.RefreshReq
	if err := c.ShouldBindJSON(&refreshDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid refresh data",
			"details": err.Error(),
		})
		return
	}

	if refreshDTO.RefreshToken == "" {
		refresh, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    apperror.BadRequest,
				"message": "Invalid refresh data",
				"details": err.Error(),
			})
			return
		}
		refreshDTO.RefreshToken = refresh
	}

	accessToken, refreshToken, err := h.userService.Refresh(
		c.Request.Context(),
		refreshDTO.RefreshToken,
		refreshDTO.Fingerprint,
	)
	if err != nil {
		c.Error(err)
		return
	}

	response := dto.RefreshResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	setRefreshCookie(c, refreshToken, h.basePath, h.domain, h.refreshLife)

	c.JSON(http.StatusOK, response)
}

func (h *userHandler) Logout(c *gin.Context) {
	var logoutDTO dto.LogoutReq
	if err := c.ShouldBindJSON(&logoutDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid refresh data",
			"details": err.Error(),
		})
		return
	}

	if logoutDTO.RefreshToken == "" {
		refresh, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    apperror.BadRequest,
				"message": "Invalid refresh data",
				"details": err.Error(),
			})
			return
		}
		logoutDTO.RefreshToken = refresh
	}

	if err := h.userService.Logout(
		c.Request.Context(),
		logoutDTO.RefreshToken,
		logoutDTO.Fingerprint,
	); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func setRefreshCookie(c *gin.Context, refreshToken, basePath, domain string, maxAge int) {
	jwtCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     basePath + "/auth",
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
	}

	c.SetCookie(
		jwtCookie.Name,
		jwtCookie.Value,
		jwtCookie.MaxAge,
		jwtCookie.Path,
		jwtCookie.Domain,
		jwtCookie.Secure,
		jwtCookie.HttpOnly,
	)

	c.SetSameSite(http.SameSiteStrictMode)
}
