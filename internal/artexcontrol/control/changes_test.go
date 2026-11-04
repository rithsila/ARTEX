package control_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Autumn-27/artex/internal/artexcontrol/control"
)

type memoryAudit struct {
	mu           sync.Mutex
	events       []control.AuditEvent
	failIntent   bool
	failOutcome  bool
	intentCalls  int
	outcomeCalls int
}

func (m *memoryAudit) Append(ctx context.Context, ev control.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ev.DispatchState == "intent" {
		m.intentCalls++
		if m.failIntent {
			return errors.New("simulated intent audit failure")
		}
	} else {
		m.outcomeCalls++
		if m.failOutcome {
			return errors.New("simulated outcome audit failure")
		}
	}
	m.events = append(m.events, ev)
	return nil
}

func TestPreviewNeverMutates(t *testing.T) {
	fakeNow := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return fakeNow }

	ops := []control.ChangeRequest{
		{
			Operation: "setting",
			Key:       "workers",
			Value:     json.RawMessage(`4`),
		},
		{
			Operation: "agent_config",
			Target:    "planner",
			Key:       "max_turns",
			Value:     json.RawMessage(`20`),
		},
		{
			Operation: "agent_prompt",
			Target:    "planner",
			Value:     json.RawMessage(`{"template":"New template","note":"test"}`),
		},
		{
			Operation: "profile_activate",
			Target:    "42",
		},
		{
			Operation: "task_control",
			Target:    "100",
			Key:       "pause",
		},
	}

	for _, req := range ops {
		t.Run(req.Operation, func(t *testing.T) {
			api := &mockAPI{
				doFunc: func(ctx context.Context, method, path string, body any, out any) error {
					if method != http.MethodGet {
						t.Fatalf("preview executed non-GET request: %s %s", method, path)
					}
					switch path {
					case "/api/settings":
						return json.Unmarshal([]byte(`{"workers":2}`), out)
					case "/api/agents/planner":
						return json.Unmarshal([]byte(`{"agent":{"key":"planner","id":"1","name":"Planner","max_turns":10},"prompt":"Old template"}`), out)
					case "/api/llm/profiles":
						return json.Unmarshal([]byte(`{"profiles":[{"id":"42","name":"Target","model":"m","format":"f","is_default":false}]}`), out)
					case "/api/tasks/100":
						return json.Unmarshal([]byte(`{"id":"100","name":"Scan","status":"running","paused":false,"queued":false}`), out)
					}
					return nil
				},
			}

			svc := control.NewService(api)
			audit := &memoryAudit{}
			mgr := control.NewChangeManager(svc, audit, nowFn)

			prev, err := mgr.Preview(context.Background(), req)
			if err != nil {
				t.Fatalf("Preview failed: %v", err)
			}

			if prev.ChangeID == "" {
				t.Errorf("expected non-empty ChangeID")
			}
			if !prev.ExpiresAt.Equal(fakeNow.Add(300 * time.Second)) {
				t.Errorf("expected ExpiresAt = now+300s, got %v", prev.ExpiresAt)
			}
			if len(prev.Before) == 0 {
				t.Errorf("expected non-empty Before state")
			}
			if len(prev.After) == 0 {
				t.Errorf("expected non-empty After state")
			}
			if prev.Effect == "" {
				t.Errorf("expected non-empty Effect description")
			}
		})
	}
}

func TestApplyDispatchMapping(t *testing.T) {
	fakeNow := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return fakeNow }

	largeID := "9007199254740993"

	cases := []struct {
		name       string
		req        control.ChangeRequest
		wantMethod string
		wantPath   string
		checkBody  func(t *testing.T, body any)
	}{
		{
			name: "setting PUT single key",
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "workers",
				Value:     json.RawMessage(`5`),
			},
			wantMethod: http.MethodPut,
			wantPath:   "/api/settings",
			checkBody: func(t *testing.T, body any) {
				m, ok := body.(map[string]any)
				if !ok {
					t.Fatalf("expected map[string]any, got %T", body)
				}
				if len(m) != 1 || fmt.Sprint(m["workers"]) != "5" {
					t.Errorf("expected single key workers=5, got %+v", m)
				}
			},
		},
		{
			name: "agent config PUT single field with large ID",
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "llm_profile_id",
				Value:     json.RawMessage(fmt.Sprintf(`"%s"`, largeID)),
			},
			wantMethod: http.MethodPut,
			wantPath:   "/api/agents/planner/config",
			checkBody: func(t *testing.T, body any) {
				data, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("marshal body: %v", err)
				}
				dec := json.NewDecoder(strings.NewReader(string(data)))
				dec.UseNumber()
				var m map[string]json.Number
				if err := dec.Decode(&m); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if m["llm_profile_id"].String() != largeID {
					t.Errorf("expected lossless numeric ID %s, got %s", largeID, m["llm_profile_id"].String())
				}
			},
		},
		{
			name: "agent prompt PUT template and note only",
			req: control.ChangeRequest{
				Operation: "agent_prompt",
				Target:    "planner",
				Value:     json.RawMessage(`{"template":"New text","note":"v2"}`),
			},
			wantMethod: http.MethodPut,
			wantPath:   "/api/agents/planner/prompt",
			checkBody: func(t *testing.T, body any) {
				m, ok := body.(map[string]string)
				if !ok {
					t.Fatalf("expected map[string]string, got %T", body)
				}
				if len(m) != 2 || m["template"] != "New text" || m["note"] != "v2" {
					t.Errorf("expected template and note only, got %+v", m)
				}
			},
		},
		{
			name: "profile activate POST lossless numeric ID",
			req: control.ChangeRequest{
				Operation: "profile_activate",
				Target:    largeID,
			},
			wantMethod: http.MethodPost,
			wantPath:   "/api/llm/profiles/active",
			checkBody: func(t *testing.T, body any) {
				data, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
				dec := json.NewDecoder(strings.NewReader(string(data)))
				dec.UseNumber()
				var m map[string]json.Number
				if err := dec.Decode(&m); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if m["id"].String() != largeID {
					t.Errorf("expected lossless numeric ID %s, got %s", largeID, m["id"].String())
				}
			},
		},
		{
			name: "task control POST action pause",
			req: control.ChangeRequest{
				Operation: "task_control",
				Target:    "77",
				Key:       "pause",
			},
			wantMethod: http.MethodPost,
			wantPath:   "/api/tasks/77/control",
			checkBody: func(t *testing.T, body any) {
				m, ok := body.(map[string]string)
				if !ok {
					t.Fatalf("expected map[string]string, got %T", body)
				}
				if m["action"] != "pause" {
					t.Errorf("expected action=pause, got %+v", m)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var dispatchedMutation mockCall
			api := &mockAPI{
				doFunc: func(ctx context.Context, method, path string, body any, out any) error {
					if method != http.MethodGet {
						dispatchedMutation = mockCall{Method: method, Path: path, Body: body}
						return nil
					}
					// Return reads for preview and read-back
					switch path {
					case "/api/settings":
						return json.Unmarshal([]byte(`{"workers":5}`), out)
					case "/api/agents/planner":
						return json.Unmarshal([]byte(fmt.Sprintf(`{"agent":{"key":"planner","id":"1","name":"Planner","llm_profile_id":%s},"prompt":"New text"}`, largeID)), out)
					case "/api/llm/profiles":
						return json.Unmarshal([]byte(fmt.Sprintf(`{"profiles":[{"id":"%s","name":"P","model":"m","format":"f","is_default":true}]}`, largeID)), out)
					case "/api/tasks/77":
						paused := dispatchedMutation.Method != ""
						return json.Unmarshal([]byte(fmt.Sprintf(`{"id":"77","name":"T","status":"running","paused":%t,"queued":false}`, paused)), out)
					}
					return nil
				},
			}

			svc := control.NewService(api)
			audit := &memoryAudit{}
			mgr := control.NewChangeManager(svc, audit, nowFn)

			prev, err := mgr.Preview(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("Preview failed: %v", err)
			}

			res, err := mgr.Apply(context.Background(), prev.ChangeID, "ses-1")
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			if res.Outcome != "verified" {
				t.Errorf("expected outcome verified, got %s", res.Outcome)
			}
			if dispatchedMutation.Method != tc.wantMethod {
				t.Errorf("expected method %s, got %s", tc.wantMethod, dispatchedMutation.Method)
			}
			if dispatchedMutation.Path != tc.wantPath {
				t.Errorf("expected path %s, got %s", tc.wantPath, dispatchedMutation.Path)
			}
			tc.checkBody(t, dispatchedMutation.Body)
		})
	}
}

func TestPreviewExpiryAndCapacity(t *testing.T) {
	currTime := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return currTime }

	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			return json.Unmarshal([]byte(`{"workers":3}`), out)
		},
	}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, nowFn)

	// 1. Create 100 previews
	var firstID string
	for i := 0; i < 100; i++ {
		prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
			Operation: "setting",
			Key:       "workers",
			Value:     json.RawMessage(`4`),
		})
		if err != nil {
			t.Fatalf("preview %d failed: %v", i, err)
		}
		if i == 0 {
			firstID = prev.ChangeID
		}
	}

	// 2. Advance time past 300s -> all 100 are now expired
	currTime = currTime.Add(301 * time.Second)

	// Attempting to apply an expired preview fails
	_, err := mgr.Apply(context.Background(), firstID, "ses")
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expired error, got: %v", err)
	}

	// Adding 101st preview purges expired entries and succeeds without hitting capacity
	_, err = mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`5`),
	})
	if err != nil {
		t.Fatalf("preview after expiry failed: %v", err)
	}
}

func TestStalePreviewNoWrite(t *testing.T) {
	nowFn := func() time.Time { return time.Now().UTC() }
	workersVal := 3
	var wrote bool

	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodGet && path == "/api/settings" {
				return json.Unmarshal([]byte(fmt.Sprintf(`{"workers":%d}`, workersVal)), out)
			}
			if method == http.MethodPut {
				wrote = true
				return nil
			}
			return nil
		},
	}

	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, nowFn)

	// 1. Preview when workers = 3
	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}

	// 2. Concurrently workers changed on backend to 10
	workersVal = 10

	// 3. Apply must detect fingerprint change and refuse to write
	_, err = mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Errorf("expected stale error, got: %v", err)
	}
	if wrote {
		t.Errorf("stale preview should not have dispatched mutation")
	}
}

func TestConcurrentApplyWritesOnce(t *testing.T) {
	nowFn := func() time.Time { return time.Now().UTC() }
	var mutationCount int32

	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodGet {
				return json.Unmarshal([]byte(`{"workers":4}`), out)
			}
			if method == http.MethodPut {
				atomic.AddInt32(&mutationCount, 1)
				time.Sleep(50 * time.Millisecond)
				return nil
			}
			return nil
		},
	}

	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, nowFn)

	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}

	var wg sync.WaitGroup
	var successCount int32
	var errCount int32

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := mgr.Apply(context.Background(), prev.ChangeID, "ses")
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&errCount, 1)
			}
		}()
	}
	wg.Wait()

	if mutationCount != 1 {
		t.Errorf("expected exactly 1 mutation dispatched, got %d", mutationCount)
	}
	if successCount != 1 {
		t.Errorf("expected exactly 1 successful apply, got %d", successCount)
	}
	if errCount != 4 {
		t.Errorf("expected 4 rejected applies, got %d", errCount)
	}
}

func TestUnknownIDNoWrite(t *testing.T) {
	var wrote bool
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method != http.MethodGet {
				wrote = true
			}
			return nil
		},
	}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

	_, err := mgr.Apply(context.Background(), "unknown-change-id", "ses")
	if err == nil {
		t.Errorf("expected error for unknown ID, got nil")
	}
	if wrote {
		t.Errorf("mutation dispatched on unknown change ID")
	}
}

func TestAuditIntentFailureBlocksWrite(t *testing.T) {
	var wrote bool
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodGet {
				return json.Unmarshal([]byte(`{"workers":3}`), out)
			}
			wrote = true
			return nil
		},
	}
	svc := control.NewService(api)
	audit := &memoryAudit{failIntent: true}
	mgr := control.NewChangeManager(svc, audit, time.Now)

	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}

	_, err = mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err == nil || !strings.Contains(err.Error(), "audit") {
		t.Errorf("expected audit intent failure, got: %v", err)
	}
	if wrote {
		t.Errorf("mutation must not dispatch if audit intent fails")
	}
}

func TestOutcomeAuditFailureReported(t *testing.T) {
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodGet {
				return json.Unmarshal([]byte(`{"workers":4}`), out)
			}
			return nil
		},
	}
	svc := control.NewService(api)
	audit := &memoryAudit{failOutcome: true}
	mgr := control.NewChangeManager(svc, audit, time.Now)

	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}

	res, err := mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !res.AuditIncomplete {
		t.Errorf("expected AuditIncomplete to be true")
	}
	if res.Outcome != "verified" {
		t.Errorf("expected verified outcome, got %s", res.Outcome)
	}
}

func TestWriteSavedThen500(t *testing.T) {
	// Server returned 500 on PUT, but read-back shows the workers value actually changed to 4!
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodPut {
				return errors.New("500 internal server error after DB write")
			}
			if method == http.MethodGet {
				return json.Unmarshal([]byte(`{"workers":4}`), out)
			}
			return nil
		},
	}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}

	res, err := mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	// Per spec: 500 error after mutation is partial_or_failed even if observed state matches!
	if res.Outcome != "partial_or_failed" {
		t.Errorf("expected partial_or_failed for 500 response, got %s", res.Outcome)
	}
}

func TestTimeoutAfterPossibleCommit(t *testing.T) {
	var mutationAttempted bool
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodPut {
				mutationAttempted = true
				return errors.New("i/o timeout")
			}
			if !mutationAttempted {
				return json.Unmarshal([]byte(`{"workers":3}`), out)
			}
			return errors.New("i/o timeout")
		},
	}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}

	res, err := mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if res.Outcome != "outcome_unknown" {
		t.Errorf("expected outcome_unknown on timeout with lost read-back, got %s", res.Outcome)
	}
}

func TestReadBackFailureUnknown(t *testing.T) {
	var mutationDone bool
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodPut {
				mutationDone = true
				return nil // success
			}
			if !mutationDone {
				return json.Unmarshal([]byte(`{"workers":3}`), out)
			}
			return errors.New("read-back failed")
		},
	}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}

	res, err := mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.Outcome != "outcome_unknown" {
		t.Errorf("expected outcome_unknown when read-back fails, got %s", res.Outcome)
	}
}

func TestVerificationMismatch(t *testing.T) {
	// Dispatch succeeded (200), but read-back shows old value (3 instead of 4)
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodPut {
				return nil
			}
			return json.Unmarshal([]byte(`{"workers":3}`), out)
		},
	}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

	prev, _ := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})

	res, err := mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.Outcome != "partial_or_failed" {
		t.Errorf("expected partial_or_failed on mismatch, got %s", res.Outcome)
	}
}

func TestQueuedResumeReported(t *testing.T) {
	var resumed bool
	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodPost {
				resumed = true
				return nil
			}
			if !resumed {
				return json.Unmarshal([]byte(`{"id":"99","name":"Scan","status":"paused","paused":true,"queued":false}`), out)
			}
			return json.Unmarshal([]byte(`{"id":"99","name":"Scan","status":"queued","paused":false,"queued":true}`), out)
		},
	}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

	prev, err := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "task_control",
		Target:    "99",
		Key:       "resume",
	})
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	res, err := mgr.Apply(context.Background(), prev.ChangeID, "ses")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.Outcome != "verified" {
		t.Errorf("expected verified for observed unpaused queued task, got %s", res.Outcome)
	}
}

func TestTerminalOrAlreadyPausedRejected(t *testing.T) {
	cases := []struct {
		name      string
		action    string
		taskJSON  string
		wantError string
	}{
		{
			name:      "pause already paused",
			action:    "pause",
			taskJSON:  `{"id":"10","name":"T","status":"paused","paused":true,"queued":false}`,
			wantError: "task is already paused",
		},
		{
			name:      "pause terminal task done",
			action:    "pause",
			taskJSON:  `{"id":"10","name":"T","status":"done","paused":false,"queued":false}`,
			wantError: "cannot pause terminal task",
		},
		{
			name:      "resume not paused",
			action:    "resume",
			taskJSON:  `{"id":"10","name":"T","status":"running","paused":false,"queued":false}`,
			wantError: "task is not paused",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api := &mockAPI{
				doFunc: func(ctx context.Context, method, path string, body any, out any) error {
					return json.Unmarshal([]byte(tc.taskJSON), out)
				},
			}
			svc := control.NewService(api)
			mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

			_, err := mgr.Preview(context.Background(), control.ChangeRequest{
				Operation: "task_control",
				Target:    "10",
				Key:       tc.action,
			})
			if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("expected error containing %q, got: %v", tc.wantError, err)
			}
		})
	}
}

func TestCallerCancellationAfterCommit(t *testing.T) {
	// Caller context canceled immediately after mutation HTTP request returns
	ctx, cancel := context.WithCancel(context.Background())
	var auditCalls int32

	api := &mockAPI{
		doFunc: func(ctx context.Context, method, path string, body any, out any) error {
			if method == http.MethodPut {
				cancel() // Cancel caller context immediately after mutation
				return nil
			}
			return json.Unmarshal([]byte(`{"workers":4}`), out)
		},
	}
	audit := &memoryAudit{}
	svc := control.NewService(api)
	mgr := control.NewChangeManager(svc, audit, time.Now)

	prev, _ := mgr.Preview(context.Background(), control.ChangeRequest{
		Operation: "setting",
		Key:       "workers",
		Value:     json.RawMessage(`4`),
	})

	res, err := mgr.Apply(ctx, prev.ChangeID, "ses")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.Outcome != "verified" {
		t.Errorf("expected verified outcome despite caller cancellation, got %s", res.Outcome)
	}
	if len(audit.events) != 2 {
		t.Errorf("expected 2 audit events (intent + outcome), got %d", len(audit.events))
	}
	_ = auditCalls
}

func TestRuntimeEffectNotes(t *testing.T) {
	cases := []struct {
		req        control.ChangeRequest
		wantEffect string
	}{
		{
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "workers",
				Value:     json.RawMessage(`4`),
			},
			wantEffect: "new tasks",
		},
		{
			req: control.ChangeRequest{
				Operation: "setting",
				Key:       "noa_compaction",
				Value:     json.RawMessage(`true`),
			},
			wantEffect: "next agent run",
		},
		{
			req: control.ChangeRequest{
				Operation: "agent_prompt",
				Target:    "planner",
				Value:     json.RawMessage(`{"template":"text","note":"v"}`),
			},
			wantEffect: "saved; running-turn adoption not verified",
		},
		{
			req: control.ChangeRequest{
				Operation: "agent_config",
				Target:    "planner",
				Key:       "max_turns",
				Value:     json.RawMessage(`15`),
			},
			wantEffect: "saved state verified; runtime adoption not verified",
		},
		{
			req: control.ChangeRequest{
				Operation: "profile_activate",
				Target:    "1",
			},
			wantEffect: "saved state verified; runtime adoption not verified",
		},
	}

	for _, tc := range cases {
		t.Run(tc.req.Operation+"_"+tc.req.Key, func(t *testing.T) {
			api := &mockAPI{
				doFunc: func(ctx context.Context, method, path string, body any, out any) error {
					switch path {
					case "/api/settings":
						return json.Unmarshal([]byte(`{"workers":2,"noa_compaction":false}`), out)
					case "/api/agents/planner":
						return json.Unmarshal([]byte(`{"agent":{"key":"planner","id":"1","name":"P","max_turns":10},"prompt":"old"}`), out)
					case "/api/llm/profiles":
						return json.Unmarshal([]byte(`{"profiles":[{"id":"1","name":"P","model":"m","format":"f","is_default":false}]}`), out)
					}
					return nil
				},
			}
			svc := control.NewService(api)
			mgr := control.NewChangeManager(svc, &memoryAudit{}, time.Now)

			prev, err := mgr.Preview(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("Preview: %v", err)
			}
			if prev.Effect != tc.wantEffect {
				t.Errorf("expected Effect %q, got %q", tc.wantEffect, prev.Effect)
			}
		})
	}
}
