package mcphttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A tool call must outlive the old 2-minute HTTP cap. Worker runs default to
// 20 minutes (agents.run_seconds). The client timeout is the cap that fired
// as "Client.Timeout exceeded while awaiting headers".
func TestStreamableHTTPToolCallOutlivesTwoMinuteCap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.ID == 0 {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{}}`, req.ID)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c, err := New(ctx, "hexstrike", srv.URL+"/mcp", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if c.http.Timeout < 20*time.Minute {
		t.Fatalf("tool call HTTP cap is %s, want at least 20m so a worker run is not cut at 2m", c.http.Timeout)
	}
}
