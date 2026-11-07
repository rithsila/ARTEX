package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	actool "github.com/Autumn-27/norma/tool"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
)

type recordingToolUsage struct {
	rows []*db.ToolUsage
	err  error
}

func (r *recordingToolUsage) InsertToolUsage(row *db.ToolUsage) error {
	copy := *row
	r.rows = append(r.rows, &copy)
	return r.err
}

func TestMeterToolRecordsAttributionAndDelegates(t *testing.T) {
	calls := 0
	base := actool.Build(actool.Spec{
		Name: "record_fact",
		Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
			calls++
			return actool.Result{}, nil
		},
	})
	recorder := &recordingToolUsage{}
	ri := agent.RunInfo{TaskID: 42, ExplorationID: 8, IntentID: 9, SessionID: "session-1"}
	wrapped := meterTool(base, recorder, "record_fact", "worker", ri)

	if _, err := wrapped.Call(context.Background(), json.RawMessage(`{"secret":"not persisted"}`), nil); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("delegate calls: want 1, got %d", calls)
	}
	if len(recorder.rows) != 1 {
		t.Fatalf("usage rows: want 1, got %d", len(recorder.rows))
	}
	got := recorder.rows[0]
	if got.ToolKey != "record_fact" || got.AgentKey != "worker" || got.TaskID != 42 ||
		got.ExplorationID != 8 || got.IntentID != 9 || got.SessionID != "session-1" {
		t.Fatalf("unexpected attribution: %+v", got)
	}
	if got.OutLen != 0 {
		t.Fatalf("empty result out_len: want 0, got %d", got.OutLen)
	}
}

func TestMeterToolRecordsOutLenFromResultText(t *testing.T) {
	base := actool.Build(actool.Spec{
		Name: "list_facts",
		Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
			return actool.Text("hello world"), nil
		},
	})
	recorder := &recordingToolUsage{}
	wrapped := meterTool(base, recorder, "list_facts", "worker", agent.RunInfo{})

	if _, err := wrapped.Call(context.Background(), json.RawMessage(`{}`), nil); err != nil {
		t.Fatal(err)
	}
	if len(recorder.rows) != 1 {
		t.Fatalf("usage rows: want 1, got %d", len(recorder.rows))
	}
	if want := len("hello world"); recorder.rows[0].OutLen != want {
		t.Fatalf("out_len: want %d, got %d", want, recorder.rows[0].OutLen)
	}
}

func TestMeterToolFailureDoesNotBreakInvocation(t *testing.T) {
	calls := 0
	base := actool.Build(actool.Spec{
		Name: "list_facts",
		Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
			calls++
			return actool.Result{}, nil
		},
	})
	recorder := &recordingToolUsage{err: errors.New("ledger unavailable")}
	wrapped := meterTool(base, recorder, "list_facts", "worker", agent.RunInfo{})

	if _, err := wrapped.Call(context.Background(), nil, nil); err != nil {
		t.Fatalf("metering error leaked into tool call: %v", err)
	}
	if calls != 1 {
		t.Fatalf("delegate calls: want 1, got %d", calls)
	}
}
