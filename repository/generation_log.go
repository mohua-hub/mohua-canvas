package repository

import (
	"strings"

	"github.com/tigerowo/infinite-canvas/model"
)

func ListImageGenerationLogs(workspaceID string, limit int) ([]model.ImageGenerationLog, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 500
	}
	var logs []model.ImageGenerationLog
	err = db.Where("workspace_id = ? AND deleted_at = ?", workspaceID, "").Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func HasAnyImageGenerationLog(workspaceID string) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Model(&model.ImageGenerationLog{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count > 0, err
}

func UpsertImageGenerationLogs(workspaceID string, logs []model.ImageGenerationLog) error {
	db, err := DB()
	if err != nil {
		return err
	}
	for _, log := range logs {
		log.WorkspaceID = workspaceID
		if strings.TrimSpace(log.ID) == "" || isDeletedImageGenerationLog(workspaceID, log) {
			continue
		}
		var existing model.ImageGenerationLog
		found := false
		if err := db.Where("workspace_id = ? AND id = ?", workspaceID, log.ID).First(&existing).Error; err == nil {
			found = true
		} else if log.TaskID != "" {
			if err := db.Where("workspace_id = ? AND deleted_at = ? AND task_id = ?", workspaceID, "", log.TaskID).First(&existing).Error; err == nil {
				found = true
			}
		}
		if !found && log.ImageID != "" {
			if err := db.Where("workspace_id = ? AND deleted_at = ? AND image_id = ?", workspaceID, "", log.ImageID).First(&existing).Error; err == nil {
				found = true
			}
		}
		if found {
			log.ID = existing.ID
			log.WorkspaceID = workspaceID
			log.DeletedAt = ""
		}
		if err := db.Save(&log).Error; err != nil {
			return err
		}
	}
	return nil
}

func SoftDeleteImageGenerationLog(workspaceID string, id string, deletedAt string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	var log model.ImageGenerationLog
	err = db.Where("workspace_id = ? AND (id = ? OR task_id = ? OR image_id = ?)", workspaceID, id, id, id).First(&log).Error
	if err != nil {
		return nil
	}
	return db.Model(&log).Updates(map[string]any{
		"deleted_at":   deletedAt,
		"updated_at":   deletedAt,
		"payload_json": "",
	}).Error
}

func SoftDeleteImageGenerationLogs(workspaceID string, ids []string, deletedAt string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	keys := generationLogIdentityValues(ids...)
	if len(keys) == 0 {
		return nil
	}
	return db.Model(&model.ImageGenerationLog{}).
		Where("workspace_id = ? AND (id IN ? OR task_id IN ? OR image_id IN ?)", workspaceID, keys, keys, keys).
		Updates(map[string]any{
			"deleted_at":   deletedAt,
			"updated_at":   deletedAt,
			"payload_json": "",
		}).Error
}

func CleanupDeletedImageGenerationLogs(before string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Where("deleted_at <> ? AND deleted_at < ?", "", before).Delete(&model.ImageGenerationLog{}).Error
}

func isDeletedImageGenerationLog(workspaceID string, log model.ImageGenerationLog) bool {
	db, err := DB()
	if err != nil {
		return false
	}
	keys := generationLogIdentityValues(log.ID, log.TaskID, log.ImageID)
	if len(keys) == 0 {
		return false
	}
	var count int64
	_ = db.Model(&model.ImageGenerationLog{}).
		Where("workspace_id = ? AND deleted_at <> ? AND (id IN ? OR task_id IN ? OR image_id IN ?)", workspaceID, "", keys, keys, keys).
		Count(&count).Error
	return count > 0
}

func generationLogIdentityValues(values ...string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}

