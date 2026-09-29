package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withRunningHubTestClient(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	previous := safeProxyHTTPClient
	safeProxyHTTPClient = server.Client()
	t.Cleanup(func() {
		safeProxyHTTPClient = previous
		server.Close()
	})
	return server
}

func TestInspectRunningHubAppUsesDocumentedRequest(t *testing.T) {
	server := withRunningHubTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("RunningHub App 参数接口使用了 %s，期望 GET", r.Method)
		}
		if r.URL.Path != "/api/webapp/apiCallDemo" {
			t.Fatalf("请求路径 = %q", r.URL.Path)
		}
		if r.URL.Query().Get("apiKey") != "积分密钥" || r.URL.Query().Get("webappId") != "123" {
			t.Fatalf("查询参数 = %v", r.URL.Query())
		}
		if r.Header.Get("Authorization") != "Bearer 积分密钥" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "success",
			"data": map[string]any{"nodeInfoList": []any{map[string]any{"nodeId": "1", "fieldName": "prompt", "fieldValue": ""}}},
		})
	}))

	entry, err := InspectRunningHub(context.Background(), RunningHubInspectInput{
		BaseURL: server.URL, APIKey: "积分密钥", Kind: "app", WorkflowID: "123", Capability: "image",
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.Kind != "app" || len(entry.Fields) != 1 || entry.Fields[0].NodeID != "1" {
		t.Fatalf("工作流条目 = %#v", entry)
	}
}

func TestRunningHubRequestAddsBearerAuthorization(t *testing.T) {
	server := withRunningHubTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer 积分密钥" {
			t.Fatalf("请求 = %s，Authorization = %q", r.Method, r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"taskId":"task-1"}}`))
	}))

	var response map[string]any
	err := runningHubRequest(context.Background(), server.URL, map[string]any{"apiKey": "积分密钥"}, &response, nil)
	if err != nil {
		t.Fatal(err)
	}
	if response["code"] != float64(0) {
		t.Fatalf("响应 = %#v", response)
	}
}

func TestUploadRunningHubReferenceAddsBearerAuthorization(t *testing.T) {
	server := withRunningHubTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer 企业密钥" {
			t.Fatalf("请求 = %s，Authorization = %q", r.Method, r.Header.Get("Authorization"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"fileName":"input/test.png"}}`))
	}))

	name, err := uploadRunningHubReference(context.Background(), server.URL, "企业密钥", "data:image/png;base64,aGVsbG8=", nil)
	if err != nil {
		t.Fatal(err)
	}
	if name != "input/test.png" {
		t.Fatalf("fileName = %q", name)
	}
}
