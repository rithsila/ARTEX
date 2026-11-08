package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Autumn-27/artex/db"
)

func TestMCPHealthConnected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID any `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.ID == nil || req.ID == float64(0) {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%v,"result":{}}`, req.ID)
	}))
	defer srv.Close()

	item := mcpHealthOf(context.Background(), &db.MCPServer{
		ID: 7, Name: "hexstrike", Transport: "http", URL: srv.URL + "/mcp", Enabled: true,
	}, time.Second)
	if item.Status != "connected" || item.Error != "" || item.ID != 7 {
		t.Fatalf("got %+v, want connected with no error", item)
	}
}

func TestMCPHealthOffSkipsProbe(t *testing.T) {
	item := mcpHealthOf(context.Background(), &db.MCPServer{
		ID: 3, Name: "rea", Transport: "http", URL: "http://127.0.0.1:1/mcp", Enabled: false,
	}, time.Second)
	if item.Status != "off" || item.Error != "" {
		t.Fatalf("got %+v, want off", item)
	}
}

func TestMCPHealthDisconnectedIsOneLine(t *testing.T) {
	item := mcpHealthOf(context.Background(), &db.MCPServer{
		ID: 4, Name: "down", Transport: "http", URL: "http://127.0.0.1:1/mcp", Enabled: true,
	}, time.Second)
	if item.Status != "disconnected" || item.Error == "" || strings.Contains(item.Error, "\n") {
		t.Fatalf("got %+v, want one-line disconnected error", item)
	}
}
