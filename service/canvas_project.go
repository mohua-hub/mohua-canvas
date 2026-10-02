package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
)

type canvasProjectMetadata struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func canvasProjectFromRaw(
	workspaceID string,
	raw json.RawMessage,
) (model.CanvasProject, error) {
	var metadata canvasProjectMetadata
	if len(raw) == 0 || json.Unmarshal(raw, &metadata) != nil {
		return model.CanvasProject{}, errors.New("画布项目数据无效")
	}

	metadata.ID = strings.TrimSpace(metadata.ID)
	metadata.CreatedAt = strings.TrimSpace(metadata.CreatedAt)
	metadata.UpdatedAt = strings.TrimSpace(metadata.UpdatedAt)
	if metadata.ID == "" || metadata.CreatedAt == "" ||
		metadata.UpdatedAt == "" {
		return model.CanvasProject{}, errors.New("画布项目数据无效")
	}

	return model.CanvasProject{
		WorkspaceID:      strings.TrimSpace(workspaceID),
		ID:          metadata.ID,
		ProjectData: string(raw),
		CreatedAt:   metadata.CreatedAt,
		UpdatedAt:   metadata.UpdatedAt,
	}, nil
}

func canvasProjectData(
	projects []model.CanvasProject,
) []json.RawMessage {
	result := make([]json.RawMessage, 0, len(projects))
	for _, project := range projects {
		if strings.TrimSpace(project.ProjectData) != "" {
			result = append(
				result,
				json.RawMessage(project.ProjectData),
			)
		}
	}
	return result
}

func CurrentWorkspaceCanvasProjects(
	ctx context.Context,
) ([]json.RawMessage, error) {
	workspaceID := WorkspaceID

	projects, err := repository.ListWorkspaceCanvasProjects(workspaceID)
	if err != nil {
		return nil, err
	}
	return canvasProjectData(projects), nil
}

func SaveCurrentWorkspaceCanvasProject(
	ctx context.Context,
	raw json.RawMessage,
) (json.RawMessage, error) {
	workspaceID := WorkspaceID

	project, err := canvasProjectFromRaw(workspaceID, raw)
	if err != nil {
		return nil, err
	}
	saved, err := repository.SaveWorkspaceCanvasProject(project)
	if err != nil {
		return nil, err
	}
	if saved.DeletedAt != "" {
		return nil, errors.New("画布项目已删除")
	}
	return json.RawMessage(saved.ProjectData), nil
}

func SyncCurrentWorkspaceCanvasProjects(
	ctx context.Context,
	rawProjects []json.RawMessage,
) ([]json.RawMessage, error) {
	workspaceID := WorkspaceID

	projects := make([]model.CanvasProject, 0, len(rawProjects))
	for _, raw := range rawProjects {
		project, err := canvasProjectFromRaw(workspaceID, raw)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	saved, err := repository.SaveWorkspaceCanvasProjects(workspaceID, projects)
	if err != nil {
		return nil, err
	}
	return canvasProjectData(saved), nil
}

func DeleteCurrentWorkspaceCanvasProjects(
	ctx context.Context,
	projectIDs []string,
) error {
	workspaceID := WorkspaceID

	for _, projectID := range projectIDs {
		if strings.TrimSpace(projectID) != "" {
			return repository.SoftDeleteWorkspaceCanvasProjects(
				workspaceID,
				projectIDs,
				time.Now().UTC().Format(time.RFC3339Nano),
			)
		}
	}
	return errors.New("画布项目参数无效")
}
