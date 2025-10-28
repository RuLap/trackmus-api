package task

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeImage MediaType = "image"
)

type LinkType string

const (
	LinkTypeYoutube LinkType = "youtube"
	LinkTypeSpotify LinkType = "spotify"
	LinkTypeOther   LinkType = "other"
)

type Task struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Title     string    `db:"title"`
	TargetBPM int       `db:"target_bpm"`
	CreatedAt time.Time `db:"created_at"`
}

type Session struct {
	ID         uuid.UUID `db:"id"`
	TaskID     uuid.UUID `db:"task_id"`
	BPM        int       `db:"bpm"`
	Note       string    `db:"note"`
	Confidence int       `db:"confidence"`
	StartTime  time.Time `db:"start_time"`
	EndTime    time.Time `db:"end_time"`
}

type Media struct {
	ID        uuid.UUID `db:"id"`
	TaskID    uuid.UUID `db:"task_id"`
	Type      MediaType `db:"type"`
	Filename  string    `db:"filename"`
	URL       string    `db:"url"`
	Size      int64     `db:"size"`
	Duration  int       `db:"duration"`
	CreatedAt time.Time `db:"created_at"`
}

type Link struct {
	ID        uuid.UUID `db:"id"`
	TaskID    uuid.UUID `db:"task_id"`
	URL       string    `db:"url"`
	Title     string    `db:"title"`
	Type      LinkType  `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}

func (mt MediaType) IsValid() bool {
	switch mt {
	case MediaTypeVideo, MediaTypeAudio, MediaTypeImage:
		return true
	}

	return false
}

func (lt LinkType) IsValid() bool {
	switch lt {
	case LinkTypeYoutube, LinkTypeSpotify, LinkTypeOther:
		return true
	}

	return false
}
