package auth

import (
	"encoding/json"
	"net/http"

	validation "github.com/RuLap/trackmus-api/internal/pkg/validator"
	"github.com/darahayes/go-boom"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "Ошибки валидации", errors)
		return
	}

	response, err := h.service.Register(r.Context(), req)
	if err != nil {
		boom.BadRequest(w, err.Error())
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, err.Error())
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "Ошибки валидации", errors)
		return
	}

	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		boom.BadRequest(w, err.Error())
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	var req GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "Неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "Ошибки валидации", errors)
		return
	}

	response, err := h.service.GoogleAuth(r.Context(), req)
	if err != nil {
		boom.Unathorized(w, err.Error())
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) GoogleAuthURL(w http.ResponseWriter, r *http.Request) {
	url, state, err := h.service.GenerateGoogleOAuthURL()
	if err != nil {
		boom.Internal(w, "Не удалось сгенерировать URL для авторизации")
		return
	}

	response := map[string]string{
		"url":   url,
		"state": state,
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Copy this code and use in POST /auth/google",
		"code":    code,
		"state":   state,
	})
}

func (h *Handler) SendConfirmationLink(w http.ResponseWriter, r *http.Request) {
	var req ConfirmEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "Ошибки валидации", errors)
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		boom.Unathorized(w, "требуется аутентификация")
		return
	}

	if err := h.service.ConfirmEmail(r.Context(), req.Token, userID); err != nil {
		boom.BadRequest(w, err.Error())
		return
	}

	h.sendJSON(w, map[string]interface{}{
		"success": true,
		"message": "Email успешно подтвержден",
	}, http.StatusOK)
}

func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var req ConfirmEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "Ошибки валидации", errors)
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		boom.Unathorized(w, "требуется аутентификация")
		return
	}

	if err := h.service.ConfirmEmail(r.Context(), req.Token, userID); err != nil {
		boom.BadRequest(w, err.Error())
		return
	}

	h.sendJSON(w, map[string]interface{}{
		"success": true,
		"message": "Email успешно подтвержден",
	}, http.StatusOK)
}

func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "Неверный формат запроса")
		return
	}

	if req.RefreshToken == "" {
		boom.BadRequest(w, "Refresh token обязателен")
		return
	}

	tokens, err := h.service.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		boom.Unathorized(w, "Не удалось обновить токены")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		boom.Unathorized(w, "Пользователь не авторизован")
		return
	}

	err := h.service.Logout(r.Context(), userID)
	if err != nil {
		boom.Internal(w, "Не удалось выполнить выход")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
