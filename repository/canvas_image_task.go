package repository

import (
	"strings"

	"github.com/tigerowo/infinite-canvas/model"
	"gorm.io/gorm"
)

func SaveCanvasImageTask(task model.CanvasImageTask) (model.CanvasImageTask, error) {
	db, err := DB()
	if err != nil {
		return task, err
	}
	return task, db.Save(&task).Error
}

func UpdateCanvasImageTask(task model.CanvasImageTask) (model.CanvasImageTask, error) {
	db, err := DB()
	if err != nil {
		return task, err
	}

	return task, db.Model(&model.CanvasImageTask{}).
		Where("workspace_id = ? AND id = ?", task.WorkspaceID, task.ID).
		Select("*").
		Updates(&task).Error
}

func GetWorkspaceCanvasImageTask(workspaceID string, id string) (model.CanvasImageTask, bool, error) {
	db, err := DB()
	if err != nil {
		return model.CanvasImageTask{}, false, err
	}
	var task model.CanvasImageTask
	err = db.First(&task, "workspace_id = ? AND id = ?", workspaceID, id).Error
	if err != nil {
		return model.CanvasImageTask{}, false, nil
	}
	return task, true, nil
}

func HasActiveCanvasImageTasks() (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Model(&model.CanvasImageTask{}).
		Where("status IN ?", []string{"queued", "processing", "running", "in_progress"}).
		Count(&count).Error
	return count > 0, err
}

func DeleteFinishedCanvasImageTasksBefore(before string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Where("completed_at <> ? AND completed_at < ?", "", before).
		Where("status IN ?", []string{"completed", "failed"}).
		Delete(&model.CanvasImageTask{}).Error
}

func DeleteWorkspaceCanvasTasks(workspaceID string, sourceID string, nodeIDs []string) error {
	db, err := DB()
	if err != nil {
		return err
	}

	nodeIDs = uniqueTrimmedValues(nodeIDs...)

	return db.Transaction(func(tx *gorm.DB) error {
		deleteTasks := func(task any) error {
			query := tx.Where(
				"workspace_id = ? AND source = ? AND source_id = ?",
				workspaceID,
				"canvas",
				sourceID,
			)
			if len(nodeIDs) > 0 {
				query = query.Where("node_id IN ?", nodeIDs)
			}
			return query.Delete(task).Error
		}

		if err := deleteTasks(&model.CanvasImageTask{}); err != nil {
			return err
		}
		return deleteTasks(&model.CanvasAudioTask{})
	})
}

func uniqueTrimmedValues(values ...string) []string {
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
