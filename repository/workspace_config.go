package repository

import (
	"errors"
	"time"

	"github.com/tigerowo/infinite-canvas/model"
	"gorm.io/gorm"
)

func workspaceConfigTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func GetWorkspaceConfig(workspaceID string) (model.WorkspaceConfig, bool, error) {
	db, err := DB()
	if err != nil {
		return model.WorkspaceConfig{}, false, err
	}

	var config model.WorkspaceConfig
	err = db.First(&config, "workspace_id = ?", workspaceID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.WorkspaceConfig{}, false, nil
	}
	if err != nil {
		return model.WorkspaceConfig{}, false, err
	}
	return config, true, nil
}

func SaveWorkspaceConfig(config model.WorkspaceConfig) (model.WorkspaceConfig, error) {
	db, err := DB()
	if err != nil {
		return config, err
	}

	config.UpdatedAt = workspaceConfigTimestamp()
	return config, db.Save(&config).Error
}
