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

func TestValidationRejectsUnknownOrWrongTypes(t *testing.T) {
	cases := []struct {
		name    string
		req     control.ChangeRequest
		wantErr string
	}{
		{
			name: "unsupported operation",
			req: control.ChangeRequest{
				Operation: "delete_database",
			},
			wantErr: "unsupported operation",
		},
		{
			name: "setting with non-empty target",
			req: control.ChangeRequest{
				Operation: "setting",
				Target:    "planner",
				Key:       "workers",
				Value:     json.RawMessage(`4`),
			},
			wantErr: "setting target must be empty",
		},
		{
			name: "setting forbidden secret key",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "brave_key",
				Value:     json.RawMessage(`"some-key"`),
			},
			wantErr: "unsupported or read-only setting key",
		},
		{
			name: "setting proxy key forbidden",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "global_proxy",
				Value:     json.RawMessage(`"http://127.0.0.1:8080"`),
			},
			wantErr: "unsupported or read-only setting key",
		},
		{
			name: "setting concurrency limit read-only",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "task_concurrency_limit",
				Value:     json.RawMessage(`5`),
			},
			wantErr: "concurrency settings are read-only",
		},
		{
			name: "setting concurrency switch read-only",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "task_concurrency_enabled",
				Value:     json.RawMessage(`true`),
			},
			wantErr: "concurrency settings are read-only",
		},
		{
			name: "setting boolean given string false",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "traffic_capture",
				Value:     json.RawMessage(`"false"`),
			},
			wantErr: "must be a boolean",
		},
		{
			name: "setting workers non-positive",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "workers",
				Value:     json.RawMessage(`0`),
			},
			wantErr: "workers must be a positive integer",
		},
		{
			name: "setting notify digest interval out of range",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "notify_digest_interval_min",
				Value:     json.RawMessage(`1441`),
			},
			wantErr: "notify_digest_interval_min must be between 1 and 1440",
		},
		{
			name: "agent config invalid key name",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "invalid_key",
				Value:     json.RawMessage(`10`),
			},
			wantErr: "unsupported agent_config key",
		},
		{
			name: "agent config invalid agent key syntax",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "123-bad-agent",
				Key:       "max_turns",
				Value:     json.RawMessage(`10`),
			},
			wantErr: "invalid agent key format",
		},
		{
			name: "agent config max_turns null",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "max_turns",
				Value:     json.RawMessage(`null`),
			},
			wantErr: "max_turns cannot be null",
		},
		{
			name: "agent config max_turns decimal",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "max_turns",
				Value:     json.RawMessage(`10.5`),
			},
			wantErr: "max_turns must be an integer",
		},
		{
			name: "agent config max_turns negative",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "max_turns",
				Value:     json.RawMessage(`-1`),
			},
			wantErr: "max_turns must be non-negative",
		},
		{
			name: "agent config profile_id negative",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "llm_profile_id",
				Value:     json.RawMessage(`"-5"`),
			},
			wantErr: "profile ID must be a positive decimal string",
		},
		{
			name: "agent config profile_id exponent",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "llm_profile_id",
				Value:     json.RawMessage(`"1e5"`),
			},
			wantErr: "profile ID must be a positive decimal string",
		},
		{
			name: "agent config profile_id overflow",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "llm_profile_id",
				Value:     json.RawMessage(`"999999999999999999999999999999"`),
			},
			wantErr: "profile ID exceeds 64-bit range",
		},
		{
			name: "agent prompt empty template",
			req: control.ChangeRequest{
				Operation: "agent_prompt",
				Target:    "planner",
				Value:     json.RawMessage(`{"template":"","note":"update"}`),
			},
			wantErr: "template cannot be empty",
		},
		{
			name: "agent prompt oversized template",
			req: control.ChangeRequest{
				Operation: "agent_prompt",
				Target:    "planner",
				Value:     json.RawMessage(fmt.Sprintf(`{"template":"%s","note":"update"}`, strings.Repeat("a", 65537))),
			},
			wantErr: "template exceeds 65536 bytes",
		},
		{
			name: "agent prompt note too long",
			req: control.ChangeRequest{
				Operation: "agent_prompt",
				Target:    "planner",
				Value:     json.RawMessage(fmt.Sprintf(`{"template":"valid","note":"%s"}`, strings.Repeat("n", 257))),
			},
			wantErr: "note exceeds 256 bytes",
		},
		{
			name: "agent prompt unknown field in payload",
			req: control.ChangeRequest{
				Operation: "agent_prompt",
				Target:    "planner",
				Value:     json.RawMessage(`{"template":"valid","note":"update","extra":"forbidden"}`),
			},
			wantErr: "unknown field in prompt payload",
		},
		{
			name: "profile_activate invalid ID",
			req: control.ChangeRequest{
				Operation: "profile_activate",
				Target:    "not-a-number",
			},
			wantErr: "profile ID must be a positive decimal string",
		},
		{
			name: "task_control invalid action",
			req: control.ChangeRequest{
				Operation: "task_control",
				Target:    "123",
				Key:       "stop",
			},
			wantErr: "task_control key must be pause or resume",
		},
		{
			name: "task_control invalid task ID",
			req: control.ChangeRequest{
				Operation: "task_control",
				Target:    "-1",
				Key:       "pause",
			},
			wantErr: "task ID must be a positive decimal string",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := control.ValidateChange(tc.req)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestConcurrencyReadOnly(t *testing.T) {
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method != http.MethodGet {
				t.Fatalf("concurrency test attempted mutation: %s %s", method, path)
			}
			return json.Unmarshal([]byte(`{
				"task_concurrency_enabled": false,
				"task_concurrency_limit": 5
			}`), out)
		},
	}

	svc := control.NewService(api)
	settings, err := svc.Settings(context.Background())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if settings.TaskConcurrencyEnabled != false || settings.TaskConcurrencyLimit != 5 {
		t.Errorf("unexpected settings values: %+v", settings)
	}

	// ValidateChange rejects attempts to write concurrency fields
	for _, key := range []string{"task_concurrency_enabled", "task_concurrency_limit"} {
		err := control.ValidateChange(control.ChangeRequest{
			Operation: "setting",
			Key:       key,
			Value:     json.RawMessage(`true`),
		})
		if err == nil {
			t.Errorf("expected ValidateChange to reject writing %s", key)
		}
	}
}

func TestProfileBindingThreeStates(t *testing.T) {
	// Three states:
	// 1. null -> unbind
	// 2. "9007199254740993" -> large numeric ID
	// 3. invalid values (negative, exponent, string text) -> rejected
	reqNull := control.ChangeRequest{
		Operation: "agent_config",
		Target:    "planner",
		Key:       "llm_profile_id",
		Value:     json.RawMessage(`null`),
	}
	if err := control.ValidateChange(reqNull); err != nil {
		t.Errorf("expected null profile unbind to be valid, got: %v", err)
	}

	reqLarge := control.ChangeRequest{
		Operation: "agent_config",
		Target:    "planner",
		Key:       "llm_profile_id",
		Value:     json.RawMessage(`"9007199254740993"`),
	}
	if err := control.ValidateChange(reqLarge); err != nil {
		t.Errorf("expected large numeric profile binding to be valid, got: %v", err)
	}
}

func TestOutputLimit(t *testing.T) {
	// If the server returns a prompt or list that marshals to > 256 KiB, control.Service must return response_too_large error
	hugePrompt := strings.Repeat("A", 300*1024)
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			return json.Unmarshal([]byte(fmt.Sprintf(`{"agent":{"key":"planner","id":"1","name":"Planner"},"prompt":"%s"}`, hugePrompt)), out)
		},
	}

	svc := control.NewService(api)
	_, err := svc.Agent(context.Background(), "planner")
	if err == nil {
		t.Fatalf("expected error for oversize output, got nil")
	}
	if !strings.Contains(err.Error(), "response_too_large") && !strings.Contains(err.Error(), "output exceeded 256 KiB limit") {
		t.Errorf("expected response_too_large error, got: %v", err)
	}
}

func TestReadsDoNotMutate(t *testing.T) {
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method != http.MethodGet {
				t.Fatalf("expected only GET requests, got: %s %s", method, path)
			}
			switch path {
			case "/api/health":
				return json.Unmarshal([]byte(`{"ok":true}`), out)
			case "/api/settings":
				return json.Unmarshal([]byte(`{"workers":3}`), out)
			case "/api/llm/profiles":
				return json.Unmarshal([]byte(`{"profiles":[]}`), out)
			case "/api/agents":
				return json.Unmarshal([]byte(`{"agents":[]}`), out)
			case "/api/agents/planner":
				return json.Unmarshal([]byte(`{"agent":{"key":"planner","id":"1","name":"P"},"prompt":""}`), out)
			case "/api/tasks":
				return json.Unmarshal([]byte(`{"tasks":[]}`), out)
			case "/api/tasks/1":
				return json.Unmarshal([]byte(`{"id":"1","name":"T","status":"done"}`), out)
			}
			return nil
		},
	}

	svc := control.NewService(api)
	ctx := context.Background()

	_, _ = svc.Health(ctx)
	_, _ = svc.Settings(ctx)
	_, _ = svc.Profiles(ctx)
	_, _ = svc.Agents(ctx)
	_, _ = svc.Agent(ctx, "planner")
	_, _ = svc.Tasks(ctx, 0, 20)
	_, _ = svc.Task(ctx, "1")

	for _, c := range api.calls {
		if c.Method != http.MethodGet {
			t.Errorf("expected call method GET, got %s for path %s", c.Method, c.Path)
		}
		if c.Body != nil {
			t.Errorf("expected nil body for read call to %s", c.Path)
		}
	}
}
