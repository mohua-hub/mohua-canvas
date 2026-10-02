package repository

import "github.com/tigerowo/infinite-canvas/model"

func SaveCanvasAudioTask(task model.CanvasAudioTask) (model.CanvasAudioTask, error) {
	db, err := DB()
	if err != nil {
		return task, err
	}
	return task, db.Save(&task).Error
}

func UpdateCanvasAudioTask(task model.CanvasAudioTask) (model.CanvasAudioTask, error) {
	db, err := DB()
	if err != nil {
		return task, err
	}

	return task, db.Model(&model.CanvasAudioTask{}).
		Where("workspace_id = ? AND id = ?", task.WorkspaceID, task.ID).
		Select("*").
		Updates(&task).Error
}

func GetWorkspaceCanvasAudioTask(workspaceID string, id string) (model.CanvasAudioTask, bool, error) {
	db, err := DB()
	if err != nil {
		return model.CanvasAudioTask{}, false, err
	}
	var task model.CanvasAudioTask
	err = db.First(&task, "workspace_id = ? AND id = ?", workspaceID, id).Error
	if err != nil {
		return model.CanvasAudioTask{}, false, nil
	}
	return task, true, nil
}
