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

type workspaceModelConfigInput struct {
	LocalChannels []localModelChannelInput `json:"localChannels"`
}

type localModelChannelInput struct {
	ID       string   `json:"id"`
	Protocol string   `json:"protocol"`
	Name     string   `json:"name"`
	BaseURL  string   `json:"baseUrl"`
	APIKey   string   `json:"apiKey"`
	Models   []string `json:"models"`
}

func SelectLocalModelChannelForModel(workspaceID string, modelName string, channelID string) (model.ModelChannel, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	modelName = strings.TrimSpace(modelName)
	channelID = strings.TrimSpace(channelID)
	if workspaceID == "" {
		return model.ModelChannel{}, errors.New("缺少工作区标识")
	}
	if modelName == "" {
		return model.ModelChannel{}, errors.New("缺少模型名称")
	}
	if channelID == "" {
		return model.ModelChannel{}, errors.New("缺少模型渠道")
	}
	config, ok, err := repository.GetWorkspaceConfig(workspaceID)
	if err != nil {
		return model.ModelChannel{}, err
	}
	if !ok || strings.TrimSpace(config.ModelConfig) == "" {
		return model.ModelChannel{}, errors.New("本地渠道不存在")
	}
	var modelConfig workspaceModelConfigInput
	if err := json.Unmarshal([]byte(config.ModelConfig), &modelConfig); err != nil {
		return model.ModelChannel{}, err
	}
	for _, channel := range modelConfig.LocalChannels {
		if strings.TrimSpace(channel.ID) != channelID {
			continue
		}
		baseURL := strings.TrimSpace(channel.BaseURL)
		apiKey := strings.TrimSpace(channel.APIKey)
		if baseURL == "" || apiKey == "" {
			return model.ModelChannel{}, errors.New("本地渠道配置不完整")
		}
		models := localChannelModels(channel.Models)
		if len(models) > 0 && !localChannelHasModel(models, modelName) {
			return model.ModelChannel{}, errors.New("本地渠道不支持该模型")
		}
		protocol := strings.ToLower(strings.TrimSpace(channel.Protocol))
		if protocol == "" {
			protocol = "openai"
		}
		return model.ModelChannel{
			ID:       channelID,
			Protocol: protocol,
			Name:     firstVideoTaskValue(strings.TrimSpace(channel.Name), "本地直连"),
			BaseURL:  baseURL,
			APIKey:   apiKey,
			Models:   models,
			Weight:   1,
			Timeout:  600,
			Enabled:  true,
		}, nil
	}
	return model.ModelChannel{}, errors.New("本地渠道不存在")
}

func localChannelModels(models []string) []string {
	result := make([]string, 0, len(models))
	seen := map[string]bool{}
	for _, item := range models {
		modelName := strings.TrimSpace(item)
		if modelName == "" || seen[modelName] {
			continue
		}
		result = append(result, modelName)
		seen[modelName] = true
	}
	return result
}

func localChannelHasModel(models []string, modelName string) bool {
	for _, item := range models {
		if strings.EqualFold(strings.TrimSpace(item), modelName) {
			return true
		}
	}
	return false
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
		result.ModelConfig = json.RawMessage(config.ModelConfig)
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
	config.ModelConfig = string(raw)
	config.UpdatedAt = current
	if _, err := repository.SaveWorkspaceConfig(config); err != nil {
		return WorkspaceConfigPayload{}, err
	}
	return CurrentWorkspaceConfig(ctx)
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
