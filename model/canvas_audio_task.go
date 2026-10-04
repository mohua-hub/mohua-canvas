package model

type CanvasAudioTask struct {
	ID              string  `json:"id" gorm:"primaryKey"`
	WorkspaceID          string  `json:"workspaceId" gorm:"index:idx_canvas_audio_tasks_workspace_source_node,priority:1"`
	Source          string  `json:"source" gorm:"index:idx_canvas_audio_tasks_workspace_source_node,priority:2"`
	SourceID        string  `json:"sourceId" gorm:"index:idx_canvas_audio_tasks_workspace_source_node,priority:3"`
	NodeID          string  `json:"nodeId" gorm:"index:idx_canvas_audio_tasks_workspace_source_node,priority:4"`
	Model           string  `json:"model"`
	ChannelID       string  `json:"channelId"`
	ChannelName     string  `json:"channelName"`
	WorkflowRef     string  `json:"workflowRef,omitempty" gorm:"type:text"`
	Status          string  `json:"status"`
	Progress        int     `json:"progress"`
	Prompt          string  `json:"prompt" gorm:"type:text"`
	Endpoint        string  `json:"endpoint"`
	ContentType     string  `json:"contentType"`
	RequestBody     string  `json:"requestBody" gorm:"type:text"`
	ResponseBody    string  `json:"responseBody" gorm:"type:text"`
	Error           string  `json:"error" gorm:"type:text"`
	ErrorDetail     string  `json:"errorDetail" gorm:"type:text"`
	AudioURL        string  `json:"audioUrl" gorm:"type:text"`
	StorageKey      string  `json:"storageKey"`
	MimeType        string  `json:"mimeType"`
	Bytes           int64   `json:"bytes"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
	StartedAt       string  `json:"startedAt"`
	CompletedAt     string  `json:"completedAt"`
}
