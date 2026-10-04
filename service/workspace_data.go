package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
)

type WorkspaceConfigPayload struct {
	ModelConfig      json.RawMessage       `json:"modelConfig,omitempty"`
	StorageProvider  *CustomStorageProviders `json:"storageProvider,omitempty"`
	ImageHistory     json.RawMessage       `json:"imageHistory,omitempty"`
	AssetData        json.RawMessage       `json:"assetData,omitempty"`
	SyncCapabilities map[string]bool       `json:"syncCapabilities,omitempty"`
}

type StorageObjectProviderInput struct {
	Enabled         *bool  `json:"enabled,omitempty"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	PublicBaseURL   string `json:"publicBaseUrl"`
	PathPrefix      string `json:"pathPrefix"`
	Username        string `json:"username"`
	Password        string `json:"password"`
}

type CustomStorageProviders struct {
	S3     *StorageObjectProviderInput `json:"s3,omitempty"`
	WebDAV *StorageObjectProviderInput `json:"webdav,omitempty"`
}

type workspaceWorkflowChannel struct {
	ChannelID string                `json:"channelId"`
	Protocol  string                `json:"protocol"`
	Workflows []model.WorkflowEntry `json:"workflows"`
}

func personalWorkflowChannels(workspaceID string) ([]workspaceWorkflowChannel, error) {
	stored, exists, err := repository.GetWorkspaceConfig(workspaceID)
	if err != nil || !exists || strings.TrimSpace(stored.ModelConfig) == "" {
		return nil, err
	}
	var config struct {
		WorkflowChannels []workspaceWorkflowChannel `json:"workflowChannels"`
	}
	if err := json.Unmarshal([]byte(stored.ModelConfig), &config); err != nil {
		return nil, errors.New("本地工作流配置格式无效")
	}
	return config.WorkflowChannels, nil
}

func CurrentWorkspaceConfig(ctx context.Context) (WorkspaceConfigPayload, error) {
	workspaceID := WorkspaceID
	config, ok, err := repository.GetWorkspaceConfig(workspaceID)
	if err != nil {
		return WorkspaceConfigPayload{}, err
	}
	result := WorkspaceConfigPayload{
		SyncCapabilities: map[string]bool{
			"workspaceData":  true,
			"workflows": true,
			"assets":    true,
		},
	}
	if !ok {
		return result, nil
	}
	if strings.TrimSpace(config.ModelConfig) != "" {
		cleanConfig, err := sanitizeWorkspaceModelConfig(json.RawMessage(config.ModelConfig))
		if err != nil {
			return WorkspaceConfigPayload{}, err
		}
		result.ModelConfig = cleanConfig
	}
	if strings.TrimSpace(config.StorageProvider) != "" {
		providers := readCustomStorageProviders(config.StorageProvider)
		var syncFlags struct {
			SyncStorageConfig       bool `json:"syncStorageConfig"`
			SyncWebDAVStorageConfig bool `json:"syncWebDAVStorageConfig"`
		}
		_ = json.Unmarshal(result.ModelConfig, &syncFlags)
		if !syncFlags.SyncStorageConfig {
			providers.S3 = nil
		}
		if !syncFlags.SyncWebDAVStorageConfig {
			providers.WebDAV = nil
		}
		if providers.S3 != nil || providers.WebDAV != nil {
			result.StorageProvider = &providers
		}
	}
	if strings.TrimSpace(config.ImageHistory) != "" {
		result.ImageHistory = json.RawMessage(config.ImageHistory)
	}
	if strings.TrimSpace(config.AssetData) != "" {
		result.AssetData = json.RawMessage(config.AssetData)
	}
	return result, nil
}

func readCustomStorageProviders(raw string) CustomStorageProviders {
	var providers CustomStorageProviders
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &providers)
	}
	return providers
}

func SaveCurrentWorkspaceModelConfig(ctx context.Context, raw json.RawMessage) (WorkspaceConfigPayload, error) {
	cleanConfig, err := sanitizeWorkspaceModelConfig(raw)
	if err != nil {
		return WorkspaceConfigPayload{}, err
	}
	workspaceID := WorkspaceID
	config, _, err := repository.GetWorkspaceConfig(workspaceID)
	if err != nil {
		return WorkspaceConfigPayload{}, err
	}
	current := now()
	if config.WorkspaceID == "" {
		config.WorkspaceID = workspaceID
		config.CreatedAt = current
	}
	config.ModelConfig = string(cleanConfig)
	config.UpdatedAt = current
	if _, err := repository.SaveWorkspaceConfig(config); err != nil {
		return WorkspaceConfigPayload{}, err
	}
	return CurrentWorkspaceConfig(ctx)
}

func sanitizeWorkspaceModelConfig(raw json.RawMessage) (json.RawMessage, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if config, ok := value.(map[string]any); ok {
		for key := range config {
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "channelmode", "baseurl", "apikey", "api_key", "localchannels", "publicchannels":
				delete(config, key)
			}
		}
	}
	return json.Marshal(value)
}

func CurrentWorkspaceImageHistory(ctx context.Context) (json.RawMessage, error) {
	config, err := currentWorkspaceConfig(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.ImageHistory) == "" {
		return json.RawMessage(`{"logs":[],"categories":[]}`), nil
	}
	return json.RawMessage(config.ImageHistory), nil
}

func SaveCurrentWorkspaceImageHistory(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	config, err := saveCurrentWorkspaceConfigField(ctx, func(config *model.WorkspaceConfig) {
		config.ImageHistory = string(raw)
	})
	if err != nil {
		return nil, err
	}
	return json.RawMessage(config.ImageHistory), nil
}

func CurrentWorkspaceAssetData(ctx context.Context) (json.RawMessage, error) {
	config, err := currentWorkspaceConfig(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.AssetData) == "" {
		return json.RawMessage(`{"assets":[]}`), nil
	}
	return json.RawMessage(config.AssetData), nil
}

func SaveCurrentWorkspaceAssetData(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	config, err := saveCurrentWorkspaceConfigField(ctx, func(config *model.WorkspaceConfig) {
		config.AssetData = string(raw)
	})
	if err != nil {
		return nil, err
	}
	return json.RawMessage(config.AssetData), nil
}

func currentWorkspaceConfig(ctx context.Context) (model.WorkspaceConfig, error) {
	workspaceID := WorkspaceID
	config, _, err := repository.GetWorkspaceConfig(workspaceID)
	if err != nil {
		return model.WorkspaceConfig{}, err
	}
	if config.WorkspaceID == "" {
		config.WorkspaceID = workspaceID
	}
	return config, nil
}

func saveCurrentWorkspaceConfigField(ctx context.Context, patch func(config *model.WorkspaceConfig)) (model.WorkspaceConfig, error) {
	workspaceID := WorkspaceID
	config, _, err := repository.GetWorkspaceConfig(workspaceID)
	if err != nil {
		return model.WorkspaceConfig{}, err
	}
	current := now()
	if config.WorkspaceID == "" {
		config.WorkspaceID = workspaceID
		config.CreatedAt = current
	}
	patch(&config)
	config.UpdatedAt = current
	return repository.SaveWorkspaceConfig(config)
}
