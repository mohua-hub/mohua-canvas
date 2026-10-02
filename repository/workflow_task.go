package repository

import (
	"errors"
	"strings"

	"github.com/tigerowo/infinite-canvas/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func InsertWorkflowTask(task any) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Create(task).Error
}

type RunningHubWorkflowTask struct {
	ID             string
	WorkspaceID         string
	WorkflowRef    string
	UpstreamTaskID string
	CreatedAt      string
}

func ListDueRunningHubWorkflowTasks(afterCreatedAt, afterID string, limit int) ([]RunningHubWorkflowTask, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	var tasks []RunningHubWorkflowTask
	err = db.Raw(`
		SELECT id, workspace_id, workflow_ref, upstream_task_id, created_at
		FROM (
			SELECT id, workspace_id, workflow_ref, response_body AS upstream_task_id, created_at
			FROM canvas_image_tasks
			WHERE status = 'running' AND workflow_ref <> '' AND response_body LIKE '%"upstreamTaskId":%'
			UNION ALL
			SELECT id, workspace_id, workflow_ref, upstream_task_id, created_at
			FROM video_tasks
			WHERE status = 'running' AND workflow_ref <> '' AND upstream_task_id <> ''
			UNION ALL
			SELECT id, workspace_id, workflow_ref, response_body AS upstream_task_id, created_at
			FROM canvas_audio_tasks
			WHERE status = 'running' AND workflow_ref <> '' AND response_body LIKE '%"upstreamTaskId":%'
		) AS workflow_tasks
		WHERE (? = '' OR created_at > ? OR (created_at = ? AND id > ?))
		ORDER BY created_at ASC, id ASC
		LIMIT ?`, afterCreatedAt, afterCreatedAt, afterCreatedAt, afterID, limit).Scan(&tasks).Error
	return tasks, err
}

func CompleteWorkflowTask(taskID, status string, urls []string, failure string) (bool, string, string, error) {
	db, err := DB()
	if err != nil {
		return false, "", "", err
	}
	changed := false
	workflowRef := ""
	workspaceID := ""
	err = db.Transaction(func(tx *gorm.DB) error {
		current := workspaceConfigTimestamp()
		switch {
		case strings.HasPrefix(taskID, "wf-image-"):
			var task model.CanvasImageTask
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, "id = ?", taskID).Error; err != nil {
				return err
			}
			if task.Status == "completed" || task.Status == "failed" {
				return nil
			}
			task.Status, task.Error, task.UpdatedAt, task.CompletedAt = status, failure, current, current
			if status == "completed" {
				task.Progress = 100
				task.ImageURLs = urls
				if len(urls) > 0 {
					task.ImageURL = urls[0]
				}
			}
			workflowRef, workspaceID, changed = task.WorkflowRef, task.WorkspaceID, true
			if err := tx.Save(&task).Error; err != nil {
				return err
			}
		case strings.HasPrefix(taskID, "wf-video-"):
			var task model.VideoTask
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, "id = ?", taskID).Error; err != nil {
				return err
			}
			if task.Status == "completed" || task.Status == "failed" {
				return nil
			}
			task.Status, task.Error, task.UpdatedAt, task.CompletedAt = status, failure, current, current
			if status == "completed" {
				task.Progress = 100
				if len(urls) > 0 {
					task.VideoURL = urls[0]
				}
			}
			workflowRef, workspaceID, changed = task.WorkflowRef, task.WorkspaceID, true
			if err := tx.Save(&task).Error; err != nil {
				return err
			}
		case strings.HasPrefix(taskID, "wf-audio-"):
			var task model.CanvasAudioTask
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, "id = ?", taskID).Error; err != nil {
				return err
			}
			if task.Status == "completed" || task.Status == "failed" {
				return nil
			}
			task.Status, task.Error, task.UpdatedAt, task.CompletedAt = status, failure, current, current
			if status == "completed" {
				task.Progress = 100
				if len(urls) > 0 {
					task.AudioURL = urls[0]
				}
			}
			workflowRef, workspaceID, changed = task.WorkflowRef, task.WorkspaceID, true
			if err := tx.Save(&task).Error; err != nil {
				return err
			}
		default:
			return errors.New("工作流任务 ID 无效")
		}
		return nil
	})
	return changed, workflowRef, workspaceID, err
}
