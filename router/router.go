package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tigerowo/infinite-canvas/handler"
	"github.com/tigerowo/infinite-canvas/middleware"
)

func New() *gin.Engine {
	router := gin.Default()
	router.RedirectTrailingSlash = false
	_ = router.SetTrustedProxies(nil)
	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	api.GET("/settings", gin.WrapF(handler.Settings))
	api.GET("/storage/config", gin.WrapF(handler.StorageConfig))
	api.GET("/files/:id", func(c *gin.Context) {
		handler.FileInfo(c.Writer, c.Request, c.Param("id"))
	})
	api.GET("/files/:id/content", func(c *gin.Context) {
		handler.FileContent(c.Writer, c.Request, c.Param("id"))
	})
	api.POST("/ai/direct-request", gin.WrapF(handler.PrepareDirectAIRequest))
	api.POST("/ai/autodl/workflows", gin.WrapF(handler.AutoDLWorkflows))
	v1 := api.Group("/v1")
	v1.POST("/images/generations", gin.WrapF(handler.AIImagesGenerations))
	v1.POST("/images/edits", gin.WrapF(handler.AIImagesEdits))
	v1.POST("/responses", gin.WrapF(handler.AIResponses))
	v1.POST("/chat/completions", gin.WrapF(handler.AIChatCompletions))
	v1.POST("/audio/speech", gin.WrapF(handler.AIAudioSpeech))
	v1.GET("/tts/voices", gin.WrapF(handler.AITTSVoices))
	v1.POST("/canvas/tasks/delete", gin.WrapF(handler.DeleteWorkspaceCanvasTasks))
	v1.POST("/canvas/image-tasks", gin.WrapF(handler.CreateCanvasImageTask))
	v1.GET("/canvas/image-tasks", gin.WrapF(handler.WorkspaceCanvasImageTasks))
	v1.POST("/canvas/image-tasks/status", gin.WrapF(handler.BatchCanvasImageTasks))
	v1.GET("/canvas/image-tasks/:id", func(c *gin.Context) {
		handler.GetCanvasImageTask(c.Writer, c.Request, c.Param("id"))
	})
	v1.DELETE("/canvas/image-tasks/:id", func(c *gin.Context) {
		handler.DeleteWorkspaceCanvasImageTask(c.Writer, c.Request, c.Param("id"))
	})
	v1.POST("/canvas/audio-tasks", gin.WrapF(handler.CreateCanvasAudioTask))
	v1.GET("/canvas/audio-tasks/:id", func(c *gin.Context) {
		handler.GetCanvasAudioTask(c.Writer, c.Request, c.Param("id"))
	})
	v1.POST("/ai-logs", gin.WrapF(handler.ClientAICallLog))
	v1.POST("/videos", gin.WrapF(handler.AIVideos))
	v1.GET("/video-tasks", gin.WrapF(handler.WorkspaceVideoTasks))
	v1.DELETE("/video-tasks/:id", func(c *gin.Context) {
		handler.DeleteWorkspaceVideoTask(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/videos/:id", func(c *gin.Context) {
		handler.AIVideo(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/videos/:id/content", func(c *gin.Context) {
		handler.AIVideoContent(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/workflows", gin.WrapF(handler.WorkspaceWorkflows))
	v1.POST("/workflows", gin.WrapF(handler.SaveWorkspaceWorkflow))
	v1.POST("/workflows/agent-draft", gin.WrapF(handler.DraftWorkspaceWorkflow))
	v1.DELETE("/workflows/:id", func(c *gin.Context) {
		handler.DeleteWorkspaceWorkflow(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/agent-skills", gin.WrapF(handler.WorkspaceAgentSkills))
	v1.POST("/agent-skills", gin.WrapF(handler.SaveWorkspaceAgentSkill))
	v1.DELETE("/agent-skills/:id", func(c *gin.Context) {
		handler.DeleteWorkspaceAgentSkill(c.Writer, c.Request, c.Param("id"))
	})
	v1.POST("/storage/measure", gin.WrapF(handler.MeasureCustomStorageProvider))
	v1.POST("/files", gin.WrapF(handler.UploadFile))
	v1.POST("/files/direct", gin.WrapF(handler.RegisterDirectFile))
	v1.DELETE("/files/:id", func(c *gin.Context) {
		handler.DeleteFile(c.Writer, c.Request, c.Param("id"))
	})
	v1.DELETE("/files/:id/record", func(c *gin.Context) {
		handler.DeleteDirectFileRecord(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/config", gin.WrapF(handler.WorkspaceConfig))
	v1.POST("/config/model", gin.WrapF(handler.SaveWorkspaceModelConfig))
	v1.POST("/workflow-providers/runninghub/inspect", gin.WrapF(handler.RunningHubInspect))
	v1.GET("/runninghub/collection", gin.WrapF(handler.RunningHubCollection))
	v1.POST("/runninghub/entry", gin.WrapF(handler.RunningHubCanvasEntry))
	v1.POST("/runninghub/tasks", gin.WrapF(handler.CreateRunningHubCanvasTask))
	v1.GET("/runninghub/tasks/:id", func(c *gin.Context) { handler.GetRunningHubCanvasTask(c.Writer, c.Request, c.Param("id")) })
	v1.POST("/workflow-tasks", gin.WrapF(handler.CreateWorkflowTask))
	v1.GET("/workflow-tasks/:id", func(c *gin.Context) { handler.GetWorkflowTask(c.Writer, c.Request, c.Param("id")) })
	v1.GET("/comfy-bridges", gin.WrapF(handler.WorkspaceComfyBridges))
	v1.POST("/comfy-bridges", gin.WrapF(handler.WorkspaceComfyBridges))
	v1.POST("/comfy-bridges/inspect", gin.WrapF(handler.WorkspaceComfyBridgeInspect))
	v1.DELETE("/comfy-bridges/:id", func(c *gin.Context) { handler.WorkspaceDeleteComfyBridge(c.Writer, c.Request, c.Param("id")) })
	v1.POST("/config/storage", gin.WrapF(handler.SaveCustomStorageProvider))
	v1.GET("/canvas/projects", gin.WrapF(handler.WorkspaceCanvasProjects))
	v1.POST("/canvas/projects", gin.WrapF(handler.SaveWorkspaceCanvasProject))
	v1.POST("/canvas/projects/sync", gin.WrapF(handler.SyncWorkspaceCanvasProjects))
	v1.POST("/canvas/projects/delete", gin.WrapF(handler.DeleteWorkspaceCanvasProjects))
	v1.GET("/data/image-history", gin.WrapF(handler.WorkspaceImageHistory))
	v1.POST("/data/image-history", gin.WrapF(handler.SaveWorkspaceImageHistory))
	v1.GET("/generation-logs/videos", gin.WrapF(handler.WorkspaceVideoGenerationLogs))
	v1.POST("/generation-logs/videos", gin.WrapF(handler.SaveWorkspaceVideoGenerationLogs))
	v1.POST("/generation-logs/videos/delete", gin.WrapF(handler.DeleteWorkspaceVideoGenerationLogs))
	v1.DELETE("/generation-logs/videos/:id", func(c *gin.Context) {
		handler.DeleteWorkspaceVideoGenerationLog(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/generation-logs/images", gin.WrapF(handler.WorkspaceImageGenerationLogs))
	v1.POST("/generation-logs/images", gin.WrapF(handler.SaveWorkspaceImageGenerationLogs))
	v1.POST("/generation-logs/images/delete", gin.WrapF(handler.DeleteWorkspaceImageGenerationLogs))
	v1.DELETE("/generation-logs/images/:id", func(c *gin.Context) {
		handler.DeleteWorkspaceImageGenerationLog(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/data/assets", gin.WrapF(handler.WorkspaceAssetData))
	v1.POST("/data/assets", gin.WrapF(handler.SaveWorkspaceAssetData))
	api.GET("/proxy-image", gin.WrapF(handler.ProxyImage))
	api.GET("/prompts", gin.WrapF(handler.Prompts))
	api.GET("/agent-skills", gin.WrapF(handler.AgentSkills))
	api.GET("/agent-skills/:id/file", func(c *gin.Context) {
		handler.AgentSkillFile(c.Writer, c.Request, c.Param("id"))
	})
	api.GET("/assets", gin.WrapF(handler.Assets))

	settings := api.Group("/settings")
	settings.GET("/ai-logs", gin.WrapF(handler.AdminAICallLogs))
	settings.DELETE("/ai-logs", gin.WrapF(handler.AdminDeleteAICallLogs))
	settings.GET("/config", gin.WrapF(handler.AdminSettings))
	settings.POST("/config", gin.WrapF(handler.AdminSaveSettings))
	settings.POST("/config/channel-models", gin.WrapF(handler.AdminChannelModels))
	settings.POST("/config/channel-test", gin.WrapF(handler.AdminTestChannelModel))
	settings.POST("/workflow-providers/runninghub/inspect", gin.WrapF(handler.AdminRunningHubInspect))
	settings.GET("/comfy-bridges", gin.WrapF(handler.AdminComfyBridges))
	settings.POST("/comfy-bridges", gin.WrapF(handler.AdminComfyBridges))
	settings.POST("/comfy-bridges/inspect", gin.WrapF(handler.AdminComfyBridgeInspect))
	settings.DELETE("/comfy-bridges/:id", func(c *gin.Context) { handler.AdminDeleteComfyBridge(c.Writer, c.Request, c.Param("id")) })
	bridge := api.Group("/bridge/comfy")
	bridge.POST("/heartbeat", gin.WrapF(handler.BridgeComfyHeartbeat))
	bridge.GET("/poll", gin.WrapF(handler.BridgeComfyPoll))
	bridge.POST("/lease", gin.WrapF(handler.BridgeComfyLease))
	bridge.POST("/result", gin.WrapF(handler.BridgeComfyResult))
	settings.POST("/storage/measure", gin.WrapF(handler.AdminMeasureStorageProvider))
	settings.GET("/prompt-categories", gin.WrapF(handler.AdminPromptCategories))
	settings.POST("/prompt-categories/sync", gin.WrapF(handler.AdminSyncPromptCategories))
	settings.POST("/prompt-categories/sync-all", gin.WrapF(handler.AdminSyncAllPromptCategories))
	settings.GET("/prompts", gin.WrapF(handler.AdminPrompts))
	settings.POST("/prompts", gin.WrapF(handler.AdminSavePrompt))
	settings.POST("/prompts/batch-delete", gin.WrapF(handler.AdminDeletePrompts))
	settings.DELETE("/prompts/:id", func(c *gin.Context) {
		handler.AdminDeletePrompt(c.Writer, c.Request, c.Param("id"))
	})
	settings.GET("/agent-skills", gin.WrapF(handler.AdminAgentSkills))
	settings.GET("/agent-skills/:id/files", func(c *gin.Context) {
		handler.AdminAgentSkillFiles(c.Writer, c.Request, c.Param("id"))
	})
	settings.POST("/agent-skills", gin.WrapF(handler.AdminSaveAgentSkill))
	settings.DELETE("/agent-skills/:id", func(c *gin.Context) {
		handler.AdminDeleteAgentSkill(c.Writer, c.Request, c.Param("id"))
	})
	settings.GET("/assets", gin.WrapF(handler.AdminAssets))
	settings.POST("/assets", gin.WrapF(handler.AdminSaveAsset))
	settings.DELETE("/assets/:id", func(c *gin.Context) {
		handler.AdminDeleteAsset(c.Writer, c.Request, c.Param("id"))
	})

	router.NoRoute(middleware.NotFoundJSON)

	return router
}
