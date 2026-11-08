package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
)

func TestMCPConnectionFailureTelemetry(t *testing.T) {
	dsn, _, err := db.DSN()
	if err != nil {
		t.Skipf("no database config (%v) — skipping", err)
	}
	pg, err := db.Open(dsn)
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer pg.Close()

	// Setup a task to link the fact to
	task, err := pg.CreateTask("telemetry test", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer pg.DeleteTask(task.ID)

	// Register a failing MCP server
	mcpID, err := pg.SaveMCP(&db.MCPServer{
		Name:      "failing-mcp",
		Transport: "http",
		URL:       "http://127.0.0.1:54321/mcp",
		Enabled:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pg.DeleteMCP(mcpID)

	// Make MCP visible to planner
	ag, _ := pg.GetAgentByKey("planner")
	if err := pg.SetAgentVisibilityKind(ag.ID, "mcp", []int64{mcpID}); err != nil {
		t.Fatal(err)
	}
	defer pg.SetAgentVisibilityKind(ag.ID, "mcp", nil)

	// Augment tools with a context that has the TaskID
	ctx := agent.WithRunInfo(context.Background(), agent.RunInfo{TaskID: task.ID})
	wireAgentAugment(pg, t.TempDir(), nil)
	t.Cleanup(func() { agent.ToolAugment = nil })

	_, _, cleanup := agent.ToolAugment(ctx, "planner")
	defer cleanup()

	// Verify fact exists
	exp := pg.Exploration(task.ExplorationID)
	facts, err := exp.ListByKind(db.KindFact, 10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range facts {
		var p map[string]any
		_ = json.Unmarshal(f.Payload, &p)
		if sm, ok := p["summary"].(string); ok && strings.Contains(sm, "failing-mcp") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected telemetry fact for failed MCP connection")
	}
}
