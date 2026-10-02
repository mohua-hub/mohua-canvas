package model

// WorkspaceConfig 工作区配置和同步数据。
type WorkspaceConfig struct {
	WorkspaceID          string `json:"workspaceId" gorm:"primaryKey"`
	ModelConfig     string `json:"modelConfig" gorm:"size:134217728"`
	StorageProvider string `json:"storageProvider" gorm:"type:text"`
	ImageHistory    string `json:"imageHistory" gorm:"type:text"`
	AssetData       string `json:"assetData" gorm:"type:text"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}
