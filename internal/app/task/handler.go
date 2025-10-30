package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RuLap/trackmus-api/internal/pkg/errors"
	validation "github.com/RuLap/trackmus-api/internal/pkg/validator"
	"github.com/darahayes/go-boom"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	log     *slog.Logger
	service Service
}

func NewHandler(log *slog.Logger, service Service) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) GetActiveTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	response, err := h.service.GetActiveTasks(r.Context(), userID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) GetCompletedTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	response, err := h.service.GetCompletedTasks(r.Context(), userID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	response, err := h.service.GetTaskByID(r.Context(), *id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	var req SaveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "ошибки валидации", errors)
		return
	}

	response, err := h.service.CreateTask(r.Context(), &req, userID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	var req SaveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "ошибки валидации", errors)
		return
	}

	response, err := h.service.UpdateTask(r.Context(), &req, *id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	response, err := h.service.CompleteTask(r.Context(), *id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) GetSessionByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	response, err := h.service.GetSessionByID(r.Context(), *id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	var req SaveSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "ошибки валидации", errors)
		return
	}

	response, err := h.service.CreateSession(r.Context(), &req, *taskID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	var req SaveMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "ошибки валидации", errors)
		return
	}

	response, err := h.service.SaveMedia(r.Context(), &req, *taskID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) RemoveMedia(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	err = h.service.RemoveMedia(r.Context(), *taskID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	taskID, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	var req SaveLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "ошибки валидации", errors)
		return
	}

	response, err := h.service.SaveLink(r.Context(), &req, *taskID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) RemoveLink(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	err = h.service.RemoveLink(r.Context(), *id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getUrlParamUuid(r *http.Request, param string) (*uuid.UUID, error) {
	str := chi.URLParam(r, param)
	if str == "" {
		err := fmt.Errorf("параметр %s необходим", param)
		h.log.Error("Incorrect ID in URL", param, str, "error", err.Error())
		return nil, err
	}

	uid, err := uuid.Parse(str)
	if err != nil {
		err := fmt.Errorf("неверный формат параметра %s", param)
		h.log.Error("Incorrect ID in URL", param, str, "error", err.Error())
		return nil, err
	}

	return &uid, nil
}

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := ctx.Value("userID").(string)
	if !ok {
		h.log.Error("Incorrect ID in URL", "error", errors.ErrCommon)
		return uuid.Nil, fmt.Errorf(errors.ErrCommon)
	}

	return uuid.Parse(userIDStr)
}
