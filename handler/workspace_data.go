package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tigerowo/infinite-canvas/service"
)

func WorkspaceConfig(w http.ResponseWriter, r *http.Request) {
	config, err := service.CurrentWorkspaceConfig(r.Context())
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, config)
}

func SaveWorkspaceModelConfig(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Config json.RawMessage `json:"config"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 129<<20)).Decode(&request); err != nil || len(request.Config) == 0 {
		Fail(w, "配置内容不能为空")
		return
	}
	config, err := service.SaveCurrentWorkspaceModelConfig(r.Context(), request.Config)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, config)
}

func DeleteWorkspaceCanvasProjects(w http.ResponseWriter, r *http.Request) {
	var request struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil ||
		len(request.IDs) == 0 {
		Fail(w, "画布项目参数无效")
		return
	}
	if err := service.DeleteCurrentWorkspaceCanvasProjects(
		r.Context(),
		request.IDs,
	); err != nil {
		Fail(w, err.Error())
		return
	}
	OK(w, map[string]bool{"deleted": true})
}

func WorkspaceImageHistory(w http.ResponseWriter, r *http.Request) {
	data, err := service.CurrentWorkspaceImageHistory(r.Context())
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, json.RawMessage(data))
}

func SaveWorkspaceImageHistory(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.Data) == 0 {
		Fail(w, "数据内容不能为空")
		return
	}
	data, err := service.SaveCurrentWorkspaceImageHistory(r.Context(), request.Data)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, json.RawMessage(data))
}

func WorkspaceAssetData(w http.ResponseWriter, r *http.Request) {
	data, err := service.CurrentWorkspaceAssetData(r.Context())
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, json.RawMessage(data))
}

func SaveWorkspaceAssetData(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.Data) == 0 {
		Fail(w, "数据内容不能为空")
		return
	}
	data, err := service.SaveCurrentWorkspaceAssetData(r.Context(), request.Data)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, json.RawMessage(data))
}
