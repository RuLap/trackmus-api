package task

import (
	"time"
)

// Task -------------------------------------------------------------------------------------
type GetTaskShortResponse struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	TargetBPM int     `json:"target_bpm"`
	Progress  float64 `json:"progress"`
}

type GetTaskResponse struct {
	ID        string               `json:"id"`
	Title     string               `json:"title"`
	TargetBPM int                  `json:"target_bpm"`
	Sessions  []GetSessionResponse `json:"sessions"`
	Media     []GetMediaResponse   `json:"media"`
	Links     []GetLinkResponse    `json:"links"`
}

type SaveTaskRequest struct {
	Title     string `json:"title" validate:"required,min=1,max=50"`
	TargetBPM int    `json:"target_bpm" validate:"required,number"`
}

// Session -------------------------------------------------------------------------------------
type GetSessionResponse struct {
	ID         string    `json:"id"`
	BPM        int       `json:"bpm"`
	Note       string    `json:"note"`
	Confidence int       `json:"confidence"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Duration   int       `json:"duration"`
}

type SaveSessionRequest struct {
	BPM        int       `json:"bpm" validate:"required,number"`
	Note       string    `json:"note"`
	Confidence int       `json:"confidence" validate:"required,number,min=1,max=5"`
	StartTime  time.Time `json:"start_time" validate:"required"`
	EndTime    time.Time `json:"end_time" validate:"required"`
}

// Media -------------------------------------------------------------------------------------
type GetMediaResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Filename  string    `json:"filename"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	Duration  int       `json:"duration"`
	CreatedAt time.Time `json:"created_at"`
}

type GetUploadURLResponse struct {
	MediaID string
	URL     string
}

type ConfirmMediaUploadRequest struct {
	Type     MediaType `json:"type" validate:"required"`
	Filename string    `json:"filename" validate:"required,min=1"`
	Size     int64     `json:"size" validate:"required,number"`
	Duration int       `json:"duration" validate:"required,number"`
}

// Link -------------------------------------------------------------------------------------
type GetLinkResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type SaveLinkRequest struct {
	Title string   `json:"title" validate:"required,min=1,max=50"`
	Type  LinkType `json:"type" validate:"required,min=1,max=50"`
}
