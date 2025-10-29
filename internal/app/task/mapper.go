package task

import "github.com/google/uuid"

// Task --------------------------------------------------------------------------------------

func TaskToGetShortResponse(model *Task, progress float64) GetTaskShortResponse {
	return GetTaskShortResponse{
		ID:        model.ID.String(),
		Title:     model.Title,
		TargetBPM: model.TargetBPM,
		Progress:  progress,
	}
}

func TaskToGetResponse(model *Task, sessions []GetSessionResponse, media []GetMediaResponse, links []GetLinkResponse) *GetTaskResponse {
	return &GetTaskResponse{
		ID:        model.ID.String(),
		Title:     model.Title,
		TargetBPM: model.TargetBPM,
		Sessions:  sessions,
		Media:     media,
		Links:     links,
	}
}

func SaveRequestToTask(req *SaveTaskRequest, userID uuid.UUID) Task {
	return Task{
		UserID:    userID,
		Title:     req.Title,
		TargetBPM: req.TargetBPM,
	}
}

// Session -----------------------------------------------------------------------------------

func SessionToGetResponse(model *Session) GetSessionResponse {
	return GetSessionResponse{
		ID:         model.ID.String(),
		BPM:        model.BPM,
		Note:       model.Note,
		Confidence: model.Confidence,
		StartTime:  model.StartTime,
		EndTime:    model.EndTime,
		Duration:   model.GetDurationSeconds(),
	}
}

func SaveRequestToSession(req *SaveSessionRequest, userID uuid.UUID) Session {
	return Session{
		BPM:        req.BPM,
		Note:       req.Note,
		Confidence: req.Confidence,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	}
}

// Media -------------------------------------------------------------------------------------

func MediaToGetResponse(model *Media) GetMediaResponse {
	return GetMediaResponse{
		ID:        model.ID.String(),
		Type:      string(model.Type),
		Filename:  model.Filename,
		URL:       model.URL,
		Size:      model.Size,
		Duration:  model.Duration,
		CreatedAt: model.CreatedAt,
	}
}

func SaveRequestToMedia(req *SaveMediaRequest, taskID uuid.UUID) Media {
	return Media{
		Type:     MediaType(req.Type),
		Filename: req.Filename,
		URL:      req.URL,
		Size:     req.Size,
		Duration: req.Duration,
	}
}

// Link --------------------------------------------------------------------------------------

func LinkToGetResponse(model *Link) GetLinkResponse {
	return GetLinkResponse{
		ID:        model.ID.String(),
		URL:       model.URL,
		Title:     model.Title,
		Type:      string(model.Type),
		CreatedAt: model.CreatedAt,
	}
}

func SaveRequestToLink(req *SaveLinkRequest, taskID uuid.UUID) Link {
	return Link{
		URL:   req.URL,
		Title: req.Title,
		Type:  req.Type,
	}
}
