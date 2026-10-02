package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
)

const generationLogLimit = 1000

func CurrentWorkspaceVideoGenerationLogs(ctx context.Context) ([]json.RawMessage, error) {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	logs, err := repository.ListVideoGenerationLogs(workspaceID, generationLogLimit)
	if err != nil {
		return nil, err
	}
	return videoGenerationPayloads(logs), nil
}

func SaveCurrentWorkspaceVideoGenerationLogs(ctx context.Context, raws []json.RawMessage) ([]json.RawMessage, error) {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	logs := make([]model.VideoGenerationLog, 0, len(raws))
	for _, raw := range raws {
		log := videoGenerationLogFromPayload(raw)
		if log.ID != "" {
			logs = append(logs, log)
		}
	}
	if err := repository.UpsertVideoGenerationLogs(workspaceID, logs); err != nil {
		return nil, err
	}
	return CurrentWorkspaceVideoGenerationLogs(ctx)
}

func DeleteCurrentWorkspaceVideoGenerationLog(ctx context.Context, id string) error {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	return repository.SoftDeleteVideoGenerationLog(workspaceID, strings.TrimSpace(id), now())
}

func DeleteCurrentWorkspaceVideoGenerationLogs(ctx context.Context, ids []string) error {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	return repository.SoftDeleteVideoGenerationLogs(workspaceID, ids, now())
}

func CurrentWorkspaceImageGenerationLogs(ctx context.Context) ([]json.RawMessage, error) {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	if err := migrateWorkspaceImageGenerationLogs(workspaceID); err != nil {
		return nil, err
	}
	logs, err := repository.ListImageGenerationLogs(workspaceID, generationLogLimit)
	if err != nil {
		return nil, err
	}
	return imageGenerationPayloads(logs), nil
}

func SaveCurrentWorkspaceImageGenerationLogs(ctx context.Context, raws []json.RawMessage) ([]json.RawMessage, error) {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	logs := make([]model.ImageGenerationLog, 0, len(raws))
	for _, raw := range raws {
		log := imageGenerationLogFromPayload(raw)
		if log.ID != "" {
			logs = append(logs, log)
		}
	}
	if err := repository.UpsertImageGenerationLogs(workspaceID, logs); err != nil {
		return nil, err
	}
	return CurrentWorkspaceImageGenerationLogs(ctx)
}

func DeleteCurrentWorkspaceImageGenerationLog(ctx context.Context, id string) error {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	return repository.SoftDeleteImageGenerationLog(workspaceID, strings.TrimSpace(id), now())
}

func DeleteCurrentWorkspaceImageGenerationLogs(ctx context.Context, ids []string) error {
	workspaceID := WorkspaceID
	cleanupGenerationLogs()
	return repository.SoftDeleteImageGenerationLogs(workspaceID, ids, now())
}

func videoGenerationPayloads(logs []model.VideoGenerationLog) []json.RawMessage {
	result := make([]json.RawMessage, 0, len(logs))
	for _, log := range logs {
		if strings.TrimSpace(log.PayloadJSON) != "" {
			result = append(result, json.RawMessage(log.PayloadJSON))
		}
	}
	return result
}

func imageGenerationPayloads(logs []model.ImageGenerationLog) []json.RawMessage {
	result := make([]json.RawMessage, 0, len(logs))
	for _, log := range logs {
		if strings.TrimSpace(log.PayloadJSON) != "" {
			result = append(result, json.RawMessage(log.PayloadJSON))
		}
	}
	return result
}

func videoGenerationLogFromPayload(raw json.RawMessage) model.VideoGenerationLog {
	record := parseGenerationLogRecord(raw)
	task := generationLogRecord(record["task"])
	video := generationLogRecord(record["video"])
	current := now()
	createdAt := generationLogCreatedAt(record)
	if createdAt == "" {
		createdAt = current
	}
	return model.VideoGenerationLog{
		ID:          generationLogString(record["id"]),
		TaskID:      firstGenerationLogValue(generationLogString(task["id"]), generationLogString(task["task_id"]), generationLogString(record["taskId"])),
		VideoID:     firstGenerationLogValue(generationLogString(task["video_id"]), generationLogString(video["id"]), generationLogString(record["videoId"])),
		Status:      generationLogString(record["status"]),
		PayloadJSON: string(raw),
		CreatedAt:   createdAt,
		UpdatedAt:   current,
		DeletedAt:   "",
	}
}

func imageGenerationLogFromPayload(raw json.RawMessage) model.ImageGenerationLog {
	record := parseGenerationLogRecord(raw)
	task := generationLogRecord(record["task"])
	image := firstGenerationLogRecord(record["image"], record["result"], record["output"])
	current := now()
	createdAt := generationLogCreatedAt(record)
	if createdAt == "" {
		createdAt = current
	}
	return model.ImageGenerationLog{
		ID:          generationLogString(record["id"]),
		TaskID:      firstGenerationLogValue(generationLogString(task["id"]), generationLogString(task["task_id"]), generationLogString(record["taskId"])),
		ImageID:     firstGenerationLogValue(generationLogString(image["id"]), generationLogString(image["storageKey"]), generationLogString(image["url"]), generationLogString(record["imageId"])),
		Status:      generationLogString(record["status"]),
		PayloadJSON: string(raw),
		CreatedAt:   createdAt,
		UpdatedAt:   current,
		DeletedAt:   "",
	}
}

func parseGenerationLogRecord(raw json.RawMessage) map[string]any {
	var record map[string]any
	if err := json.Unmarshal(raw, &record); err != nil {
		return map[string]any{}
	}
	return record
}

func generationLogRecord(value any) map[string]any {
	if record, ok := value.(map[string]any); ok {
		return record
	}
	return map[string]any{}
}

func firstGenerationLogRecord(values ...any) map[string]any {
	for _, value := range values {
		record := generationLogRecord(value)
		if len(record) > 0 {
			return record
		}
	}
	return map[string]any{}
}

func generationLogString(value any) string {
	switch item := value.(type) {
	case string:
		return strings.TrimSpace(item)
	case float64:
		if item == float64(int64(item)) {
			return strconv.FormatInt(int64(item), 10)
		}
		return fmt.Sprintf("%v", item)
	case bool:
		return strconv.FormatBool(item)
	default:
		return ""
	}
}

func generationLogCreatedAt(record map[string]any) string {
	for _, key := range []string{"createdAt", "created_at", "time"} {
		value := generationLogString(record[key])
		if value != "" {
			return value
		}
	}
	return ""
}

func firstGenerationLogValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func cleanupGenerationLogs() {
	before := time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	_ = repository.CleanupDeletedVideoGenerationLogs(before)
	_ = repository.CleanupDeletedImageGenerationLogs(before)
}

func migrateWorkspaceImageGenerationLogs(workspaceID string) error {
	config, found, err := repository.GetWorkspaceConfig(workspaceID)
	if err != nil || !found || strings.TrimSpace(config.ImageHistory) == "" {
		return err
	}
	var legacy struct {
		Logs []json.RawMessage `json:"logs"`
	}
	if err := json.Unmarshal([]byte(config.ImageHistory), &legacy); err != nil || len(legacy.Logs) == 0 {
		config.ImageHistory = ""
		_, saveErr := repository.SaveWorkspaceConfig(config)
		return saveErr
	}
	logs := make([]model.ImageGenerationLog, 0, len(legacy.Logs))
	for _, raw := range legacy.Logs {
		log := imageGenerationLogFromPayload(raw)
		if log.ID != "" {
			logs = append(logs, log)
		}
	}
	if err := repository.UpsertImageGenerationLogs(workspaceID, logs); err != nil {
		return err
	}
	config.ImageHistory = ""
	_, err = repository.SaveWorkspaceConfig(config)
	return err
}

