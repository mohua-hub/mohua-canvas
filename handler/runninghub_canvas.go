package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tigerowo/infinite-canvas/service"
)

func RunningHubCollection(w http.ResponseWriter, r *http.Request) {
	workspaceID := service.WorkspaceID
	items, err := service.ListRunningHubCollection(workspaceID, r.URL.Query().Get("scope"))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, items)
}

func RunningHubCanvasEntry(w http.ResponseWriter, r *http.Request) {
	workspaceID := service.WorkspaceID
	var ref service.WorkflowRef
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&ref); err != nil {
		Fail(w, "RunningHub 条目请求格式无效")
		return
	}
	entry, err := service.RunningHubCanvasEntry(workspaceID, ref)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, entry)
}

func CreateRunningHubCanvasTask(w http.ResponseWriter, r *http.Request) {
	workspaceID := service.WorkspaceID
	var input service.WorkflowRunInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 96<<20)).Decode(&input); err != nil {
		Fail(w, "RunningHub 请求格式无效或超过 96MB")
		return
	}
	result, err := service.CreateRunningHubCanvasTask(r.Context(), workspaceID, input)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func GetRunningHubCanvasTask(w http.ResponseWriter, r *http.Request, id string) {
	workspaceID := service.WorkspaceID
	result, err := service.GetRunningHubCanvasTask(r.Context(), workspaceID, id)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}
