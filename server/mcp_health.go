package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Autumn-27/artex/db"
)

const mcpHealthTimeout = 8 * time.Second

// mcpHealthItem is one MCP card status. status is connected, disconnected, or off.
type mcpHealthItem struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// mcpHealthOf probes one server with a short timeout. A disabled server is off
// and is not dialed. Success is the MCP initialize handshake, not a tool list.
func mcpHealthOf(ctx context.Context, m *db.MCPServer, timeout time.Duration) mcpHealthItem {
	if m == nil {
		return mcpHealthItem{Status: "off"}
	}
	item := mcpHealthItem{ID: m.ID, Status: "off"}
	if !m.Enabled {
		return item
	}
	if timeout <= 0 {
		timeout = mcpHealthTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cl, err := connectMCP(ctx, m)
	if err != nil {
		item.Status = "disconnected"
		item.Error = oneLineMCPError(err)
		return item
	}
	_ = cl.Close()
	item.Status = "connected"
	return item
}

func oneLineMCPError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if i := strings.IndexAny(msg, "\r\n"); i >= 0 {
		msg = strings.TrimSpace(msg[:i])
	}
	const max = 160
	if utf8.RuneCountInString(msg) <= max {
		return msg
	}
	runes := []rune(msg)
	return string(runes[:max-1]) + "…"
}

func (s *Server) pgMCPHealth(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	ms, err := pg.ListMCP()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	items := make([]mcpHealthItem, len(ms))
	var wg sync.WaitGroup
	for i, m := range ms {
		wg.Add(1)
		go func(i int, m *db.MCPServer) {
			defer wg.Done()
			items[i] = mcpHealthOf(r.Context(), m, mcpHealthTimeout)
		}(i, m)
	}
	wg.Wait()
	if items == nil {
		items = []mcpHealthItem{}
	}
	writeJSON(w, 200, map[string]any{"servers": items})
}
