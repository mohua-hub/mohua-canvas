package service

import (
	"strings"

	"github.com/google/uuid"
	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
)

type CanvasAudioTaskCreateInput struct {
	WorkspaceID          string
	SourceID        string
	NodeID          string
	ClientTaskID    string
	Model           string
	ChannelID       string
	LocalChannelID   string
	ChannelName     string
	WorkflowRef     string
	Prompt          string
	Endpoint        string
	ContentType     string
	RequestBody     string
}

func CreateCanvasAudioTask(input CanvasAudioTaskCreateInput) (model.CanvasAudioTask, error) {
	current := now()
	task := model.CanvasAudioTask{
		ID:              firstVideoTaskValue(input.ClientTaskID, "canvas_audio_task_"+uuid.NewString()),
		WorkspaceID:          strings.TrimSpace(input.WorkspaceID),
		Source:          "canvas",
		SourceID:        strings.TrimSpace(input.SourceID),
		NodeID:          strings.TrimSpace(input.NodeID),
		Model:           strings.TrimSpace(input.Model),
		ChannelID:       strings.TrimSpace(input.ChannelID),
		LocalChannelID:   strings.TrimSpace(input.LocalChannelID),
		ChannelName:     strings.TrimSpace(input.ChannelName),
		WorkflowRef:     input.WorkflowRef,
		Status:          "queued",
		Progress:        0,
		Prompt:          strings.TrimSpace(input.Prompt),
		Endpoint:        strings.TrimSpace(input.Endpoint),
		ContentType:     strings.TrimSpace(input.ContentType),
		RequestBody:     input.RequestBody,
		CreatedAt:       current,
		UpdatedAt:       current,
	}
	return repository.SaveCanvasAudioTask(task)
}

func GetWorkspaceCanvasAudioTask(workspaceID string, id string) (model.CanvasAudioTask, bool, error) {
	return repository.GetWorkspaceCanvasAudioTask(strings.TrimSpace(workspaceID), strings.TrimSpace(id))
}

func SaveCanvasAudioTask(task model.CanvasAudioTask) (model.CanvasAudioTask, error) {
	task.UpdatedAt = now()
	return repository.UpdateCanvasAudioTask(task)
}

func CanvasAudioTaskResponse(task model.CanvasAudioTask) map[string]any {
	result := map[string]any{
		"id":           task.ID,
		"object":       "canvas.audio.task",
		"source":       task.Source,
		"source_id":    task.SourceID,
		"node_id":      task.NodeID,
		"model":        task.Model,
		"status":       task.Status,
		"progress":     task.Progress,
		"prompt":       task.Prompt,
		"created_at":   task.CreatedAt,
		"updated_at":   task.UpdatedAt,
		"started_at":   task.StartedAt,
		"completed_at": task.CompletedAt,
		"createdAt":    task.CreatedAt,
		"updatedAt":    task.UpdatedAt,
	}
	if task.AudioURL != "" {
		result["url"] = task.AudioURL
		result["audio_url"] = task.AudioURL
		result["storageKey"] = task.StorageKey
		result["mimeType"] = task.MimeType
		result["bytes"] = task.Bytes
	}
	if task.Error != "" || task.ErrorDetail != "" {
		result["error"] = map[string]any{"message": firstVideoTaskValue(task.Error, task.ErrorDetail)}
		result["error_detail"] = task.ErrorDetail
	}
	return result
}
