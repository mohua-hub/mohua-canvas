package repository

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"github.com/tigerowo/infinite-canvas/config"
	"github.com/tigerowo/infinite-canvas/model"
	"gorm.io/gorm"
)

var promptCategories = []model.PromptCategory{
	{Category: "system", Name: "系统", Description: "系统提示词分类"},
	{Category: "gpt-image-2-prompts", Name: "GPT Image 2 Prompts", Description: "TigerOWO 的 GPT Image 2 案例提示词分类", GithubURL: "https://github.com/tigerowo/awesome-gpt-image-2-prompts", Remote: true},
	{Category: "awesome-gpt-image", Name: "Awesome GPT Image", Description: "ZeroLu 的中文 GPT Image 提示词分类", GithubURL: "https://github.com/ZeroLu/awesome-gpt-image", Remote: true},
	{Category: "awesome-gpt4o-image-prompts", Name: "Awesome GPT4o Image Prompts", Description: "ImgEdify 的 GPT-4o 图像提示词分类", GithubURL: "https://github.com/ImgEdify/Awesome-GPT4o-Image-Prompts", Remote: true},
	{Category: "xianyu-awesome-gptimage2", Name: "Xianyu Awesome GPT Image 2", Description: "xianyu110 的 GPT Image 2 提示词分类", GithubURL: "https://github.com/xianyu110/awesome-gptimage2", Remote: true},
	{Category: "youmind-gpt-image-2", Name: "YouMind GPT Image 2", Description: "YouMind OpenLab 的 GPT Image 2 中文提示词分类", GithubURL: "https://github.com/YouMind-OpenLab/awesome-gpt-image-2", Remote: true},
	{Category: "youmind-nano-banana-pro", Name: "YouMind Nano Banana Pro", Description: "YouMind OpenLab 的 Nano Banana Pro 中文提示词分类", GithubURL: "https://github.com/YouMind-OpenLab/awesome-nano-banana-pro-prompts", Remote: true},
	{Category: "davidwu-gpt-image2-prompts", Name: "awesome-gpt-image2-prompts", Description: "davidwuw0811-boop 整理的 GPT Image 2 提示词分类", GithubURL: "https://github.com/davidwuw0811-boop/awesome-gpt-image2-prompts", Remote: true},
}

var (
	db     *gorm.DB
	dbOnce sync.Once
	dbErr  error
)

// DB 初始化并返回全局数据库连接。
func DB() (*gorm.DB, error) {
	dbOnce.Do(func() {
		dsn := config.Cfg.DatabaseDSN
		if dsn != ":memory:" {
			_ = os.MkdirAll(filepath.Dir(dsn), 0755)
		}
		db, dbErr = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if dbErr != nil {
			return
		}
		dbErr = db.AutoMigrate(
			&model.Prompt{},
			&model.AgentSkill{},
			&model.AgentSkillFile{},
			&model.Asset{},
			&model.Setting{},
			&model.CreativeWorkflow{},
			&model.WorkspaceConfig{},
			&model.AICallLog{},
			&model.StorageObject{},
			&model.VideoTask{},
			&model.ImageGenerationLog{},
			&model.CanvasImageTask{},
			&model.CanvasAudioTask{},
			&model.CanvasProject{},
			&model.ComfyBridge{},
			&model.ComfyBridgeRequest{},
		)
	})
	return db, dbErr
}

