package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tigerowo/infinite-canvas/service"
)

func WorkspaceVideoGenerationLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := service.CurrentWorkspaceVideoGenerationLogs(r.Context())
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, logs)
}

func SaveWorkspaceVideoGenerationLogs(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Logs []json.RawMessage `json:"logs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		Fail(w, "历史记录不能为空")
		return
	}
	logs, err := service.SaveCurrentWorkspaceVideoGenerationLogs(r.Context(), request.Logs)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, logs)
}

func DeleteWorkspaceVideoGenerationLog(w http.ResponseWriter, r *http.Request, id string) {
	if strings.TrimSpace(id) == "" {
		Fail(w, "删除历史记录参数无效")
		return
	}
	if err := service.DeleteCurrentWorkspaceVideoGenerationLog(r.Context(), id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, map[string]bool{"deleted": true})
}

func DeleteWorkspaceVideoGenerationLogs(w http.ResponseWriter, r *http.Request) {
	var request struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		Fail(w, "删除历史记录参数无效")
		return
	}
	if len(request.IDs) == 0 {
		OK(w, map[string]bool{"deleted": true})
		return
	}
	if err := service.DeleteCurrentWorkspaceVideoGenerationLogs(r.Context(), request.IDs); err != nil {
		FailError(w, err)
		return
	}
	OK(w, map[string]bool{"deleted": true})
}

func WorkspaceImageGenerationLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := service.CurrentWorkspaceImageGenerationLogs(r.Context())
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, logs)
}

func SaveWorkspaceImageGenerationLogs(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Logs []json.RawMessage `json:"logs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		Fail(w, "历史记录不能为空")
		return
	}
	logs, err := service.SaveCurrentWorkspaceImageGenerationLogs(r.Context(), request.Logs)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, logs)
}

func DeleteWorkspaceImageGenerationLog(w http.ResponseWriter, r *http.Request, id string) {
	if strings.TrimSpace(id) == "" {
		Fail(w, "历史记录不存在")
		return
	}
	if err := service.DeleteCurrentWorkspaceImageGenerationLog(r.Context(), id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, map[string]bool{"deleted": true})
}

func DeleteWorkspaceImageGenerationLogs(w http.ResponseWriter, r *http.Request) {
	var request struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		Fail(w, "删除历史记录参数无效")
		return
	}
	if len(request.IDs) == 0 {
		OK(w, map[string]bool{"deleted": true})
		return
	}
	if err := service.DeleteCurrentWorkspaceImageGenerationLogs(r.Context(), request.IDs); err != nil {
		FailError(w, err)
		return
	}
	OK(w, map[string]bool{"deleted": true})
}
