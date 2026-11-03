package server

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestSSSyncToolSchemaAndResult(t *testing.T) {
	tool := newSSSyncTool(func(ctx context.Context, req ssSyncReq) (ssSyncResponse, error) {
		if req.Dimension != "task" || !reflect.DeepEqual(req.Targets, []string{"scan"}) || req.PageSize != 50 {
			t.Fatalf("request lost: %+v", req)
		}
		return ssSyncResponse{Synced: map[string]int{"subdomain": 2}}, nil
	})
	if tool.Name() != "sync_scopesentry_assets" || tool.IsReadOnly(nil) || tool.IsConcurrencySafe(nil) {
		t.Fatal("sync must be a write tool")
	}
	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)
	if len(props) != 5 || !reflect.DeepEqual(schema["required"], []any{"dimension", "targets"}) {
		t.Fatalf("bad schema: %#v", schema)
	}
	for key, kind := range map[string]string{"dimension": "string", "targets": "array", "asset_types": "array", "create_company": "boolean", "page_size": "integer"} {
		if props[key].(map[string]any)["type"] != kind {
			t.Fatalf("wrong schema for %s", key)
		}
	}
	result, err := tool.Call(context.Background(), json.RawMessage(`{"dimension":"task","targets":["scan"],"page_size":50}`), nil)
	var response ssSyncResponse
	if err != nil || result.IsError || json.Unmarshal([]byte(result.Flatten()), &response) != nil || !reflect.DeepEqual(response.Synced, map[string]int{"subdomain": 2}) {
		t.Fatalf("result lost: %s %v", result.Flatten(), err)
	}
}

func TestSSSyncToolRejectsInvalidJSONAndSelection(t *testing.T) {
	tool := newSSSyncTool(func(context.Context, ssSyncReq) (ssSyncResponse, error) {
		t.Fatal("invalid input reached operation")
		return ssSyncResponse{}, nil
	})
	for _, input := range []string{`{`, `null`, `{"dimension":"project","targets":["invalid"]}`, `{"dimension":"task","targets":[""]}`} {
		result, err := tool.Call(context.Background(), json.RawMessage(input), nil)
		data, _ := json.Marshal(result)
		if err != nil || !result.IsError {
			t.Fatalf("expected tool error: %s %v", data, err)
		}
	}
}

func TestSSSyncHostToolCatalogIsOptIn(t *testing.T) {
	s := &Server{}
	seeds := s.scopeSentryToolSeeds()
	if len(seeds) != 1 || seeds[0].Key != "sync_scopesentry_assets" || len(seeds[0].Agents) != 0 || seeds[0].Agents == nil {
		t.Fatalf("tool must seed with explicit empty bindings: %+v", seeds)
	}
	result, err := s.toolSyncScopeSentry().Call(context.Background(), json.RawMessage(`{"dimension":"task","targets":["scan"]}`), nil)
	data, _ := json.Marshal(result)
	if err != nil || !strings.Contains(string(data), "database unavailable") {
		t.Fatalf("unavailable store: %s %v", data, err)
	}
}

func TestSSSyncHTTPValidatesBeforeStoreAccess(t *testing.T) {
	s := &Server{}
	for _, tc := range []struct {
		input   string
		status  int
		message string
	}{
		{`{`, 400, "invalid JSON:"},
		{`{"dimension":"task","targets":[]}`, 400, "targets cannot be empty"},
		{`{"dimension":"project","targets":["bad-id"]}`, 400, "project targets must be ObjectIDs"},
		{`{"dimension":"task","targets":["scan"]}`, 503, "database unavailable"},
	} {
		w := httptest.NewRecorder()
		s.syncSSRun(w, httptest.NewRequest("POST", "/api/sync/scopesentry/sync", strings.NewReader(tc.input)))
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.message) {
			t.Fatalf("%s: got %d %s", tc.input, w.Code, w.Body.String())
		}
	}
}
