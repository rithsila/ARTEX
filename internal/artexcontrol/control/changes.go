package control

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type storedPreview struct {
	preview     ChangePreview
	req         ChangeRequest
	fingerprint string
	used        bool
}

// ChangeManager coordinates immutable previews, serialized apply, read-back verification, and audit trails.
type ChangeManager struct {
	service  *Service
	audit    Audit
	now      func() time.Time
	applyMu  sync.Mutex
	prevMu   sync.Mutex
	previews map[string]*storedPreview
}

// NewChangeManager creates a new ChangeManager.
func NewChangeManager(s *Service, audit Audit, now func() time.Time) *ChangeManager {
	if now == nil {
		now = time.Now
	}
	return &ChangeManager{
		service:  s,
		audit:    audit,
		now:      now,
		previews: make(map[string]*storedPreview),
	}
}

// StorePreviewForTest injects a preview into memory for testing edge cases.
func (m *ChangeManager) StorePreviewForTest(prev ChangePreview, fingerprint string) {
	m.prevMu.Lock()
	defer m.prevMu.Unlock()
	m.previews[prev.ChangeID] = &storedPreview{
		preview:     prev,
		req:         ChangeRequest{Operation: prev.Operation, Target: prev.Target, Key: prev.Key, Value: prev.After},
		fingerprint: fingerprint,
		used:        false,
	}
}

func getEffectNote(req ChangeRequest) string {
	switch req.Operation {
	case "setting":
		if req.Key == "workers" {
			return "new tasks"
		}
		if req.Key == "noa_compaction" {
			return "next agent run"
		}
		return "immediate persistence"
	case "agent_config":
		return "saved state verified; runtime adoption not verified"
	case "agent_prompt":
		return "saved; running-turn adoption not verified"
	case "profile_activate":
		return "saved state verified; runtime adoption not verified"
	case "task_control":
		return "task execution control"
	default:
		return "runtime change"
	}
}

// Preview creates an immutable proposed change preview without mutating state.
func (m *ChangeManager) Preview(ctx context.Context, req ChangeRequest) (ChangePreview, error) {
	if err := ValidateChange(req); err != nil {
		return ChangePreview{}, err
	}

	var beforeRaw json.RawMessage
	var afterRaw json.RawMessage
	var fingerprint string

	switch req.Operation {
	case "setting":
		settings, err := m.service.Settings(ctx)
		if err != nil {
			return ChangePreview{}, fmt.Errorf("failed reading settings: %w", err)
		}
		var curVal any
		switch req.Key {
		case "workers":
			curVal = settings.Workers
		case "traffic_capture":
			curVal = settings.TrafficCapture
		case "llm_record":
			curVal = settings.LLMRecord
		case "web_search_enabled":
			curVal = settings.WebSearchEnabled
		case "llm_pool_enabled":
			curVal = settings.LLMPoolEnabled
		case "llm_pool_bind_fallback":
			curVal = settings.LLMPoolBindFallback
		case "noa_compaction":
			curVal = settings.NOACompaction
		case "notify_enabled":
			curVal = settings.NotifyEnabled
		case "notify_digest_interval_min":
			curVal = settings.NotifyDigestIntervalMin
		}
		bBytes, _ := json.Marshal(curVal)
		beforeRaw = bBytes
		afterRaw = req.Value
		fingerprint = sha256Hex(bBytes)

	case "agent_config":
		detail, err := m.service.Agent(ctx, req.Target)
		if err != nil {
			return ChangePreview{}, fmt.Errorf("failed reading agent: %w", err)
		}
		var curVal any
		switch req.Key {
		case "max_turns":
			curVal = detail.MaxTurns
		case "run_seconds":
			curVal = detail.RunSeconds
		case "web_search":
			curVal = detail.WebSearch
		case "llm_profile_id":
			curVal = detail.LLMProfileID
		}
		bBytes, _ := json.Marshal(curVal)
		beforeRaw = bBytes
		afterRaw = req.Value
		fingerprint = sha256Hex(bBytes)

	case "agent_prompt":
		detail, err := m.service.Agent(ctx, req.Target)
		if err != nil {
			return ChangePreview{}, fmt.Errorf("failed reading agent: %w", err)
		}
		var p struct {
			Template string `json:"template"`
			Note     string `json:"note"`
		}
		_ = json.Unmarshal(req.Value, &p)
		bBytes, _ := json.Marshal(detail.Prompt)
		beforeRaw = bBytes
		aBytes, _ := json.Marshal(p.Template)
		afterRaw = aBytes
		fingerprint = sha256Hex([]byte(detail.Prompt))

	case "profile_activate":
		profiles, err := m.service.Profiles(ctx)
		if err != nil {
			return ChangePreview{}, fmt.Errorf("failed reading profiles: %w", err)
		}
		var targetFound bool
		var currentDefaultID string
		for _, p := range profiles {
			if p.ID == req.Target {
				targetFound = true
			}
			if p.IsDefault {
				currentDefaultID = p.ID
			}
		}
		if !targetFound {
			return ChangePreview{}, fmt.Errorf("profile %s not found", req.Target)
		}
		bBytes, _ := json.Marshal(currentDefaultID)
		beforeRaw = bBytes
		aBytes, _ := json.Marshal(req.Target)
		afterRaw = aBytes
		fingerprint = sha256Hex([]byte(currentDefaultID))

	case "task_control":
		task, err := m.service.Task(ctx, req.Target)
		if err != nil {
			return ChangePreview{}, fmt.Errorf("failed reading task: %w", err)
		}
		if req.Key == "pause" {
			if task.Paused {
				return ChangePreview{}, errors.New("task is already paused")
			}
			st := strings.ToLower(task.Status)
			if st == "done" || st == "failed" || st == "cancelled" || st == "error" {
				return ChangePreview{}, errors.New("cannot pause terminal task")
			}
		} else if req.Key == "resume" {
			if !task.Paused {
				return ChangePreview{}, errors.New("task is not paused")
			}
		}

		type taskState struct {
			Status string `json:"status"`
			Paused bool   `json:"paused"`
			Queued bool   `json:"queued"`
		}
		bBytes, _ := json.Marshal(taskState{
			Status: task.Status,
			Paused: task.Paused,
			Queued: task.Queued,
		})
		beforeRaw = bBytes
		aBytes, _ := json.Marshal(map[string]bool{"paused": req.Key == "pause"})
		afterRaw = aBytes
		fingerprint = sha256Hex(bBytes)
	}

	changeID := randomHex(16)
	expiresAt := m.now().Add(300 * time.Second)

	preview := ChangePreview{
		ChangeID:  changeID,
		Operation: req.Operation,
		Target:    req.Target,
		Key:       req.Key,
		Before:    beforeRaw,
		After:     afterRaw,
		Effect:    getEffectNote(req),
		ExpiresAt: expiresAt,
	}

	m.prevMu.Lock()
	defer m.prevMu.Unlock()

	// Purge expired previews before capacity check
	currentTime := m.now()
	for k, v := range m.previews {
		if currentTime.After(v.preview.ExpiresAt) || v.used {
			delete(m.previews, k)
		}
	}

	// 100-entry capacity limit
	if len(m.previews) >= 100 {
		return ChangePreview{}, errors.New("preview capacity reached (100 active entries)")
	}

	m.previews[changeID] = &storedPreview{
		preview:     preview,
		req:         req,
		fingerprint: fingerprint,
		used:        false,
	}

	return preview, nil
}

// computeCurrentFingerprint reads current state to check if the preview is stale.
func (m *ChangeManager) computeCurrentFingerprint(ctx context.Context, sp *storedPreview) (string, error) {
	switch sp.preview.Operation {
	case "setting":
		settings, err := m.service.Settings(ctx)
		if err != nil {
			return "", err
		}
		var curVal any
		switch sp.preview.Key {
		case "workers":
			curVal = settings.Workers
		case "traffic_capture":
			curVal = settings.TrafficCapture
		case "llm_record":
			curVal = settings.LLMRecord
		case "web_search_enabled":
			curVal = settings.WebSearchEnabled
		case "llm_pool_enabled":
			curVal = settings.LLMPoolEnabled
		case "llm_pool_bind_fallback":
			curVal = settings.LLMPoolBindFallback
		case "noa_compaction":
			curVal = settings.NOACompaction
		case "notify_enabled":
			curVal = settings.NotifyEnabled
		case "notify_digest_interval_min":
			curVal = settings.NotifyDigestIntervalMin
		}
		bBytes, _ := json.Marshal(curVal)
		return sha256Hex(bBytes), nil

	case "agent_config":
		detail, err := m.service.Agent(ctx, sp.preview.Target)
		if err != nil {
			return "", err
		}
		var curVal any
		switch sp.preview.Key {
		case "max_turns":
			curVal = detail.MaxTurns
		case "run_seconds":
			curVal = detail.RunSeconds
		case "web_search":
			curVal = detail.WebSearch
		case "llm_profile_id":
			curVal = detail.LLMProfileID
		}
		bBytes, _ := json.Marshal(curVal)
		return sha256Hex(bBytes), nil

	case "agent_prompt":
		detail, err := m.service.Agent(ctx, sp.preview.Target)
		if err != nil {
			return "", err
		}
		return sha256Hex([]byte(detail.Prompt)), nil

	case "profile_activate":
		profiles, err := m.service.Profiles(ctx)
		if err != nil {
			return "", err
		}
		var currentDefaultID string
		for _, p := range profiles {
			if p.IsDefault {
				currentDefaultID = p.ID
			}
		}
		return sha256Hex([]byte(currentDefaultID)), nil

	case "task_control":
		task, err := m.service.Task(ctx, sp.preview.Target)
		if err != nil {
			return "", err
		}
		type taskState struct {
			Status string `json:"status"`
			Paused bool   `json:"paused"`
			Queued bool   `json:"queued"`
		}
		bBytes, _ := json.Marshal(taskState{
			Status: task.Status,
			Paused: task.Paused,
			Queued: task.Queued,
		})
		return sha256Hex(bBytes), nil

	default:
		return "", errors.New("unsupported operation")
	}
}

// Apply executes a previously previewed change after re-checking freshness and logging intent.
func (m *ChangeManager) Apply(ctx context.Context, changeID, sessionID string) (ApplyResult, error) {
	m.applyMu.Lock()
	defer m.applyMu.Unlock()

	m.prevMu.Lock()
	sp, ok := m.previews[changeID]
	if !ok {
		m.prevMu.Unlock()
		return ApplyResult{}, fmt.Errorf("unknown change_id: %s", changeID)
	}
	if sp.used {
		m.prevMu.Unlock()
		return ApplyResult{}, errors.New("preview already used")
	}
	if m.now().After(sp.preview.ExpiresAt) {
		m.prevMu.Unlock()
		return ApplyResult{}, errors.New("preview has expired")
	}

	// Consume preview immediately so it cannot be reused
	sp.used = true
	m.prevMu.Unlock()

	// 1. Re-read current state to check if stale (if fingerprint changed -> abort without dispatch)
	currFP, err := m.computeCurrentFingerprint(ctx, sp)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("failed checking current state: %w", err)
	}
	if currFP != sp.fingerprint {
		return ApplyResult{}, errors.New("stale preview: state changed concurrently")
	}

	// 2. Audit intent
	opID := randomHex(16)
	intentEvent := AuditEvent{
		Timestamp:     m.now().UTC(),
		OperationID:   opID,
		SessionID:     sessionID,
		Operation:     sp.preview.Operation,
		Target:        sp.preview.Target,
		Key:           sp.preview.Key,
		BeforeHash:    sha256Hex(sp.preview.Before),
		AfterHash:     sha256Hex(sp.preview.After),
		DispatchState: "intent",
		Outcome:       "",
		HTTPStatus:    0,
	}

	if err := m.audit.Append(ctx, intentEvent); err != nil {
		return ApplyResult{}, fmt.Errorf("audit intent failed, blocking mutation: %w", err)
	}

	// 3. Dispatch single mutation
	var dispatchErr error
	httpStatus := 200

	switch sp.preview.Operation {
	case "setting":
		var parsedVal any
		raw := strings.TrimSpace(string(sp.preview.After))
		if raw == "true" {
			parsedVal = true
		} else if raw == "false" {
			parsedVal = false
		} else {
			num, _ := strconv.Atoi(raw)
			parsedVal = num
		}
		var dummy map[string]any
		dispatchErr = m.service.api.Do(ctx, http.MethodPut, "/api/settings", map[string]any{sp.preview.Key: parsedVal}, &dummy)

	case "agent_config":
		var body map[string]any
		if sp.preview.Key == "llm_profile_id" {
			raw := strings.TrimSpace(string(sp.preview.After))
			if raw == "null" {
				body = map[string]any{"llm_profile_id": nil}
			} else {
				var idStr string
				_ = json.Unmarshal(sp.preview.After, &idStr)
				idInt, _ := strconv.ParseInt(idStr, 10, 64)
				body = map[string]any{"llm_profile_id": idInt}
			}
		} else if sp.preview.Key == "web_search" {
			raw := strings.TrimSpace(string(sp.preview.After))
			body = map[string]any{"web_search": raw == "true"}
		} else {
			num, _ := strconv.Atoi(strings.TrimSpace(string(sp.preview.After)))
			body = map[string]any{sp.preview.Key: num}
		}
		var dummy map[string]any
		dispatchErr = m.service.api.Do(ctx, http.MethodPut, "/api/agents/"+sp.preview.Target+"/config", body, &dummy)

	case "agent_prompt":
		type promptPayload struct {
			Template string `json:"template"`
			Note     string `json:"note"`
		}
		var p promptPayload
		_ = json.Unmarshal(sp.req.Value, &p)
		var dummy map[string]any
		dispatchErr = m.service.api.Do(ctx, http.MethodPut, "/api/agents/"+sp.preview.Target+"/prompt", map[string]string{
			"template": p.Template,
			"note":     p.Note,
		}, &dummy)

	case "profile_activate":
		idInt, _ := strconv.ParseInt(sp.preview.Target, 10, 64)
		var dummy map[string]any
		dispatchErr = m.service.api.Do(ctx, http.MethodPost, "/api/llm/profiles/active", map[string]int64{"id": idInt}, &dummy)

	case "task_control":
		var dummy map[string]any
		dispatchErr = m.service.api.Do(ctx, http.MethodPost, "/api/tasks/"+sp.preview.Target+"/control", map[string]string{
			"action": sp.preview.Key,
		}, &dummy)
	}

	if dispatchErr != nil {
		httpStatus = 500
	}

	// 4. Detached context for read-back and final outcome audit
	detachedCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
	defer cancel()

	outcome := "outcome_unknown"
	var observedJSON json.RawMessage

	// Read-back verification
	switch sp.preview.Operation {
	case "setting":
		st, err := m.service.Settings(detachedCtx)
		if err == nil {
			var curVal any
			switch sp.preview.Key {
			case "workers":
				curVal = st.Workers
			case "traffic_capture":
				curVal = st.TrafficCapture
			case "llm_record":
				curVal = st.LLMRecord
			case "web_search_enabled":
				curVal = st.WebSearchEnabled
			case "llm_pool_enabled":
				curVal = st.LLMPoolEnabled
			case "llm_pool_bind_fallback":
				curVal = st.LLMPoolBindFallback
			case "noa_compaction":
				curVal = st.NOACompaction
			case "notify_enabled":
				curVal = st.NotifyEnabled
			case "notify_digest_interval_min":
				curVal = st.NotifyDigestIntervalMin
			}
			obsBytes, _ := json.Marshal(curVal)
			observedJSON = obsBytes

			if dispatchErr == nil {
				if bytes.Equal(obsBytes, sp.preview.After) {
					outcome = "verified"
				} else {
					outcome = "partial_or_failed"
				}
			} else {
				outcome = "partial_or_failed"
			}
		}

	case "agent_config":
		detail, err := m.service.Agent(detachedCtx, sp.preview.Target)
		if err == nil {
			var curVal any
			switch sp.preview.Key {
			case "max_turns":
				curVal = detail.MaxTurns
			case "run_seconds":
				curVal = detail.RunSeconds
			case "web_search":
				curVal = detail.WebSearch
			case "llm_profile_id":
				curVal = detail.LLMProfileID
			}
			obsBytes, _ := json.Marshal(curVal)
			observedJSON = obsBytes

			if dispatchErr == nil {
				if bytes.Equal(obsBytes, sp.preview.After) {
					outcome = "verified"
				} else {
					outcome = "partial_or_failed"
				}
			} else {
				outcome = "partial_or_failed"
			}
		}

	case "agent_prompt":
		detail, err := m.service.Agent(detachedCtx, sp.preview.Target)
		if err == nil {
			obsBytes, _ := json.Marshal(detail.Prompt)
			observedJSON = obsBytes

			if dispatchErr == nil {
				if bytes.Equal(obsBytes, sp.preview.After) {
					outcome = "verified"
				} else {
					outcome = "partial_or_failed"
				}
			} else {
				outcome = "partial_or_failed"
			}
		}

	case "profile_activate":
		profs, err := m.service.Profiles(detachedCtx)
		if err == nil {
			var activeID string
			for _, p := range profs {
				if p.IsDefault {
					activeID = p.ID
				}
			}
			obsBytes, _ := json.Marshal(activeID)
			observedJSON = obsBytes

			if dispatchErr == nil {
				if activeID == sp.preview.Target {
					outcome = "verified"
				} else {
					outcome = "partial_or_failed"
				}
			} else {
				outcome = "partial_or_failed"
			}
		}

	case "task_control":
		task, err := m.service.Task(detachedCtx, sp.preview.Target)
		if err == nil {
			type taskObs struct {
				Status string `json:"status"`
				Paused bool   `json:"paused"`
				Queued bool   `json:"queued"`
			}
			obsBytes, _ := json.Marshal(taskObs{Status: task.Status, Paused: task.Paused, Queued: task.Queued})
			observedJSON = obsBytes

			if dispatchErr == nil {
				if sp.preview.Key == "pause" {
					if task.Paused {
						outcome = "verified"
					} else {
						outcome = "partial_or_failed"
					}
				} else if sp.preview.Key == "resume" {
					if !task.Paused {
						outcome = "verified"
					} else {
						outcome = "partial_or_failed"
					}
				}
			} else {
				outcome = "partial_or_failed"
			}
		}
	}

	// 5. Audit final outcome
	outcomeEvent := AuditEvent{
		Timestamp:     m.now().UTC(),
		OperationID:   opID,
		SessionID:     sessionID,
		Operation:     sp.preview.Operation,
		Target:        sp.preview.Target,
		Key:           sp.preview.Key,
		BeforeHash:    sha256Hex(sp.preview.Before),
		AfterHash:     sha256Hex(sp.preview.After),
		DispatchState: "dispatched",
		Outcome:       outcome,
		HTTPStatus:    httpStatus,
	}

	auditErr := m.audit.Append(detachedCtx, outcomeEvent)
	auditIncomplete := auditErr != nil

	return ApplyResult{
		ChangeID:        changeID,
		Outcome:         outcome,
		Observed:        observedJSON,
		Effect:          sp.preview.Effect,
		AuditIncomplete: auditIncomplete,
	}, nil
}
