package server

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Autumn-27/artex/db"
)

type ssFakeClient struct {
	call   func(context.Context, string, map[string]any) (string, error)
	closed bool
}

func (f *ssFakeClient) Call(ctx context.Context, name string, args any) (string, error) {
	return f.call(ctx, name, args.(map[string]any))
}
func (f *ssFakeClient) Close() error { f.closed = true; return nil }

type ssFakeAssets struct{ subdomains, ips, apps, http, other int }

func (f *ssFakeAssets) UpsertSubdomain(db.UpsertSubdomainReq) (int64, error) {
	f.subdomains++
	return 1, nil
}
func (f *ssFakeAssets) UpsertIP(db.UpsertIPReq) (int64, error)   { f.ips++; return 1, nil }
func (f *ssFakeAssets) UpsertApp(db.UpsertAppReq) (int64, error) { f.apps++; return 1, nil }
func (f *ssFakeAssets) UpsertHTTPService(db.UpsertHTTPServiceReq) (int64, error) {
	f.http++
	return 1, nil
}
func (f *ssFakeAssets) UpsertOtherService(db.UpsertOtherServiceReq) (int64, error) {
	f.other++
	return 1, nil
}

type ssFakeCompanies struct {
	names, roots []string
	recomputed   bool
}

func (f *ssFakeCompanies) UpsertByName(name string) (int64, error) {
	f.names = append(f.names, name)
	return 7, nil
}
func (f *ssFakeCompanies) AddScope(_ int64, roots []string, _ string) (int, int, int, []string) {
	f.roots = roots
	return len(roots), 0, 0, nil
}
func (f *ssFakeCompanies) RecomputeAttribution() error { f.recomputed = true; return nil }

func TestSSSyncRejectsSelectionBeforeIO(t *testing.T) {
	for _, req := range []ssSyncReq{
		{}, {Dimension: "task"}, {Dimension: "task", Targets: []string{" "}},
		{Dimension: "project", Targets: []string{"not-an-object-id"}},
	} {
		_, err := syncScopeSentry(context.Background(), req, func(context.Context) (ssSyncClient, error) { t.Fatal("invalid selection reached I/O"); return nil, nil }, nil, nil)
		var input *ssSyncError
		if !errors.As(err, &input) || input.Status != 400 {
			t.Fatalf("%+v: expected input error, got %v", req, err)
		}
	}
}

func TestSSSyncFiltersAndImports(t *testing.T) {
	for _, dimension := range []string{"project", "task"} {
		t.Run(dimension, func(t *testing.T) {
			target := "scan-name"
			if dimension == "project" {
				target = "507f1f77bcf86cd799439011"
			}
			assets, companies := &ssFakeAssets{}, &ssFakeCompanies{}
			cl := &ssFakeClient{call: func(ctx context.Context, name string, args map[string]any) (string, error) {
				if name == "get_project" {
					return `{"name":"Example","target":"example.com\n foo.com\n"}`, nil
				}
				if name != "list_assets" || !reflect.DeepEqual(args["filter"], map[string]any{dimension: []string{target}}) || args["pageSize"] != 100 {
					t.Fatalf("wrong request: %s %#v", name, args)
				}
				switch args["asset_type"] {
				case "subdomain":
					return `{"list":[{"host":"a.example.com","type":"A","ip":["192.0.2.1"]}]}`, nil
				case "asset":
					return `{"list":[{"url":"https://a.example.com","service":"http"},{"ip":"192.0.2.1","port":"22","service":"ssh"}]}`, nil
				case "app":
					return `{"list":[{"name":"Example app"}]}`, nil
				}
				t.Fatal("wrong type")
				return "", nil
			}}
			got, err := syncScopeSentry(context.Background(), ssSyncReq{Dimension: dimension, Targets: []string{target}, CreateCompany: true, PageSize: 501}, func(context.Context) (ssSyncClient, error) { return cl, nil }, assets, companies)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Synced, map[string]int{"subdomain": 1, "ip": 1, "service": 2, "app": 1}) || assets.http != 1 || assets.other != 1 || !cl.closed {
				t.Fatalf("bad import: %+v %+v", got, assets)
			}
			if dimension == "project" && (!companies.recomputed || !reflect.DeepEqual(companies.roots, []string{"example.com", "foo.com"}) || !reflect.DeepEqual(got.Companies, []string{"Example"})) {
				t.Fatalf("company behavior lost: %+v", companies)
			}
			if dimension == "task" && companies.recomputed {
				t.Fatal("task must not create a company")
			}
			data, _ := json.Marshal(got)
			if !strings.Contains(string(data), `"errors":null`) || !strings.Contains(string(data), `"warnings":null`) {
				t.Fatalf("response changed: %s", data)
			}
		})
	}
}

func TestSSSyncCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := syncScopeSentry(ctx, ssSyncReq{Dimension: "task", Targets: []string{"scan"}}, func(context.Context) (ssSyncClient, error) { t.Fatal("cancelled operation connected"); return nil, nil }, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	assets := &ssFakeAssets{}
	cl := &ssFakeClient{call: func(ctx context.Context, _ string, _ map[string]any) (string, error) {
		cancel()
		return `{"list":[{"host":"do-not-import.example"}]}`, nil
	}}
	_, err = syncScopeSentry(ctx, ssSyncReq{Dimension: "task", Targets: []string{"scan"}, AssetTypes: []string{"subdomain"}}, func(context.Context) (ssSyncClient, error) { return cl, nil }, assets, &ssFakeCompanies{})
	if !errors.Is(err, context.Canceled) || assets.subdomains != 0 || !cl.closed {
		t.Fatalf("cancel lost: %v %+v", err, assets)
	}
}

func TestSSSyncConnectionCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, err := syncScopeSentry(ctx, ssSyncReq{Dimension: "task", Targets: []string{"scan"}}, func(context.Context) (ssSyncClient, error) {
		cancel()
		return nil, context.Canceled
	}, &ssFakeAssets{}, &ssFakeCompanies{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("connection lost cancellation: %v", err)
	}
}

func TestSSSyncCapAndPartialErrors(t *testing.T) {
	assets := &ssFakeAssets{}
	calls := 0
	cl := &ssFakeClient{call: func(_ context.Context, _ string, args map[string]any) (string, error) {
		if args["asset_type"] == "app" {
			return "", errors.New("upstream unavailable")
		}
		calls++
		return `{"list":[` + strings.TrimSuffix(strings.Repeat(`{"host":"a.example"},`, 500), ",") + `]}`, nil
	}}
	got, err := syncScopeSentry(context.Background(), ssSyncReq{Dimension: "task", Targets: []string{"scan"}, AssetTypes: []string{"subdomain", "app", "unknown"}, PageSize: 500}, func(context.Context) (ssSyncClient, error) { return cl, nil }, assets, &ssFakeCompanies{})
	if err != nil || assets.subdomains != 5000 || calls != 10 || len(got.Warnings) != 2 || len(got.Errors) != 1 || !strings.Contains(got.Errors[0], "upstream unavailable") {
		t.Fatalf("cap/errors changed: %+v %v calls=%d", got, err, calls)
	}
}
