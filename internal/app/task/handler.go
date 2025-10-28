package task

import "net/http"

type Handler struct {
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) GetSessionByID(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) GetMediaByID(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {}
