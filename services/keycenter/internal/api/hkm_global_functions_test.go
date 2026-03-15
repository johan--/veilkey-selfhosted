package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGlobalFunctionsCRUD(t *testing.T) {
	_, handler := setupHKMServer(t)

	save := postJSON(handler, "/api/functions/global", map[string]any{
		"name":          "gitlab-project-get",
		"function_hash": "fnhash001",
		"category":      "gitlab",
		"command":       `curl -H "PRIVATE-TOKEN: {%{GITLAB_TOKEN}%}" https://gitlab.example/api/v4/projects/{%{PROJECT_ID}%}`,
		"vars_json":     `{"GITLAB_TOKEN":{"ref":"VK:EXTERNAL:abcd1234","class":"EXTERNAL"},"PROJECT_ID":{"ref":"VE:LOCAL:project_id","class":"LOCAL"}}`,
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save expected 200, got %d: %s", save.Code, save.Body.String())
	}

	list := getJSON(handler, "/api/functions/global")
	if list.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d: %s", list.Code, list.Body.String())
	}
	if body := list.Body.String(); !containsAll(body, `"count":1`, `"name":"gitlab-project-get"`, `"category":"gitlab"`) {
		t.Fatalf("unexpected list body: %s", body)
	}

	get := getJSON(handler, "/api/functions/global/gitlab-project-get")
	if get.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d: %s", get.Code, get.Body.String())
	}
	if body := get.Body.String(); !containsAll(body, `"function_hash":"fnhash001"`, `"vars_json":"{`) {
		t.Fatalf("unexpected get body: %s", body)
	}

	del := deleteJSON(handler, "/api/functions/global/gitlab-project-get")
	if del.Code != http.StatusOK {
		t.Fatalf("delete expected 200, got %d: %s", del.Code, del.Body.String())
	}

	getMissing := getJSON(handler, "/api/functions/global/gitlab-project-get")
	if getMissing.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d: %s", getMissing.Code, getMissing.Body.String())
	}
}

func TestGlobalFunctionRun(t *testing.T) {
	_, handler := setupHKMServer(t)
	resolveAndRun := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/resolve/VE:LOCAL:project_id"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"value":"project-123"}`))
		case r.URL.Path == "/run/project-123":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer resolveAndRun.Close()
	t.Setenv("VEILKEY_ADDR", resolveAndRun.URL)

	save := postJSONFromIP(handler, "/api/functions/global", "127.0.0.1:12345", map[string]any{
		"name":          "echo-function",
		"function_hash": "fnhash-run-001",
		"category":      "test",
		"command":       "curl -sS " + resolveAndRun.URL + "/run/{%{PROJECT_ID}%}",
		"vars_json":     `{"PROJECT_ID":{"ref":"VE:LOCAL:project_id","class":"LOCAL"}}`,
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save expected 200, got %d: %s", save.Code, save.Body.String())
	}

	run := postJSONFromIP(handler, "/api/functions/global/echo-function/run", "127.0.0.1:12345", map[string]any{
		"prompt": "ignored",
	})
	if run.Code != http.StatusOK {
		t.Fatalf("run expected 200, got %d: %s", run.Code, run.Body.String())
	}

	var payload struct {
		Name       string `json:"name"`
		Command    string `json:"command"`
		Rendered   string `json:"rendered"`
		Stdout     string `json:"stdout"`
		Successful bool   `json:"successful"`
		ExitCode   int    `json:"exit_code"`
	}
	if err := json.Unmarshal(run.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if payload.Name != "echo-function" {
		t.Fatalf("name = %q", payload.Name)
	}
	if payload.Command == "" || payload.Rendered == "" {
		t.Fatalf("missing command/rendered: %+v", payload)
	}
	if !payload.Successful {
		t.Fatalf("expected successful run: %+v", payload)
	}
	if payload.ExitCode != 0 {
		t.Fatalf("expected zero exit code")
	}
	if payload.Stdout != "ok" {
		t.Fatalf("stdout = %q, want ok", payload.Stdout)
	}
}

func TestGlobalFunctionRunWithEnvOverrides(t *testing.T) {
	_, handler := setupHKMServer(t)
	runServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/run" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("X-System"); got != "frontend system prompt" {
			t.Fatalf("X-System = %q", got)
		}
		if got := r.Header.Get("X-Temperature"); got != "0.4" {
			t.Fatalf("X-Temperature = %q", got)
		}
		if got := r.Header.Get("X-Max-Tokens"); got != "2048" {
			t.Fatalf("X-Max-Tokens = %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer runServer.Close()

	save := postJSONFromIP(handler, "/api/functions/global", "127.0.0.1:12345", map[string]any{
		"name":          "env-function",
		"function_hash": "fnhash-run-002",
		"category":      "test",
		"command":       `curl -sS -H "X-System: ${VEILKEY_GEMINI_FRONTEND_SYSTEM:-}" -H "X-Temperature: ${VEILKEY_GEMINI_TEMPERATURE:-}" -H "X-Max-Tokens: ${VEILKEY_GEMINI_MAX_OUTPUT_TOKENS:-}" ` + runServer.URL + `/run`,
		"vars_json":     `{}`,
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save expected 200, got %d: %s", save.Code, save.Body.String())
	}

	run := postJSONFromIP(handler, "/api/functions/global/env-function/run", "127.0.0.1:12345", map[string]any{
		"prompt":            "ignored",
		"system_prompt":     "frontend system prompt",
		"temperature":       0.4,
		"max_output_tokens": 2048,
	})
	if run.Code != http.StatusOK {
		t.Fatalf("run expected 200, got %d: %s", run.Code, run.Body.String())
	}

	var payload struct {
		Stdout     string   `json:"stdout"`
		EnvKeys    []string `json:"env_keys"`
		InputBytes int      `json:"input_bytes"`
	}
	if err := json.Unmarshal(run.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if payload.Stdout != "ok" {
		t.Fatalf("stdout = %q, want ok", payload.Stdout)
	}
	if len(payload.EnvKeys) != 3 {
		t.Fatalf("env keys = %#v", payload.EnvKeys)
	}
	if payload.InputBytes != len("ignored") {
		t.Fatalf("input bytes = %d", payload.InputBytes)
	}
}

func TestGlobalFunctionRunTimeout(t *testing.T) {
	_, handler := setupHKMServer(t)
	runServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("late"))
	}))
	defer runServer.Close()

	save := postJSONFromIP(handler, "/api/functions/global", "127.0.0.1:12345", map[string]any{
		"name":          "timeout-function",
		"function_hash": "fnhash-run-003",
		"category":      "test",
		"command":       `curl -sS ` + runServer.URL,
		"vars_json":     `{}`,
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save expected 200, got %d: %s", save.Code, save.Body.String())
	}

	run := postJSONFromIP(handler, "/api/functions/global/timeout-function/run", "127.0.0.1:12345", map[string]any{
		"timeout_seconds": 1,
	})
	if run.Code != http.StatusBadGateway {
		t.Fatalf("run expected 502, got %d: %s", run.Code, run.Body.String())
	}

	var payload struct {
		Successful bool `json:"successful"`
		TimedOut   bool `json:"timed_out"`
	}
	if err := json.Unmarshal(run.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if payload.Successful {
		t.Fatalf("expected failed run")
	}
	if !payload.TimedOut {
		t.Fatalf("expected timeout payload: %s", run.Body.String())
	}
}

func containsAll(body string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(body, part) {
			return false
		}
	}
	return true
}
