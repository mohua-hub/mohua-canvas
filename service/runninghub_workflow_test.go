package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tigerowo/infinite-canvas/model"
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
		if r.Method != http.MethodPost || r.URL.Path != "/openapi/v2/media/upload/binary" || r.Header.Get("Authorization") != "Bearer 企业密钥" {
			t.Fatalf("请求 = %s，Authorization = %q", r.Method, r.Header.Get("Authorization"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("apiKey") != "" || r.FormValue("fileType") != "" || len(r.MultipartForm.File["file"]) != 1 {
			t.Fatalf("上传表单不符合 V2 接口：%#v", r.MultipartForm)
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

func TestRunningHubV2QueryContract(t *testing.T) {
	for _, test := range []struct {
		name     string
		response string
		done     bool
		terminal bool
	}{
		{"排队", `{"taskId":"task-1","status":"QUEUED","results":null}`, false, false},
		{"运行", `{"taskId":"task-1","status":"RUNNING","results":null}`, false, false},
		{"完成", `{"taskId":"task-1","status":"SUCCESS","results":[{"url":"https://example.com/result.png","outputType":"png"}]}`, true, false},
		{"失败", `{"taskId":"task-1","status":"FAILED","errorCode":"WORKFLOW_ERROR","errorMessage":"参数错误","results":null}`, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := withRunningHubTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/openapi/v2/query" || r.Header.Get("Authorization") != "Bearer test-key" {
					t.Fatalf("请求不符合 V2 查询接口：%s %s", r.Method, r.URL.Path)
				}
				var body map[string]any
				if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 1 || body["taskId"] != "task-1" {
					t.Fatalf("请求体 = %#v", body)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(test.response))
			}))
			urls, done, err := PollRunningHubTask(context.Background(), model.ModelChannel{BaseURL: server.URL, APIKey: "test-key"}, "task-1", nil)
			if done != test.done || errors.Is(err, errRunningHubTaskTerminal) != test.terminal || !test.terminal && err != nil {
				t.Fatalf("done=%v urls=%v err=%v", done, urls, err)
			}
			if test.done && (len(urls) != 1 || urls[0] != "https://example.com/result.png") {
				t.Fatalf("产物 = %v", urls)
			}
			if test.terminal && !strings.Contains(err.Error(), "参数错误") {
				t.Fatalf("失败原因未保留：%v", err)
			}
		})
	}
}

func TestRunningHubAppFieldDataContract(t *testing.T) {
	var response map[string]any
	if err := json.Unmarshal([]byte(`{"code":0,"data":{"webappName":"图像编辑","nodeInfoList":[{"nodeId":"37","fieldName":"model","fieldType":"LIST","fieldValue":"pro","description":"模型切换","fieldData":"[{\"name\":\"Pro\",\"index\":\"pro\",\"description\":\"默认模型\"},{\"name\":\"Max\",\"index\":\"max\"},{\"default\":\"pro\"}]"},{"nodeId":"38","fieldName":"steps","fieldType":"INT","fieldValue":20,"fieldData":"[\"INT\",{\"min\":1,\"max\":50,\"step\":1}]"}]}}`), &response); err != nil {
		t.Fatal(err)
	}
	entry, err := runningHubAppEntry(RunningHubInspectInput{WorkflowID: "123", Capability: "image"}, response)
	if err != nil || entry.Title != "图像编辑" || len(entry.Fields) != 2 {
		t.Fatalf("条目 = %#v err=%v", entry, err)
	}
	if entry.Fields[0].Label != "模型切换" || len(entry.Fields[0].Options) != 2 {
		t.Fatalf("应用枚举配置 = %#v", entry.Fields[0])
	}
	if _, err := validateWorkflowFieldValue(entry.Fields[0], "max"); err != nil {
		t.Fatalf("官方枚举值不被接受：%v", err)
	}
	if _, err := validateWorkflowFieldValue(entry.Fields[1], 51); err == nil {
		t.Fatal("数值范围没有生效")
	}
}
