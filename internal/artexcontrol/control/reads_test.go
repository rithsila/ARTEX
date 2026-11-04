package control_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Autumn-27/artex/internal/artexcontrol/control"
)

type mockAPI struct {
	doFunc func(ctx context.Context, method, path string, body any, out any) error
	calls  []mockCall
}

type mockCall struct {
	Method string
	Path   string
	Body   any
}

func (m *mockAPI) Do(ctx context.Context, method, path string, body any, out any) error {
	m.calls = append(m.calls, mockCall{Method: method, Path: path, Body: body})
	if m.doFunc != nil {
		return m.doFunc(ctx, method, path, body, out)
	}
	return nil
}

func TestReadRouteMapping(t *testing.T) {
	routes := map[string]string{
		"health":   "/api/health",
		"settings": "/api/settings",
		"profiles": "/api/llm/profiles",
		"agents":   "/api/agents",
		"agent":    "/api/agents/planner",
		"tasks":    "/api/tasks",
		"task":     "/api/tasks/42",
	}

	for name, wantRoute := range routes {
		t.Run(name, func(t *testing.T) {
			api := &mockAPI{
				doFunc: func(ctx context.Context, method, path string, body any, out any) error {
					if method != http.MethodGet {
						t.Errorf("expected GET, got %s", method)
					}
					if path != wantRoute {
						t.Errorf("expected path %s, got %s", wantRoute, path)
					}
					// Return minimal fake payload so unmarshal succeeds
					switch name {
					case "health":
						json.Unmarshal([]byte(`{"ok":true,"service":"artex"}`), out)
					case "settings":
						json.Unmarshal([]byte(`{"workers":3,"traffic_capture":true}`), out)
					case "profiles":
						json.Unmarshal([]byte(`{"profiles":[{"id":"1","name":"Default","model":"gpt-4","format":"openai","is_default":true}]}`), out)
					case "agents":
						json.Unmarshal([]byte(`{"agents":[{"key":"planner","id":"1","name":"Planner","max_turns":10,"run_seconds":60,"web_search":true,"llm_profile_id":null}]}`), out)
					case "agent":
						json.Unmarshal([]byte(`{"agent":{"key":"planner","id":"1","name":"Planner","max_turns":10,"run_seconds":60,"web_search":true,"llm_profile_id":null},"prompt":"You are planner"}`), out)
					case "tasks":
						json.Unmarshal([]byte(`{"tasks":[{"id":"42","name":"Scan","status":"running","paused":false,"queued":false}]}`), out)
					case "task":
						json.Unmarshal([]byte(`{"id":"42","name":"Scan","status":"running","paused":false,"queued":false}`), out)
					}
					return nil
				},
			}

			svc := control.NewService(api)
			ctx := context.Background()

			switch name {
			case "health":
				_, err := svc.Health(ctx)
				if err != nil {
					t.Fatalf("Health failed: %v", err)
				}
			case "settings":
				_, err := svc.Settings(ctx)
				if err != nil {
					t.Fatalf("Settings failed: %v", err)
				}
			case "profiles":
				_, err := svc.Profiles(ctx)
				if err != nil {
					t.Fatalf("Profiles failed: %v", err)
				}
			case "agents":
				_, err := svc.Agents(ctx)
				if err != nil {
					t.Fatalf("Agents failed: %v", err)
				}
			case "agent":
				_, err := svc.Agent(ctx, "planner")
				if err != nil {
					t.Fatalf("Agent failed: %v", err)
				}
			case "tasks":
				_, err := svc.Tasks(ctx, 0, 20)
				if err != nil {
					t.Fatalf("Tasks failed: %v", err)
				}
			case "task":
				_, err := svc.Task(ctx, "42")
				if err != nil {
					t.Fatalf("Task failed: %v", err)
				}
			}
		})
	}
}

func TestReadProjectionDropsSecrets(t *testing.T) {
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			switch path {
			case "/api/llm/profiles":
				return json.Unmarshal([]byte(`{
					"profiles": [{
						"id": "1",
						"name": "Profile1",
						"model": "gpt-4",
						"format": "openai",
						"is_default": true,
						"api_key": "sk-super-secret-12345",
						"api_key_hint": "sk-...12345",
						"proxy": "http://user:pass@proxy.internal:8080",
						"unknown_field": "should_be_dropped",
						"bearer_token": "token-123"
					}]
				}`), out)
			case "/api/settings":
				return json.Unmarshal([]byte(`{
					"workers": 4,
					"traffic_capture": true,
					"llm_record": false,
					"web_search_enabled": true,
					"llm_pool_enabled": false,
					"llm_pool_bind_fallback": false,
					"noa_compaction": false,
					"notify_enabled": true,
					"notify_digest_interval_min": 60,
					"task_concurrency_enabled": true,
					"task_concurrency_limit": 5,
					"brave_key": "brave-secret-key",
					"tavily_key": "tavily-secret-key",
					"web_search_proxy": "http://admin:secret@proxy:8080",
					"global_proxy": "http://admin:secret@proxy:8080",
					"python_interpreter": "/usr/bin/python3"
				}`), out)
			}
			return nil
		},
	}

	svc := control.NewService(api)
	ctx := context.Background()

	// 1. Check Profiles projection
	profiles, err := svc.Profiles(ctx)
	if err != nil {
		t.Fatalf("Profiles: %v", err)
	}
	pBytes, _ := json.Marshal(profiles)
	pStr := string(pBytes)
	for _, forbidden := range []string{"api_key", "api_key_hint", "sk-super-secret", "proxy.internal", "unknown_field", "bearer_token"} {
		if strings.Contains(pStr, forbidden) {
			t.Errorf("Profiles projection leaked forbidden field %q: %s", forbidden, pStr)
		}
	}

	// 2. Check Settings projection
	settings, err := svc.Settings(ctx)
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	sBytes, _ := json.Marshal(settings)
	sStr := string(sBytes)
	for _, forbidden := range []string{"brave_key", "tavily_key", "web_search_proxy", "global_proxy", "python_interpreter", "admin:secret"} {
		if strings.Contains(sStr, forbidden) {
			t.Errorf("Settings projection leaked forbidden field %q: %s", forbidden, sStr)
		}
	}
}

func TestLargeIDsPreserved(t *testing.T) {
	largeID := "9007199254740993" // 2^53 + 1, exceeds JS Number.MAX_SAFE_INTEGER
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			switch path {
			case "/api/llm/profiles":
				return json.Unmarshal([]byte(fmt.Sprintf(`{"profiles":[{"id":"%s","name":"Large","model":"m","format":"f","is_default":false}]}`, largeID)), out)
			case "/api/agents":
				// Backend sends numeric llm_profile_id
				return json.Unmarshal([]byte(fmt.Sprintf(`{"agents":[{"key":"worker","id":"%s","name":"Worker","max_turns":5,"run_seconds":30,"web_search":false,"llm_profile_id":%s}]}`, largeID, largeID)), out)
			case "/api/agents/worker":
				return json.Unmarshal([]byte(fmt.Sprintf(`{"agent":{"key":"worker","id":"%s","name":"Worker","max_turns":5,"run_seconds":30,"web_search":false,"llm_profile_id":%s},"prompt":"p"}`, largeID, largeID)), out)
			case "/api/tasks":
				return json.Unmarshal([]byte(fmt.Sprintf(`{"tasks":[{"id":"%s","name":"T","status":"queued","paused":false,"queued":true}]}`, largeID)), out)
			case "/api/tasks/" + largeID:
				return json.Unmarshal([]byte(fmt.Sprintf(`{"id":"%s","name":"T","status":"queued","paused":false,"queued":true}`, largeID)), out)
			}
			return nil
		},
	}

	svc := control.NewService(api)
	ctx := context.Background()

	// Profile ID check
	profs, err := svc.Profiles(ctx)
	if err != nil || len(profs) == 0 {
		t.Fatalf("Profiles: %v", err)
	}
	if profs[0].ID != largeID {
		t.Errorf("expected Profile ID %s, got %s", largeID, profs[0].ID)
	}

	// Agent llm_profile_id check
	agents, err := svc.Agents(ctx)
	if err != nil || len(agents) == 0 {
		t.Fatalf("Agents: %v", err)
	}
	if agents[0].LLMProfileID == nil || *agents[0].LLMProfileID != largeID {
		t.Errorf("expected Agent LLMProfileID %s, got %v", largeID, agents[0].LLMProfileID)
	}

	// Task ID check
	task, err := svc.Task(ctx, largeID)
	if err != nil {
		t.Fatalf("Task: %v", err)
	}
	if task.ID != largeID {
		t.Errorf("expected Task ID %s, got %s", largeID, task.ID)
	}
}

func TestTaskPage(t *testing.T) {
	// Raw list from backend unsorted: IDs "10", "2", "1", "20"
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if path != "/api/tasks" {
				return fmt.Errorf("unexpected path %s", path)
			}
			return json.Unmarshal([]byte(`{
				"tasks": [
					{"id":"10","name":"Task 10","status":"done","paused":false,"queued":false},
					{"id":"2","name":"Task 2","status":"running","paused":false,"queued":false},
					{"id":"1","name":"Task 1","status":"done","paused":false,"queued":false},
					{"id":"20","name":"Task 20","status":"queued","paused":false,"queued":true}
				]
			}`), out)
		},
	}

	svc := control.NewService(api)
	ctx := context.Background()

	// 1. Page with offset 0, limit 2
	page1, err := svc.Tasks(ctx, 0, 2)
	if err != nil {
		t.Fatalf("Tasks(0, 2): %v", err)
	}
	if page1.Total != 4 {
		t.Errorf("expected Total 4, got %d", page1.Total)
	}
	if len(page1.Tasks) != 2 {
		t.Fatalf("expected 2 tasks on page 1, got %d", len(page1.Tasks))
	}
	// Check numeric ordering: ID "1", then "2"
	if page1.Tasks[0].ID != "1" || page1.Tasks[1].ID != "2" {
		t.Errorf("tasks not sorted numerically: got %s, %s", page1.Tasks[0].ID, page1.Tasks[1].ID)
	}
	if page1.NextOffset == nil || *page1.NextOffset != 2 {
		t.Errorf("expected NextOffset 2, got %v", page1.NextOffset)
	}

	// 2. Page with offset 2, limit 2
	page2, err := svc.Tasks(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Tasks(2, 2): %v", err)
	}
	if len(page2.Tasks) != 2 {
		t.Fatalf("expected 2 tasks on page 2, got %d", len(page2.Tasks))
	}
	if page2.Tasks[0].ID != "10" || page2.Tasks[1].ID != "20" {
		t.Errorf("expected IDs 10, 20 on page 2, got %s, %s", page2.Tasks[0].ID, page2.Tasks[1].ID)
	}
	if page2.NextOffset != nil {
		t.Errorf("expected NextOffset nil at end of list, got %v", page2.NextOffset)
	}

	// 3. Offset beyond total
	pageEmpty, err := svc.Tasks(ctx, 10, 20)
	if err != nil {
		t.Fatalf("Tasks(10, 20): %v", err)
	}
	if len(pageEmpty.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(pageEmpty.Tasks))
	}
	if pageEmpty.Total != 4 {
		t.Errorf("expected Total 4, got %d", pageEmpty.Total)
	}
	if pageEmpty.NextOffset != nil {
		t.Errorf("expected NextOffset nil, got %v", pageEmpty.NextOffset)
	}
}
