package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
)

const maxOutputBytes = 256 * 1024 // 256 KiB

// Service provides typed, bounded read access to the ARTEX API.
type Service struct {
	api API
}

// NewService creates a new control Service.
func NewService(api API) *Service {
	return &Service{api: api}
}

func checkOutputSize(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to check output size: %w", err)
	}
	if len(data) > maxOutputBytes {
		return errors.New("response_too_large: output exceeded 256 KiB limit")
	}
	return nil
}

// Health checks ARTEX server health.
func (s *Service) Health(ctx context.Context) (Health, error) {
	var resp struct {
		OK      bool   `json:"ok"`
		Service string `json:"service"`
		Version string `json:"version"`
	}
	if err := s.api.Do(ctx, http.MethodGet, "/api/health", nil, &resp); err != nil {
		return Health{}, err
	}
	h := Health{
		OK:      resp.OK,
		Service: resp.Service,
		Version: resp.Version,
	}
	if err := checkOutputSize(h); err != nil {
		return Health{}, err
	}
	return h, nil
}

// Settings returns the projected, supported ARTEX settings.
func (s *Service) Settings(ctx context.Context) (Settings, error) {
	var raw struct {
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
	if err := s.api.Do(ctx, http.MethodGet, "/api/settings", nil, &raw); err != nil {
		return Settings{}, err
	}

	st := Settings{
		Workers:                 raw.Workers,
		TrafficCapture:          raw.TrafficCapture,
		LLMRecord:               raw.LLMRecord,
		WebSearchEnabled:        raw.WebSearchEnabled,
		LLMPoolEnabled:          raw.LLMPoolEnabled,
		LLMPoolBindFallback:     raw.LLMPoolBindFallback,
		NOACompaction:           raw.NOACompaction,
		NotifyEnabled:           raw.NotifyEnabled,
		NotifyDigestIntervalMin: raw.NotifyDigestIntervalMin,
		TaskConcurrencyEnabled:  raw.TaskConcurrencyEnabled,
		TaskConcurrencyLimit:    raw.TaskConcurrencyLimit,
	}
	if err := checkOutputSize(st); err != nil {
		return Settings{}, err
	}
	return st, nil
}

// Profiles lists LLM configuration profiles.
func (s *Service) Profiles(ctx context.Context) ([]Profile, error) {
	var raw struct {
		Profiles []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Model     string `json:"model"`
			Format    string `json:"format"`
			IsDefault bool   `json:"is_default"`
		} `json:"profiles"`
	}
	if err := s.api.Do(ctx, http.MethodGet, "/api/llm/profiles", nil, &raw); err != nil {
		return nil, err
	}

	out := make([]Profile, 0, len(raw.Profiles))
	for _, p := range raw.Profiles {
		out = append(out, Profile{
			ID:        p.ID,
			Name:      p.Name,
			Model:     p.Model,
			Format:    p.Format,
			IsDefault: p.IsDefault,
		})
	}
	if err := checkOutputSize(out); err != nil {
		return nil, err
	}
	return out, nil
}

type rawAgentWire struct {
	Key          string `json:"key"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	MaxTurns     int    `json:"max_turns"`
	RunSeconds   int    `json:"run_seconds"`
	WebSearch    bool   `json:"web_search"`
	LLMProfileID *int64 `json:"llm_profile_id"`
}

func projectAgent(raw rawAgentWire) Agent {
	var profID *string
	if raw.LLMProfileID != nil {
		s := strconv.FormatInt(*raw.LLMProfileID, 10)
		profID = &s
	}
	return Agent{
		Key:          raw.Key,
		ID:           raw.ID,
		Name:         raw.Name,
		MaxTurns:     raw.MaxTurns,
		RunSeconds:   raw.RunSeconds,
		WebSearch:    raw.WebSearch,
		LLMProfileID: profID,
	}
}

// Agents lists all configured agents.
func (s *Service) Agents(ctx context.Context) ([]Agent, error) {
	var raw struct {
		Agents []rawAgentWire `json:"agents"`
	}
	if err := s.api.Do(ctx, http.MethodGet, "/api/agents", nil, &raw); err != nil {
		return nil, err
	}

	out := make([]Agent, 0, len(raw.Agents))
	for _, a := range raw.Agents {
		out = append(out, projectAgent(a))
	}
	if err := checkOutputSize(out); err != nil {
		return nil, err
	}
	return out, nil
}

// Agent gets detailed information and current prompt for a specific agent.
func (s *Service) Agent(ctx context.Context, key string) (AgentDetail, error) {
	if err := ValidateAgentKey(key); err != nil {
		return AgentDetail{}, err
	}

	var raw struct {
		Agent  rawAgentWire `json:"agent"`
		Prompt string       `json:"prompt"`
	}
	if err := s.api.Do(ctx, http.MethodGet, "/api/agents/"+key, nil, &raw); err != nil {
		return AgentDetail{}, err
	}

	detail := AgentDetail{
		Agent:  projectAgent(raw.Agent),
		Prompt: raw.Prompt,
	}
	if err := checkOutputSize(detail); err != nil {
		return AgentDetail{}, err
	}
	return detail, nil
}

// Tasks returns a paginated and numerically sorted slice of background tasks.
func (s *Service) Tasks(ctx context.Context, offset, limit int) (TaskPage, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var raw struct {
		Tasks []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
			Paused bool   `json:"paused"`
			Queued bool   `json:"queued"`
		} `json:"tasks"`
	}
	if err := s.api.Do(ctx, http.MethodGet, "/api/tasks", nil, &raw); err != nil {
		return TaskPage{}, err
	}

	all := make([]Task, 0, len(raw.Tasks))
	for _, t := range raw.Tasks {
		all = append(all, Task{
			ID:     t.ID,
			Name:   t.Name,
			Status: t.Status,
			Paused: t.Paused,
			Queued: t.Queued,
		})
	}

	// Sort numerically by ID
	sort.Slice(all, func(i, j int) bool {
		idI, errI := strconv.ParseInt(all[i].ID, 10, 64)
		idJ, errJ := strconv.ParseInt(all[j].ID, 10, 64)
		if errI == nil && errJ == nil {
			return idI < idJ
		}
		return all[i].ID < all[j].ID
	})

	total := len(all)
	var paged []Task
	var nextOffset *int

	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paged = all[offset:end]
		if end < total {
			no := end
			nextOffset = &no
		}
	} else {
		paged = []Task{}
	}

	page := TaskPage{
		Tasks:      paged,
		Total:      total,
		NextOffset: nextOffset,
	}

	if err := checkOutputSize(page); err != nil {
		return TaskPage{}, err
	}
	return page, nil
}

// Task retrieves single task details.
func (s *Service) Task(ctx context.Context, id string) (Task, error) {
	if err := ValidateID(id); err != nil {
		return Task{}, err
	}

	var raw struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
		Paused bool   `json:"paused"`
		Queued bool   `json:"queued"`
	}
	if err := s.api.Do(ctx, http.MethodGet, "/api/tasks/"+id, nil, &raw); err != nil {
		return Task{}, err
	}

	t := Task{
		ID:     raw.ID,
		Name:   raw.Name,
		Status: raw.Status,
		Paused: raw.Paused,
		Queued: raw.Queued,
	}
	if err := checkOutputSize(t); err != nil {
		return Task{}, err
	}
	return t, nil
}
