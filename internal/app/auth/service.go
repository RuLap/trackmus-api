package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/RuLap/trackmus-api/internal/pkg/events"
	"github.com/RuLap/trackmus-api/internal/pkg/jwthelper"
	"github.com/RuLap/trackmus-api/internal/pkg/rabbitmq"
	"github.com/RuLap/trackmus-api/internal/pkg/redis"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	GoogleAuth(ctx context.Context, req GoogleAuthRequest) (*AuthResponse, error)
	GenerateGoogleOAuthURL() (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID string) error

	SendConfirmationLink(ctx context.Context, req *SendConfirmationEmailRequest) error
	ConfirmEmail(ctx context.Context, token string, currentUserID string) error
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type service struct {
	log          *slog.Logger
	jwtHelper    *jwthelper.JWTHelper
	googleConfig *GoogleOAuthConfig
	redis        *redis.Service
	rabbitmq     *rabbitmq.Service
	repo         Repository
}

func NewService(
	log *slog.Logger,
	jwtHelper *jwthelper.JWTHelper,
	googleConfig *GoogleOAuthConfig,
	redis *redis.Service,
	rabbitmq *rabbitmq.Service,
	repo Repository,
) Service {
	return &service{
		log:          log,
		jwtHelper:    jwtHelper,
		googleConfig: googleConfig,
		redis:        redis,
		rabbitmq:     rabbitmq,
		repo:         repo,
	}
}

func (s *service) SendConfirmationLink(ctx context.Context, req *SendConfirmationEmailRequest) error {
	if req.UserID == "" || req.Email == "" {
		return fmt.Errorf("userID и email обязательны")
	}

	token := uuid.New().String()

	err := s.redis.StoreEmailConfirmation(ctx, req.UserID, req.Email, token)
	if err != nil {
		s.log.Error("failed to store tokens in redis", "error", err, "user_id", req.UserID)
		return fmt.Errorf("не удалось сохранить токен")
	}

	confirmationURL := fmt.Sprintf("https://trakcmus.ru/confirm?token=%s", token)

	if s.rabbitmq != nil {
		event := events.EmailEvent{
			To:       req.Email,
			Template: "email_confirmation",
			Subject:  "Подтвердите ваш email",
			Data: map[string]interface{}{
				"confirmation_url": confirmationURL,
				"user_email":       req.Email,
			},
		}

		if err := s.rabbitmq.PublishEmail(event); err != nil {
			s.log.Error("failed to publish email event", "error", err)
		}
	} else {
		s.log.Warn("event service not available - email not sent")
	}

	s.log.Info("confirmation link sent", "email", req.Email, "user_id", req.UserID)
	return nil
}

func (s *service) ConfirmEmail(ctx context.Context, token string, currentUserID string) error {
	if token == "" {
		return fmt.Errorf("токен обязателен")
	}

	tokenUserID, err := s.redis.GetEmailConfirmationUserID(ctx, token)
	if err != nil {
		s.log.Warn("invalid or expired confirmation token", "token", token, "error", err)
		return fmt.Errorf("неверная или устаревшая ссылка подтверждения")
	}

	if tokenUserID != currentUserID {
		s.log.Warn("security alert: token user mismatch",
			"token_user", tokenUserID,
			"current_user", currentUserID,
			"token", token,
		)
		return fmt.Errorf("токен не принадлежит текущему пользователю")
	}

	if err := s.repo.MakeEmailConfirmed(ctx, currentUserID); err != nil {
		s.log.Error("failed to confirm email in database", "error", err, "user_id", currentUserID)
		return fmt.Errorf("не удалось подтвердить email")
	}

	err = s.redis.DeleteEmailConfirmation(ctx, currentUserID, token)
	if err != nil {
		s.log.Warn("failed to delete used tokens", "token", token, "error", err)
	}

	s.log.Info("email confirmed successfully", "user_id", currentUserID)
	return nil
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", "error", err)
		return nil, fmt.Errorf("произошла ошибка")
	}

	hashedPasswordStr := string(hashedPassword)

	user := RegisterRequestToUser(&req, hashedPasswordStr)

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.log.Error("failed to create user", "error", err, "email", user.Email)
		return nil, err
	}

	tokenPair, err := s.jwtHelper.GenerateTokenPair(*userID, req.Email)
	if err != nil {
		s.log.Error("failed to generate JWT tokens", "error", err)
		return nil, fmt.Errorf("произошла ошибка")
	}

	err = s.storeRefreshToken(ctx, *userID, tokenPair.RefreshToken)
	if err != nil {
		s.log.Error("failed to store refresh token", "error", err, "user_id", *userID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	s.log.Info("user registered successfully", "user_id", *userID, "email", req.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       *userID,
		Email:        req.Email,
	}, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetByEmailProvider(ctx, req.Email, LocalProvider)
	if err != nil {
		s.log.Warn("user not found", "email", req.Email)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	passwordHash, err := s.repo.GetPasswordHashByEmail(ctx, req.Email)
	if err != nil {
		s.log.Warn("failed to get password hash", "email", req.Email, "error", err)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	if req.Password == "" {
		s.log.Error("user entered empty password", "email", req.Email)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*passwordHash), []byte(req.Password)); err != nil {
		s.log.Error("user entered invalid password", "email", req.Email)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	tokenPair, err := s.jwtHelper.GenerateTokenPair(user.ID.String(), user.Email)
	if err != nil {
		s.log.Error("failed to generate JWT tokens", "error", err)
		return nil, fmt.Errorf("произошла ошибка")
	}

	err = s.storeRefreshToken(ctx, user.ID.String(), tokenPair.RefreshToken)
	if err != nil {
		s.log.Error("failed to store refresh token", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	s.log.Info("user logged in successfully", "user_id", user.ID, "email", req.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       user.ID.String(),
		Email:        user.Email,
	}, nil
}

func (s *service) GoogleAuth(ctx context.Context, req GoogleAuthRequest) (*AuthResponse, error) {
	token, err := s.exchangeCodeForToken(req.Code)
	if err != nil {
		s.log.Error("failed to exchange code for token", "error", err)
		return nil, fmt.Errorf("ошибка авторизации через Google")
	}

	userInfo, err := s.getGoogleUserInfo(token)
	if err != nil {
		s.log.Error("failed to get user info from Google", "error", err)
		return nil, fmt.Errorf("ошибка получения данных от Google")
	}

	user, err := s.repo.GetByEmailProvider(ctx, userInfo.Email, GoogleProvider)
	if err != nil {
		s.log.Info("creating new google user", "email", userInfo.Email)
	}

	providerID := userInfo.ID
	user = &User{
		Email:          userInfo.Email,
		Provider:       GoogleProvider,
		ProviderID:     &providerID,
		EmailConfirmed: true,
	}

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.log.Error("failed to create google user", "error", err, "email", userInfo.Email)
		return nil, err
	}

	user.ID, err = uuid.Parse(*userID)
	if err != nil {
		s.log.Error("failed to parse user id", "error", err, "email", userInfo.Email)
		return nil, fmt.Errorf("произошла ошибка")
	}

	tokenPair, err := s.jwtHelper.GenerateTokenPair(user.ID.String(), user.Email)
	if err != nil {
		s.log.Error("failed to generate JWT tokens", "error", err)
		return nil, fmt.Errorf("произошла ошибка")
	}

	err = s.storeRefreshToken(ctx, user.ID.String(), tokenPair.RefreshToken)
	if err != nil {
		s.log.Error("failed to store refresh token", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	s.log.Info("google auth successful", "user_id", user.ID, "email", user.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       user.ID.String(),
		Email:        user.Email,
	}, nil
}

func (s *service) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token обязателен")
	}

	claims, err := s.jwtHelper.ParseJWT(refreshToken)
	if err != nil {
		s.log.Warn("invalid refresh token format", "error", err)
		return nil, fmt.Errorf("неверный refresh token")
	}

	if claims.Type != "refresh" {
		s.log.Warn("attempt to use non-refresh token for refresh", "token_type", claims.Type)
		return nil, fmt.Errorf("неверный тип токена")
	}

	storedToken, err := s.redis.GetRefreshToken(ctx, claims.UserID)
	if err != nil {
		s.log.Warn("refresh token not found in storage", "user_id", claims.UserID, "error", err)
		return nil, fmt.Errorf("refresh token не найден или истек")
	}

	if storedToken != refreshToken {
		s.log.Warn("refresh token mismatch", "user_id", claims.UserID)
		return nil, fmt.Errorf("неверный refresh token")
	}

	newTokenPair, err := s.jwtHelper.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		s.log.Error("failed to generate new token pair", "error", err, "user_id", claims.UserID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	err = s.storeRefreshToken(ctx, claims.UserID, newTokenPair.RefreshToken)
	if err != nil {
		s.log.Error("failed to store new refresh token", "error", err, "user_id", claims.UserID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	s.log.Info("tokens refreshed successfully", "user_id", claims.UserID)

	return &AuthResponse{
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		ExpiresIn:    newTokenPair.ExpiresIn,
		UserID:       claims.UserID,
		Email:        claims.Email,
	}, nil
}

func (s *service) Logout(ctx context.Context, userID string) error {
	err := s.redis.DeleteRefreshToken(ctx, userID)
	if err != nil {
		s.log.Error("failed to delete refresh token", "error", err, "user_id", userID)
		return fmt.Errorf("не удалось выполнить выход")
	}

	s.log.Info("user logged out successfully", "user_id", userID)
	return nil
}

func (s *service) GenerateGoogleOAuthURL() (string, string, error) {
	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("не удалось сгенерировать параметр безопасности")
	}
	s.log.Info("generated google oauth url in service")

	u, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		return "", "", fmt.Errorf("неверный URL OAuth: %w", err)
	}

	q := u.Query()
	q.Set("client_id", s.googleConfig.ClientID)
	q.Set("redirect_uri", s.googleConfig.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "email profile")
	q.Set("state", state)
	u.RawQuery = q.Encode()

	s.log.Info("generated google oauth url in service")
	return u.String(), state, nil
}

func (s *service) ValidateToken(token string) (bool, error) {
	valid, err := s.jwtHelper.ValidateToken(token)
	if err != nil {
		s.log.Warn("token validation failed", "error", err)
		return false, fmt.Errorf("неверный токен")
	}
	return valid, nil
}

func (s *service) storeRefreshToken(ctx context.Context, userID, refreshToken string) error {
	return s.redis.StoreRefreshToken(ctx, userID, refreshToken)
}

func (s *service) getRefreshTokenKey(userID string) string {
	return fmt.Sprintf("refresh_token:%s", userID)
}

func (s *service) exchangeCodeForToken(code string) (string, error) {
	formData := url.Values{}
	formData.Set("code", code)
	formData.Set("client_id", s.googleConfig.ClientID)
	formData.Set("client_secret", s.googleConfig.ClientSecret)
	formData.Set("redirect_uri", s.googleConfig.RedirectURL)
	formData.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", formData)
	if err != nil {
		return "", fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google token exchange failed: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("google oauth error: %s", result.Error)
	}

	return result.AccessToken, nil
}

func (s *service) getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
