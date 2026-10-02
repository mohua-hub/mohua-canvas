package service

import (
	"strings"

	"github.com/google/uuid"
	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
)

type CanvasImageTaskCreateInput struct {
	WorkspaceID          string
	Source          string
	SourceID        string
	NodeID          string
	ClientTaskID    string
	Model           string
	ChannelID       string
	LocalChannelID   string
	ChannelName     string
	WorkflowRef     string
	Prompt          string
	GenerationType  string
	Endpoint        string
	ContentType     string
	RequestBody     string
}

func CreateCanvasImageTask(input CanvasImageTaskCreateInput) (model.CanvasImageTask, error) {
	current := now()
	task := model.CanvasImageTask{
		ID:              firstVideoTaskValue(input.ClientTaskID, "canvas_image_task_"+uuid.NewString()),
		WorkspaceID:          strings.TrimSpace(input.WorkspaceID),
		Source:          normalizeCanvasImageTaskSource(input.Source),
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
		GenerationType:  strings.TrimSpace(input.GenerationType),
		Endpoint:        strings.TrimSpace(input.Endpoint),
		ContentType:     strings.TrimSpace(input.ContentType),
		RequestBody:     input.RequestBody,
		CreatedAt:       current,
		UpdatedAt:       current,
	}
	saved, err := repository.SaveCanvasImageTask(task)
	if err == nil {
		WakeVideoTaskPoller()
	}
	return saved, err
}

func GetWorkspaceCanvasImageTask(workspaceID string, id string) (model.CanvasImageTask, bool, error) {
	return repository.GetWorkspaceCanvasImageTask(strings.TrimSpace(workspaceID), strings.TrimSpace(id))
}

func ListWorkspaceCanvasImageTasks(workspaceID string, sources []string, limit int) ([]map[string]any, error) {
	tasks, err := repository.ListWorkspaceCanvasImageTasks(strings.TrimSpace(workspaceID), normalizeCanvasImageTaskSources(sources), limit)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, CanvasImageTaskResponse(task))
	}
	return result, nil
}

func BatchWorkspaceCanvasImageTasks(workspaceID string, ids []string) ([]map[string]any, error) {
	tasks, err := repository.BatchWorkspaceCanvasImageTasks(strings.TrimSpace(workspaceID), ids)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, CanvasImageTaskResponse(task))
	}
	return result, nil
}

func DeleteWorkspaceCanvasImageTask(workspaceID string, id string) error {
	return repository.DeleteWorkspaceCanvasImageTask(strings.TrimSpace(workspaceID), strings.TrimSpace(id))
}

func DeleteWorkspaceCanvasTasks(workspaceID string, sourceID string, nodeIDs []string) error {
	return repository.DeleteWorkspaceCanvasTasks(strings.TrimSpace(workspaceID), strings.TrimSpace(sourceID), nodeIDs)
}

func SaveCanvasImageTask(task model.CanvasImageTask) (model.CanvasImageTask, error) {
	task.UpdatedAt = now()
	return repository.UpdateCanvasImageTask(task)
}

func CanvasImageTaskResponse(task model.CanvasImageTask) map[string]any {
	result := map[string]any{
		"id":             task.ID,
		"object":         "canvas.image.task",
		"source":         task.Source,
		"source_id":      task.SourceID,
		"node_id":        task.NodeID,
		"model":          task.Model,
		"status":         task.Status,
		"progress":       task.Progress,
		"prompt":         task.Prompt,
		"generationType": task.GenerationType,
		"created_at":     task.CreatedAt,
		"updated_at":     task.UpdatedAt,
		"started_at":     task.StartedAt,
		"completed_at":   task.CompletedAt,
		"createdAt":      task.CreatedAt,
		"updatedAt":      task.UpdatedAt,
	}
	if task.ImageURL != "" {
		result["url"] = task.ImageURL
		result["image_url"] = task.ImageURL
		if len(task.ImageURLs) > 0 {
			result["image_urls"] = task.ImageURLs
		}
		result["storageKey"] = task.StorageKey
		result["width"] = task.Width
		result["height"] = task.Height
		result["mimeType"] = task.MimeType
		result["bytes"] = task.Bytes
	}
	if task.WorkflowRef != "" {
		result["workflowRef"] = task.WorkflowRef
	}
	if task.Error != "" || task.ErrorDetail != "" {
		result["error"] = map[string]any{"message": firstVideoTaskValue(task.Error, task.ErrorDetail)}
		result["error_detail"] = task.ErrorDetail
	}
	return result
}

func normalizeCanvasImageTaskSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "image-workbench":
		return "image-workbench"
	case "workflow":
		return "workflow"
	case "canvas", "":
		return "canvas"
	default:
		return "canvas"
	}
}

func normalizeCanvasImageTaskSources(sources []string) []string {
	result := make([]string, 0, len(sources))
	seen := map[string]bool{}
	for _, source := range sources {
		normalized := normalizeCanvasImageTaskSource(source)
		if normalized != "" && !seen[normalized] {
			result = append(result, normalized)
			seen[normalized] = true
		}
	}
	return result
}
