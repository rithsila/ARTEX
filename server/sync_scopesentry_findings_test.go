package server

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
	actool "github.com/Autumn-27/norma/tool"
)

type ssFakeCatalog struct {
	rows []*db.Tool
	err  error
}

func (f *ssFakeCatalog) ListTools() ([]*db.Tool, error)        { return f.rows, f.err }
func (*ssFakeCatalog) GetAgentByKey(string) (*db.Agent, error) { return nil, nil }
func (*ssFakeCatalog) InsertToolUsage(*db.ToolUsage) error     { return nil }

func TestSSSyncResolverRequiresCatalogGrant(t *testing.T) {
	previous := agent.ToolResolve
	t.Cleanup(func() { agent.ToolResolve = previous })
	syncTool := (&Server{}).toolSyncScopeSentry()
	unmanaged := wrTool("unmanaged_host", "unchanged", nil, nil)
	domain := wrTool("existing_domain", "code default", nil, nil)
	for _, tc := range []struct {
		name string
		rows []*db.Tool
		err  error
		want []string
	}{
		{"catalog failure", nil, errors.New("catalog unavailable"), []string{"unmanaged_host", "existing_domain"}},
		{"partial rows with failure", []*db.Tool{{Key: syncTool.Name(), Enabled: true, Agents: []string{"auto"}}}, errors.New("incomplete query"), []string{"unmanaged_host", "existing_domain"}},
		{"missing row", nil, nil, []string{"unmanaged_host", "existing_domain"}},
		{"empty bindings", []*db.Tool{{Key: syncTool.Name(), Enabled: true, Agents: []string{}}}, nil, []string{"unmanaged_host", "existing_domain"}},
		{"different agent", []*db.Tool{{Key: syncTool.Name(), Enabled: true, Agents: []string{"worker"}}}, nil, []string{"unmanaged_host", "existing_domain"}},
		{"disabled", []*db.Tool{{Key: syncTool.Name(), Enabled: false, Agents: []string{"auto"}}}, nil, []string{"unmanaged_host", "existing_domain"}},
		{"explicit grant", []*db.Tool{{Key: syncTool.Name(), Enabled: true, Agents: []string{"auto"}, Description: "catalog copy"}}, nil, []string{syncTool.Name(), "unmanaged_host", "existing_domain"}},
		{"existing domain denied", []*db.Tool{{Key: "existing_domain", Enabled: true, Agents: []string{"worker"}}}, nil, []string{"unmanaged_host"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wireToolResolver(&ssFakeCatalog{rows: tc.rows, err: tc.err}, nil)
			input := []actool.CoreTool{syncTool, unmanaged, domain}
			got := agent.ToolResolve(context.Background(), "auto", input)
			var names []string
			for _, tool := range got {
				names = append(names, tool.Name())
			}
			if !reflect.DeepEqual(names, tc.want) {
				t.Fatalf("resolved %v, want %v", names, tc.want)
			}
			if input[0] != syncTool {
				t.Fatal("resolver changed incoming tool list")
			}
			if tc.name == "explicit grant" {
				if got[0].Description() != "catalog copy" {
					t.Fatal("catalog override lost")
				}
				if _, ok := got[0].(*meteredTool); !ok {
					t.Fatal("catalog metering lost")
				}
			}
		})
	}
}

func TestSSSyncResetDefaultsAreUnbound(t *testing.T) {
	s := &Server{}
	seed, ok := s.builtinToolSeed("sync_scopesentry_assets")
	if !ok {
		t.Fatal("reset cannot find ScopeSentry defaults")
	}
	bindings, err := json.Marshal(seed.Agents)
	if err != nil || string(bindings) != "[]" {
		t.Fatalf("reset must unbind, got %s (%v)", bindings, err)
	}
	tool := s.toolSyncScopeSentry()
	if seed.Desc != tool.Description() || !reflect.DeepEqual(seed.Schema, tool.InputSchema()) {
		t.Fatal("reset metadata differs from code defaults")
	}
	if _, ok := s.builtinToolSeed("list_assets"); !ok {
		t.Fatal("existing builtin reset lost")
	}
	if _, ok := s.builtinToolSeed("not_a_builtin"); ok {
		t.Fatal("unknown tool gained defaults")
	}
}

type ssCancelAssets struct {
	ssFakeAssets
	cancel context.CancelFunc
	after  string
	err    error
}

func (f *ssCancelAssets) UpsertSubdomain(req db.UpsertSubdomainReq) (int64, error) {
	id, err := f.ssFakeAssets.UpsertSubdomain(req)
	if f.after == "subdomain" {
		f.cancel()
		return id, f.err
	}
	return id, err
}
func (f *ssCancelAssets) UpsertIP(req db.UpsertIPReq) (int64, error) {
	id, err := f.ssFakeAssets.UpsertIP(req)
	if f.after == "ip" {
		f.cancel()
		return id, f.err
	}
	return id, err
}

func TestSSSyncStopsBetweenStoreWrites(t *testing.T) {
	for _, after := range []string{"subdomain", "ip"} {
		for _, storeErr := range []error{nil, context.Canceled} {
			label := "success"
			if storeErr != nil {
				label = "cancel_error"
			}
			t.Run(after+"/"+label, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				assets := &ssCancelAssets{cancel: cancel, after: after, err: storeErr}
				cl := &ssFakeClient{call: func(context.Context, string, map[string]any) (string, error) {
					return `{"list":[{"host":"a.example","ip":["192.0.2.1","192.0.2.2"]}]}`, nil
				}}
				resp, err := syncScopeSentry(ctx, ssSyncReq{Dimension: "task", Targets: []string{"scan"}, AssetTypes: []string{"subdomain"}}, func(context.Context) (ssSyncClient, error) { return cl, nil }, assets, &ssFakeCompanies{})
				wantIPs := 0
				if after == "ip" {
					wantIPs = 1
				}
				if !errors.Is(err, context.Canceled) || assets.subdomains != 1 || assets.ips != wantIPs || !cl.closed || len(resp.Errors) != 0 {
					t.Fatalf("cancelled after %s: err=%v subdomains=%d ips=%d errors=%v closed=%v", after, err, assets.subdomains, assets.ips, resp.Errors, cl.closed)
				}
			})
		}
	}
}

func TestSSSyncPropagatesStoreCancellation(t *testing.T) {
	for _, storeErr := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, after := range []string{"subdomain", "ip"} {
			t.Run(after+"/"+storeErr.Error(), func(t *testing.T) {
				// The store can return cancellation even when the parent is still active.
				assets := &ssCancelAssets{cancel: func() {}, after: after, err: storeErr}
				cl := &ssFakeClient{call: func(context.Context, string, map[string]any) (string, error) {
					return `{"list":[{"host":"a.example","ip":["192.0.2.1","192.0.2.2"]}]}`, nil
				}}
				resp, err := syncScopeSentry(context.Background(), ssSyncReq{Dimension: "task", Targets: []string{"scan"}, AssetTypes: []string{"subdomain"}}, func(context.Context) (ssSyncClient, error) { return cl, nil }, assets, &ssFakeCompanies{})
				wantIPs := 0
				if after == "ip" {
					wantIPs = 1
				}
				if !errors.Is(err, storeErr) || assets.ips != wantIPs || len(resp.Errors) != 0 || !cl.closed {
					t.Fatalf("store cancellation lost: err=%v ips=%d errors=%v closed=%v", err, assets.ips, resp.Errors, cl.closed)
				}
			})
		}
	}
}

func TestSSIngestChecksCancellationBeforeEveryWriteType(t *testing.T) {
	for _, tc := range []struct{ assetType, raw string }{
		{"subdomain", `{"host":"a.example","ip":["192.0.2.1"]}`},
		{"app", `{"name":"Example"}`},
		{"service", `{"url":"https://a.example","service":"http"}`},
		{"service", `{"ip":"192.0.2.1","port":"22","service":"ssh"}`},
	} {
		t.Run(tc.assetType+tc.raw, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			assets := &ssFakeAssets{}
			err := ssIngest(ctx, assets, tc.assetType, json.RawMessage(tc.raw), map[string]int{})
			if !errors.Is(err, context.Canceled) || *assets != (ssFakeAssets{}) {
				t.Fatalf("cancelled ingest wrote assets: %v %+v", err, assets)
			}
		})
	}
}
