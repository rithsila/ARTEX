package db

import (
	"database/sql"
	"time"
)

// ToolUsage is one catalog tool invocation. It stores attribution dimensions only:
// tool arguments and results are deliberately excluded from the ledger — out_len
// keeps just the result's byte size so "which tools bloat the context" stays
// answerable without storing result content.
type ToolUsage struct {
	ToolKey       string `json:"tool_key"`
	AgentKey      string `json:"agent_key"`
	TaskID        int64  `json:"task_id"`
	ExplorationID int64  `json:"exploration_id"`
	IntentID      int64  `json:"intent_id"`
	SessionID     string `json:"session_id"`
	OutLen        int    `json:"out_len"` // tool_result size in bytes
}

// InsertToolUsage appends one ledger row. Runtime callers treat metering as
// best-effort so a statistics failure never interrupts the tool itself.
func (d *DB) InsertToolUsage(u *ToolUsage) error {
	_, err := d.Exec(`
INSERT INTO tool_usage(tool_key, agent_key, task_id, exploration_id, intent_id, session_id, out_len)
VALUES ($1, NULLIF($2,''), $3, $4, $5, NULLIF($6,''), $7)`,
		u.ToolKey, u.AgentKey, nullIfZero(u.TaskID), nullIfZero(u.ExplorationID),
		nullIfZero(u.IntentID), u.SessionID, u.OutLen)
	return err
}

// ToolUsageStat is one catalog tool's usage-ledger aggregate — the token-hog
// ranking row. (Distinct from commands.go's ToolStat, the execution tally.)
type ToolUsageStat struct {
	ToolKey  string     `json:"tool_key"`
	Calls    int        `json:"calls"`
	OutAvg   int        `json:"out_avg"`  // average tool_result size in bytes
	OutSum   int64      `json:"out_sum"`  // total tool_result bytes — the ranking key
	LastUsed *time.Time `json:"last_used"` // nil for pre-out_len rows only
}

// ToolUsageStats ranks catalog tools by total tool_result bytes, biggest first —
// the input to "which compact formatters are worth writing". Live ledger rows
// only: archived-task aggregates keep call counts, not sizes, so tools whose
// history was archived rank by live traffic alone. Tools that never ran are
// absent; rows written before out_len existed count as 0 bytes.
func (d *DB) ToolUsageStats() ([]ToolUsageStat, error) {
	rows, err := d.Query(`
SELECT tool_key, COUNT(*),
       COALESCE(AVG(out_len), 0)::int,
       COALESCE(SUM(out_len), 0)::bigint,
       MAX(ts)
FROM tool_usage
GROUP BY tool_key
ORDER BY COALESCE(SUM(out_len), 0) DESC, tool_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ToolUsageStat{}
	for rows.Next() {
		var s ToolUsageStat
		var lastUsed sql.NullTime
		if err := rows.Scan(&s.ToolKey, &s.Calls, &s.OutAvg, &s.OutSum, &lastUsed); err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			t := lastUsed.Time
			s.LastUsed = &t
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ToolUsageCounts returns invocation totals keyed by catalog tool key. Entries
// without calls are absent; API callers merge the result into the tools catalog.
func (d *DB) ToolUsageCounts() (map[string]int, error) {
	rows, err := d.Query(`
SELECT tool_key, COUNT(*)
FROM tool_usage
GROUP BY tool_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var key string
		var calls int
		if err := rows.Scan(&key, &calls); err != nil {
			return nil, err
		}
		out[key] = calls
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	archived, err := d.archivedTaskAggregates()
	if err != nil {
		return nil, err
	}
	for _, aggregate := range archived {
		for key, calls := range aggregate.Tools {
			out[key] += calls
		}
	}
	return out, nil
}
