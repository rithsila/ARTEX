package server

import (
	"context"
	"encoding/json"
	"log"

	actool "github.com/Autumn-27/norma/tool"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
)

type toolUsageRecorder interface {
	InsertToolUsage(*db.ToolUsage) error
}

// meteredResultLen measures what a tool result injects into the model context:
// the summed text of its Content blocks. Extra messages (skill instructions the
// Skill tool appends as standalone user messages) are excluded — they are
// prompts, not tool output. Error results carry their text in Content too, so
// failed calls are measured like any other.
func meteredResultLen(res actool.Result) int {
	n := 0
	for _, b := range res.Content {
		n += len(b.Text)
	}
	return n
}

// meteredTool records one row after a catalog tool completes — out_len needs the
// result in hand. It embeds the resolved tool so schema overrides, defaults and
// permission behavior remain unchanged. The ledger row is still best-effort: a
// metering failure logs and never breaks the call. The one semantic shift from
// metering-before: a call that never returns (hang/panic) is not ledgered.
type meteredTool struct {
	actool.CoreTool
	recorder toolUsageRecorder
	toolKey  string
	agentKey string
	ri       agent.RunInfo
}

func meterTool(t actool.CoreTool, recorder toolUsageRecorder, toolKey, agentKey string, ri agent.RunInfo) actool.CoreTool {
	if recorder == nil {
		return t
	}
	return &meteredTool{CoreTool: t, recorder: recorder, toolKey: toolKey, agentKey: agentKey, ri: ri}
}

func (m *meteredTool) Call(ctx context.Context, input json.RawMessage, tc *actool.ToolContext) (actool.Result, error) {
	res, err := m.CoreTool.Call(ctx, input, tc)
	if err := m.recorder.InsertToolUsage(&db.ToolUsage{
		ToolKey:       m.toolKey,
		AgentKey:      m.agentKey,
		TaskID:        m.ri.TaskID,
		ExplorationID: m.ri.ExplorationID,
		IntentID:      m.ri.IntentID,
		SessionID:     m.ri.SessionID,
		OutLen:        meteredResultLen(res),
	}); err != nil {
		log.Printf("[toolusage] insert %s: %v", m.toolKey, err)
	}
	return res, err
}
