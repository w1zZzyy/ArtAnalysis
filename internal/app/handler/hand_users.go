package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// simple singleton for demo auth
var creatorUserID uint = 1

func currentUserID() uint { return creatorUserID }

// Register регистрирует нового пользователя
// @Summary Регистрация пользователя
// @Description Создает нового пользователя с указанными логином и паролем
// @Tags Users
// @Accept json
// @Produce json
// @Param request body DTO_Req_UserReg true "Данные для регистрации"
// @Success 200 {object} DTO_Resp_User "Зарегистрированный пользователь"
// @Failure 400 {object} string "Invalid input data"
// @Failure 500 {object} string "Internal server error"
// @Router /users [post]
func (h *Handler) Register(ctx *gin.Context) {
	var req DTO_Req_UserReg
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := h.Repository.RegisterUser(req.Login, string(hashedPassword)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, DTO_Resp_User{Login: req.Login})
}

// Login выполняет аутентификацию пользователя
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя и возвращает JWT токен
// @Tags Users
// @Accept json
// @Produce json
// @Param request body DTO_Req_UserReg true "Учетные данные"
// @Success 200 {object} DTO_Resp_TokenLogin "Токен и данные пользователя"
// @Failure 400 {object} string "Invalid input data"
// @Failure 401 {object} string "Invalid credentials"
// @Failure 500 {object} string "Internal server error"
// @Router /login [post]
func (h *Handler) Login(ctx *gin.Context) {
	var req DTO_Req_UserReg
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByUsername(req.Login)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		customErr := errors.New(user.Password)
		h.errorHandler(ctx, http.StatusUnauthorized, customErr)
		return
	}

	claims := model.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.JWTConfig.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:      user.ID_user,
		IsModerator: user.IsModerator,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.JWTConfig.Secret))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	response := DTO_Resp_TokenLogin{
		Token: tokenString,
		User: DTO_User{
			ID_user: user.ID_user,
			Login:   user.Login,
			IsAdmin: user.IsModerator,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// ApiMe возвращает информацию о текущем пользователе
// @Summary Получить текущего пользователя
// @Description Возвращает данные аутентифицированного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DTO_Resp_User "Данные пользователя"
// @Failure 401 {object} string "Unauthorized"
// @Failure 500 {object} string "Internal server error"
// @Router /api/users/me [get]
func (h *Handler) ApiMe(ctx *gin.Context) {
	user, err := h.Repository.GetUserByID(currentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, DTO_Resp_User{Login: user.Login})
}

// ApiUpdateMe обновляет данные текущего пользователя
// @Summary Обновить данные пользователя
// @Description Обновляет пароль текущего аутентифицированного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DTO_Req_UserUpd true "Новые данные пользователя"
// @Success 200 {object} DTO_Resp_User "Обновленные данные пользователя"
// @Failure 400 {object} string "Invalid input data"
// @Failure 401 {object} string "Unauthorized"
// @Failure 500 {object} string "Internal server error"
// @Router /api/users/me [put]
func (h *Handler) ApiUpdateMe(ctx *gin.Context) {
	var req DTO_Req_UserUpd
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	updated, err := h.Repository.UpdateUser(currentUserID(), req.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, DTO_Resp_User{Login: updated.Login})
}

// Logout выполняет выход пользователя
// @Summary Выход из системы
// @Description Добавляет JWT токен в черный список и выполняет деавторизацию
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "Сообщение об успешном выходе" {"message": "Деавторизация прошла успешно"}
// @Failure 400 {object} string "Invalid authorization header"
// @Failure 500 {object} string "Internal server error"
// @Router /api/auth/logout [post]
func (h *Handler) Logout(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid header"))
		return
	}
	tokenStr := authHeader[len("Bearer "):]

	err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), tokenStr, h.JWTConfig.ExpiresIn)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Деавторизация прошла успешно",
	})
}

// getUserIDFromContext извлекает ID пользователя из контекста
// Вспомогательная функция для внутреннего использования, не экспортируется в Swagger
func getUserIDFromContext(ctx *gin.Context) (uint, error) {
	value, exists := ctx.Get(userCtx)
	if !exists {
		return 0, errors.New("user ID not found in context")
	}

	userID, ok := value.(uint)
	if !ok {
		return 0, errors.New("invalid user ID type in context")
	}

	return userID, nil
}

/*
// ApiLogout выполняет выход пользователя (альтернативная версия)
// @Summary Выход из системы (альтернативная версия)
// @Description Альтернативная реализация выхода из системы
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DTO_Resp_UserLogout "Статус выхода"
// @Router /api/logout [post]
func (h *Handler) ApiLogout(ctx *gin.Context) {
    ctx.JSON(http.StatusOK, DTO_Resp_UserLogout{Logout: true})
}
*/
