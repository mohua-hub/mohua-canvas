package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestNormalizeKIEKlingOmniVideoInput(t *testing.T) {
	tests := []struct {
		name  string
		model string
		input map[string]any
		want  map[string]any
	}{
		{
			name: "text smart shots", model: "kling-3.0-omni/text-to-video",
			input: map[string]any{"prompt": "scene", "mode": "pro", "multi_shot": true, "shot_type": "intelligence", "multi_prompt": []any{map[string]any{"prompt": "shot", "duration": "5"}}, "image_urls": []any{"image"}, "video_urls": []any{"video"}, "negative_prompt": "bad"},
			want:  map[string]any{"prompt": "scene", "resolution": "1080p", "customize_multi_shots": false, "prefer_multi_shots": true},
		},
		{
			name: "image first and last frames", model: "kling-3.0-omni/image-to-video",
			input: map[string]any{"prompt": "scene", "mode": "4k", "aspect_ratio": "16:9", "image_urls": []any{"first", "last"}, "multi_shot": true, "shot_type": "customize", "multi_prompt": []any{map[string]any{"prompt": "shot", "duration": "20"}}},
			want:  map[string]any{"prompt": "scene", "resolution": "4k", "aspect_ratio": "auto", "image_urls": []any{"first", "last"}, "customize_multi_shots": true, "prefer_multi_shots": false, "multi_prompt": []map[string]any{{"prompt": "shot", "duration": 15}}},
		},
		{
			name: "reference video constraints", model: "kling-3.0-omni/reference-to-video",
			input: map[string]any{"prompt": "scene", "mode": "std", "aspect_ratio": "9:16", "duration": 10, "image_urls": []any{"image"}, "video_urls": []any{"video"}, "audio": true, "multi_shot": true, "shot_type": "customize", "prefer_multi_shots": true, "multi_prompt": []any{map[string]any{"prompt": "shot", "duration": "5"}}, "element_list": []any{map[string]any{"name": "role", "element_input_urls": []any{"element"}}}},
			want:  map[string]any{"prompt": "scene", "resolution": "720p", "aspect_ratio": "auto", "duration": 10, "image_urls": []any{"image"}, "video_urls": []any{"video"}, "audio": false, "customize_multi_shots": true, "multi_prompt": []map[string]any{{"prompt": "shot", "duration": 5}}, "elements": []map[string]any{{"name": "role", "description": "", "element_input_urls": []string{"image"}}}},
		},
		{
			name: "pure video transformation", model: "kling-3.0-omni/transformation",
			input: map[string]any{"prompt": "scene", "mode": "std", "aspect_ratio": "16:9", "duration": 10, "video_urls": []any{"video"}, "audio": true, "multi_shot": true, "shot_type": "customize", "multi_prompt": []any{}, "elements": []any{map[string]any{"name": "role", "element_input_urls": []any{"image"}}}},
			want:  map[string]any{"prompt": "scene", "resolution": "720p", "aspect_ratio": "auto", "video_urls": []any{"video"}, "audio": true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalizeKIEKlingOmniVideoInput(test.input, test.model)
			if !reflect.DeepEqual(test.input, test.want) {
				t.Fatalf("unexpected input:\nwant: %#v\n got: %#v", test.want, test.input)
			}
		})
	}
}

func protocolJSON(t *testing.T, text string) any {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatalf("invalid fixture JSON: %v", err)
	}
	return value
}

func assertProtocolJSONValue(t *testing.T, got any, want string) {
	t.Helper()
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(protocolJSON(t, string(encoded)), protocolJSON(t, want)) {
		t.Fatalf("got %s; want %s", encoded, want)
	}
}

func blockProtocolNetwork(t *testing.T) {
	t.Helper()
	protocolMockHTTP(t, func(request *http.Request) (*http.Response, error) {
		t.Errorf("unexpected upstream request: %s %s", request.Method, request.URL)
		return nil, errors.New("contract test forbids upstream network I/O")
	})
}

func assertProtocolBytes(t *testing.T, got, want []byte) {
	t.Helper()
	if json.Valid(got) && json.Valid(want) {
		assertProtocolJSONValue(t, protocolJSON(t, string(got)), string(want))
	} else if !bytes.Equal(got, want) {
		t.Fatalf("raw payload changed: got %q, want %q", got, want)
	}
}

type protocolTransport func(*http.Request) (*http.Response, error)

func (transport protocolTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func protocolMockHTTP(t *testing.T, transport protocolTransport) {
	t.Helper()
	previous := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
}

type protocolResponseBody struct {
	io.Reader
	closed *int
}

func (body *protocolResponseBody) Close() error {
	*body.closed++
	return nil
}

func testRecord(t *testing.T, value any) map[string]any {
	t.Helper()
	record, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %#v", value)
	}
	return record
}
