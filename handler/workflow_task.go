package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tigerowo/infinite-canvas/service"
)

func CreateWorkflowTask(w http.ResponseWriter, r *http.Request) {
	workspaceID := service.WorkspaceID
	var input service.WorkflowRunInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 96<<20)).Decode(&input); err != nil {
		Fail(w, "工作流请求格式无效或超过 96MB")
		return
	}
	result, err := service.CreateWorkflowTask(r.Context(), workspaceID, input)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func GetWorkflowTask(w http.ResponseWriter, r *http.Request, id string) {
	workspaceID := service.WorkspaceID
	result, err := service.GetWorkflowTask(r.Context(), workspaceID, id)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowTaskNotFound) {
			FailWithStatus(w, http.StatusNotFound, err.Error())
		} else {
			FailWithStatus(w, http.StatusInternalServerError, "查询工作流任务失败")
		}
		return
	}
	OK(w, result)
}
