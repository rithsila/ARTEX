package control

import (
	"context"
	"encoding/json"
	"time"
)

// API defines the low-level HTTP interaction required by the control service.
type API interface {
	Do(ctx context.Context, method, path string, body any, out any) error
}

// Health represents the health status of the ARTEX server.
type Health struct {
	OK      bool   `json:"ok"`
	Service string `json:"service,omitempty"`
	Version string `json:"version,omitempty"`
}

// Settings represents the projected, supported ARTEX settings.
// Concurrency fields are read-only.
type Settings struct {
	Workers                 int  `json:"workers"`
	TrafficCapture          bool `json:"traffic_capture"`
	LLMRecord               bool `json:"llm_record"`
	WebSearchEnabled        bool `json:"web_search_enabled"`
	LLMPoolEnabled          bool `json:"llm_pool_enabled"`
	LLMPoolBindFallback     bool `json:"llm_pool_bind_fallback"`
	NOACompaction           bool `json:"noa_compaction"`
	NotifyEnabled           bool `json:"notify_enabled"`
	NotifyDigestIntervalMin int  `json:"notify_digest_interval_min"`
	TaskConcurrencyEnabled  bool `json:"task_concurrency_enabled"`
	TaskConcurrencyLimit    int  `json:"task_concurrency_limit"`
}

// Profile represents a projected LLM configuration profile.
type Profile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Model     string `json:"model"`
	Format    string `json:"format"`
	IsDefault bool   `json:"is_default"`
}

// Agent represents a projected conversational or pipeline agent.
type Agent struct {
	Key          string  `json:"key"`
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	MaxTurns     int     `json:"max_turns"`
	RunSeconds   int     `json:"run_seconds"`
	WebSearch    bool    `json:"web_search"`
	LLMProfileID *string `json:"llm_profile_id"`
}

// AgentDetail contains agent configuration and its current prompt text.
type AgentDetail struct {
	Agent
	Prompt string `json:"prompt"`
}

// Task represents a projected background pentest task.
type Task struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Paused bool   `json:"paused"`
	Queued bool   `json:"queued"`
}

// TaskPage represents a paginated view of tasks with total count.
type TaskPage struct {
	Tasks      []Task `json:"tasks"`
	Total      int    `json:"total"`
	NextOffset *int   `json:"next_offset"`
}

// ChangeRequest specifies a requested mutation before preview.
type ChangeRequest struct {
	Operation string          `json:"operation"` // setting|agent_config|agent_prompt|profile_activate|task_control
	Target    string          `json:"target,omitempty"`    // agent key, profile ID, or task ID; empty for setting
	Key       string          `json:"key,omitempty"`       // field name, or pause|resume for task_control
	Value     json.RawMessage `json:"value,omitempty"`     // scalar or {template,note}; absent for activation/control
}

// ChangePreview represents an immutable proposed change.
type ChangePreview struct {
	ChangeID  string          `json:"change_id"`
	Operation string          `json:"operation"`
	Target    string          `json:"target"`
	Key       string          `json:"key"`
	Before    json.RawMessage `json:"before"`
	After     json.RawMessage `json:"after"`
	Effect    string          `json:"effect"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// ApplyResult represents the outcome of an applied change.
type ApplyResult struct {
	ChangeID        string          `json:"change_id"`
	Outcome         string          `json:"outcome"` // verified|partial_or_failed|outcome_unknown
	Observed        json.RawMessage `json:"observed,omitempty"`
	Effect          string          `json:"effect"`
	AuditIncomplete bool            `json:"audit_incomplete"`
}

// AuditEvent records a mutation attempt and its outcome.
type AuditEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	OperationID   string    `json:"operation_id"`
	SessionID     string    `json:"session_id,omitempty"`
	Operation     string    `json:"operation"`
	Target        string    `json:"target,omitempty"`
	Key           string    `json:"key,omitempty"`
	BeforeHash    string    `json:"before_hash"`
	AfterHash     string    `json:"after_hash"`
	DispatchState string    `json:"dispatch_state"`
	Outcome       string    `json:"outcome"`
	HTTPStatus    int       `json:"http_status"`
}

// Audit is the interface for recording mutation audit events.
type Audit interface {
	Append(ctx context.Context, event AuditEvent) error
}
