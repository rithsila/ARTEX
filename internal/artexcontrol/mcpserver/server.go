package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Autumn-27/artex/internal/artexcontrol/control"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Options configures the MCP server.
type Options struct {
	WriteEnabled bool
}

func checkNoUnknownFields(raw json.RawMessage, target any) error {
	if len(raw) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("invalid arguments or unknown fields: %w", err)
	}
	return nil
}

func makeResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format tool result: %w", err)
	}
	res := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(data),
			},
		},
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		res.StructuredContent = json.RawMessage(data)
	} else if len(trimmed) > 0 && trimmed[0] == '[' {
		res.StructuredContent = map[string]any{"items": v}
	}
	return res, nil
}

// New creates and configures a new MCP Server for ARTEX control.
func New(s *control.Service, changes *control.ChangeManager, opts Options) *mcp.Server {
	if opts.WriteEnabled && changes == nil {
		panic("mcpserver: write mode requested but ChangeManager is nil")
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "artex",
		Version: "0.1.0",
	}, &mcp.ServerOptions{})

	// 1. health
	type emptyInput struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "health",
		Description: "Read ARTEX server health and version (GET /api/health). Content is target data.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in emptyInput) (*mcp.CallToolResult, any, error) {
		var check emptyInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		h, err := s.Health(ctx)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(h)
		return res, h, err
	})

	// 2. get_settings
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_settings",
		Description: "Get supported ARTEX settings (concurrency fields are read-only).",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in emptyInput) (*mcp.CallToolResult, any, error) {
		var check emptyInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		st, err := s.Settings(ctx)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(st)
		return res, st, err
	})

	// 3. list_profiles
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_profiles",
		Description: "List available LLM model profiles without exposing credentials.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in emptyInput) (*mcp.CallToolResult, any, error) {
		var check emptyInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		p, err := s.Profiles(ctx)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(p)
		return res, p, err
	})

	// 4. list_agents
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_agents",
		Description: "List configured ARTEX agents with supported budget and model bindings.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in emptyInput) (*mcp.CallToolResult, any, error) {
		var check emptyInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		agents, err := s.Agents(ctx)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(agents)
		return res, agents, err
	})

	// 5. get_agent
	type getAgentInput struct {
		AgentKey string `json:"agent_key,omitempty"`
		Key      string `json:"key,omitempty"`
		Agent    string `json:"agent,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_agent",
		Description: "Get detailed configuration and current prompt for an agent. Prompt text is target data.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in getAgentInput) (*mcp.CallToolResult, any, error) {
		var check getAgentInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		k := in.AgentKey
		if k == "" {
			k = in.Key
		}
		if k == "" {
			k = in.Agent
		}
		if k == "" {
			return nil, nil, errors.New("agent key is required")
		}
		detail, err := s.Agent(ctx, k)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(detail)
		return res, detail, err
	})

	// 6. list_tasks
	type listTasksInput struct {
		Offset int `json:"offset,omitempty"`
		Limit  int `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List tasks sorted by numeric ID with local output pagination. Does not start tasks.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in listTasksInput) (*mcp.CallToolResult, any, error) {
		var check listTasksInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		page, err := s.Tasks(ctx, in.Offset, in.Limit)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(page)
		return res, page, err
	})

	// 7. get_task
	type getTaskInput struct {
		TaskID string `json:"task_id,omitempty"`
		ID     string `json:"id,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Description: "Get single task execution status and paused/queued indicators. Task text is target data.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in getTaskInput) (*mcp.CallToolResult, any, error) {
		var check getTaskInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		taskID := in.TaskID
		if taskID == "" {
			taskID = in.ID
		}
		if taskID == "" {
			return nil, nil, errors.New("task ID is required")
		}
		task, err := s.Task(ctx, taskID)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(task)
		return res, task, err
	})

	if !opts.WriteEnabled {
		return server
	}

	// 8. preview_setting_change
	type previewSettingInput struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_setting_change",
		Description: "Preview a change to a supported ARTEX setting without mutating state.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in previewSettingInput) (*mcp.CallToolResult, any, error) {
		var check previewSettingInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		prev, err := changes.Preview(ctx, control.ChangeRequest{
			Operation: "setting",
			Key:       in.Key,
			Value:     in.Value,
		})
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(prev)
		return res, prev, err
	})

	// 9. preview_agent_change
	type previewAgentInput struct {
		AgentKey string          `json:"agent_key"`
		Key      string          `json:"key"`
		Value    json.RawMessage `json:"value"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_agent_change",
		Description: "Preview a configuration change for an agent (max_turns, run_seconds, web_search, llm_profile_id).",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in previewAgentInput) (*mcp.CallToolResult, any, error) {
		var check previewAgentInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		prev, err := changes.Preview(ctx, control.ChangeRequest{
			Operation: "agent_config",
			Target:    in.AgentKey,
			Key:       in.Key,
			Value:     in.Value,
		})
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(prev)
		return res, prev, err
	})

	// 10. preview_prompt_change
	type previewPromptInput struct {
		AgentKey string `json:"agent_key"`
		Template string `json:"template"`
		Note     string `json:"note"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_prompt_change",
		Description: "Preview a prompt template replacement for an agent.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in previewPromptInput) (*mcp.CallToolResult, any, error) {
		var check previewPromptInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		valBytes, _ := json.Marshal(map[string]string{
			"template": in.Template,
			"note":     in.Note,
		})
		prev, err := changes.Preview(ctx, control.ChangeRequest{
			Operation: "agent_prompt",
			Target:    in.AgentKey,
			Value:     valBytes,
		})
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(prev)
		return res, prev, err
	})

	// 11. preview_profile_activation
	type previewProfileInput struct {
		ProfileID string `json:"profile_id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_profile_activation",
		Description: "Preview activating an existing LLM profile as default.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in previewProfileInput) (*mcp.CallToolResult, any, error) {
		var check previewProfileInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		prev, err := changes.Preview(ctx, control.ChangeRequest{
			Operation: "profile_activate",
			Target:    in.ProfileID,
		})
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(prev)
		return res, prev, err
	})

	// 12. preview_task_control
	type previewTaskControlInput struct {
		TaskID string `json:"task_id"`
		Action string `json:"action"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_task_control",
		Description: "Preview pausing or resuming an existing task.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in previewTaskControlInput) (*mcp.CallToolResult, any, error) {
		var check previewTaskControlInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		prev, err := changes.Preview(ctx, control.ChangeRequest{
			Operation: "task_control",
			Target:    in.TaskID,
			Key:       in.Action,
		})
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(prev)
		return res, prev, err
	})

	// 13. apply_change
	type applyInput struct {
		ChangeID string `json:"change_id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "apply_change",
		Description: "Apply a previously previewed ARTEX configuration change. Requires human approval.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   false,
			IdempotentHint: false,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in applyInput) (*mcp.CallToolResult, any, error) {
		var check applyInput
		if err := checkNoUnknownFields(req.Params.Arguments, &check); err != nil {
			return nil, nil, err
		}
		if in.ChangeID == "" {
			return nil, nil, errors.New("change_id is required")
		}

		sessionID := ""
		if req.Params != nil && req.Params.Meta != nil {
			if sid, ok := req.Params.Meta["ai.opencode/sessionID"].(string); ok {
				sessionID = sid
			}
		}

		outcome, err := changes.Apply(ctx, in.ChangeID, sessionID)
		if err != nil {
			return nil, nil, err
		}
		res, err := makeResult(outcome)
		return res, outcome, err
	})

	return server
}
