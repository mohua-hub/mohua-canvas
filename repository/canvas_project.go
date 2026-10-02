package repository

import (
	"errors"
	"strings"

	"github.com/tigerowo/infinite-canvas/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ListWorkspaceCanvasProjects(workspaceID string) ([]model.CanvasProject, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	var projects []model.CanvasProject
	err = db.Where(
		"workspace_id = ? AND deleted_at = ''",
		strings.TrimSpace(workspaceID),
	).Order("updated_at DESC").Find(&projects).Error
	return projects, err
}

func SaveWorkspaceCanvasProject(
	project model.CanvasProject,
) (model.CanvasProject, error) {
	db, err := DB()
	if err != nil {
		return project, err
	}

	project.WorkspaceID = strings.TrimSpace(project.WorkspaceID)
	project.ID = strings.TrimSpace(project.ID)

	var current model.CanvasProject
	err = db.First(
		&current,
		"workspace_id = ? AND id = ?",
		project.WorkspaceID,
		project.ID,
	).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return project, db.Create(&project).Error
	}
	if err != nil {
		return project, err
	}
	if current.DeletedAt != "" || current.UpdatedAt > project.UpdatedAt {
		return current, nil
	}

	result := db.Model(&model.CanvasProject{}).
		Where(
			"workspace_id = ? AND id = ? AND deleted_at = '' AND updated_at <= ?",
			project.WorkspaceID,
			project.ID,
			project.UpdatedAt,
		).
		Updates(map[string]any{
			"project_data": project.ProjectData,
			"updated_at":   project.UpdatedAt,
		})
	if result.Error != nil {
		return project, result.Error
	}
	if result.RowsAffected == 0 {
		if err := db.First(
			&current,
			"workspace_id = ? AND id = ?",
			project.WorkspaceID,
			project.ID,
		).Error; err != nil {
			return project, err
		}
		return current, nil
	}
	return project, nil
}

func SaveWorkspaceCanvasProjects(
	workspaceID string,
	projects []model.CanvasProject,
) ([]model.CanvasProject, error) {
	for _, project := range projects {
		project.WorkspaceID = workspaceID
		if _, err := SaveWorkspaceCanvasProject(project); err != nil {
			return nil, err
		}
	}
	return ListWorkspaceCanvasProjects(workspaceID)
}

func SoftDeleteWorkspaceCanvasProjects(
	workspaceID string,
	ids []string,
	deletedAt string,
) error {
	ids = uniqueTrimmedValues(ids...)
	if len(ids) == 0 {
		return nil
	}

	db, err := DB()
	if err != nil {
		return err
	}

	records := make([]model.CanvasProject, 0, len(ids))
	for _, id := range ids {
		records = append(records, model.CanvasProject{
			WorkspaceID:    strings.TrimSpace(workspaceID),
			ID:        id,
			CreatedAt: deletedAt,
			UpdatedAt: deletedAt,
			DeletedAt: deletedAt,
		})
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "workspace_id"},
			{Name: "id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"project_data",
			"updated_at",
			"deleted_at",
		}),
	}).Create(&records).Error
}

func CleanupDeletedCanvasProjects(before string) error {
	db, err := DB()
	if err != nil {
		return err
	}

	return db.Where(
		"deleted_at <> '' AND deleted_at < ?",
		before,
	).Delete(&model.CanvasProject{}).Error
}
