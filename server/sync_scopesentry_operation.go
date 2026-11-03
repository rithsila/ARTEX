package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Autumn-27/artex/db"
)

const (
	syncMaxPerType  = 5000 // Per target/type import protection cap.
	syncDefaultPage = 100
)

type ssSyncReq struct {
	Dimension     string   `json:"dimension"` // project ObjectIDs or task names
	Targets       []string `json:"targets"`
	AssetTypes    []string `json:"asset_types"` // subdomain | app | service
	CreateCompany bool     `json:"create_company"`
	PageSize      int      `json:"page_size"`
}

type ssSyncResponse struct {
	Synced    map[string]int `json:"synced"`
	Companies []string       `json:"companies"`
	Warnings  []string       `json:"warnings"`
	Errors    []string       `json:"errors"`
}

type ssSyncError struct {
	Status  int
	Message string
}

func (e *ssSyncError) Error() string { return e.Message }

// Narrow boundaries let contract tests exercise the real import without opening a DB.
type ssSyncClient interface {
	Call(context.Context, string, any) (string, error)
	Close() error
}
type ssAssetStore interface {
	UpsertSubdomain(db.UpsertSubdomainReq) (int64, error)
	UpsertIP(db.UpsertIPReq) (int64, error)
	UpsertApp(db.UpsertAppReq) (int64, error)
	UpsertHTTPService(db.UpsertHTTPServiceReq) (int64, error)
	UpsertOtherService(db.UpsertOtherServiceReq) (int64, error)
}
type ssCompanyStore interface {
	UpsertByName(string) (int64, error)
	AddScope(int64, []string, string) (int, int, int, []string)
	RecomputeAttribution() error
}

func validateSSSync(req ssSyncReq) error {
	bad := func(message string) error { return &ssSyncError{Status: 400, Message: message} }
	if req.Dimension != "project" && req.Dimension != "task" {
		return bad("dimension must be project or task")
	}
	if len(req.Targets) == 0 {
		return bad("targets cannot be empty")
	}
	for _, target := range req.Targets {
		if strings.TrimSpace(target) == "" {
			return bad("targets cannot contain empty values")
		}
		if req.Dimension == "project" {
			if _, err := hex.DecodeString(target); len(target) != 24 || err != nil {
				return bad("project targets must be ObjectIDs")
			}
		}
	}
	return nil
}

// syncScopeSentry is the single import path for HTTP sync and the opt-in host tool.
// It preserves partial fetch/ingest errors in the response. Cancellation stops new
// work, but cannot roll back rows already imported (store methods are synchronous).
func syncScopeSentry(ctx context.Context, req ssSyncReq, connect func(context.Context) (ssSyncClient, error), as ssAssetStore, cs ssCompanyStore) (ssSyncResponse, error) {
	if err := validateSSSync(req); err != nil {
		return ssSyncResponse{}, err
	}
	if err := ctx.Err(); err != nil {
		return ssSyncResponse{}, err
	}
	if as == nil || cs == nil {
		return ssSyncResponse{}, &ssSyncError{Status: 503, Message: "database unavailable"}
	}
	if len(req.AssetTypes) == 0 {
		req.AssetTypes = []string{"subdomain", "service", "app"}
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 500 {
		pageSize = syncDefaultPage
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	cl, err := connect(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return ssSyncResponse{}, ctx.Err()
		}
		return ssSyncResponse{}, &ssSyncError{Status: 502, Message: err.Error()}
	}
	defer cl.Close()
	resp := ssSyncResponse{Synced: map[string]int{"subdomain": 0, "app": 0, "service": 0, "ip": 0}}
	madeCompany := false
	for _, target := range req.Targets {
		if err := ctx.Err(); err != nil {
			return resp, err
		}
		filter := map[string]any{}
		switch req.Dimension {
		case "project":
			filter["project"] = []string{target}
			if req.CreateCompany {
				name, roots, perr := ssProjectMeta(ctx, cl, target)
				if err := ctx.Err(); err != nil {
					return resp, err
				}
				if perr != nil {
					resp.Warnings = append(resp.Warnings, fmt.Sprintf("failed to get project %s details: %v", target, perr))
				} else if name != "" {
					cid, cerr := cs.UpsertByName(name)
					if err := ctx.Err(); err != nil {
						return resp, err
					}
					if cerr != nil {
						resp.Warnings = append(resp.Warnings, fmt.Sprintf("failed to create company %s: %v", name, cerr))
					} else {
						resp.Companies = append(resp.Companies, name)
						madeCompany = true
						if err := ctx.Err(); err != nil {
							return resp, err
						}
						if len(roots) > 0 {
							cs.AddScope(cid, roots, "scopesentry:"+name)
						}
					}
				}
			}
		case "task":
			filter["task"] = []string{target}
		}
		for _, at := range req.AssetTypes {
			if err := ctx.Err(); err != nil {
				return resp, err
			}
			ssType, ok := map[string]string{"subdomain": "subdomain", "service": "asset", "app": "app"}[at]
			if !ok {
				resp.Warnings = append(resp.Warnings, "unknown asset type, skipped: "+at)
				continue
			}
			items, truncated, ferr := ssPageAll(ctx, cl, ssType, filter, pageSize)
			if err := ctx.Err(); err != nil {
				return resp, err
			}
			if ferr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("failed to fetch %s(%s): %v", at, target, ferr))
				continue
			}
			if truncated {
				resp.Warnings = append(resp.Warnings, fmt.Sprintf("%s(%s) reached %d limit, truncated", at, target, syncMaxPerType))
			}
			for _, raw := range items {
				if err := ctx.Err(); err != nil {
					return resp, err
				}
				if e := ssIngest(ctx, as, at, raw, resp.Synced); e != nil {
					if errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) {
						return resp, e
					}
					resp.Errors = append(resp.Errors, e.Error())
				}
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return resp, err
	}
	if madeCompany {
		_ = cs.RecomputeAttribution()
	}
	return resp, ctx.Err()
}

// ssProjectMeta reads newline-separated roots for company scope creation.
func ssProjectMeta(ctx context.Context, cl ssSyncClient, projectID string) (name string, roots []string, err error) {
	text, err := cl.Call(ctx, "get_project", map[string]any{"id": projectID})
	if err != nil {
		return "", nil, err
	}
	var p struct {
		Name   string `json:"name"`
		Target string `json:"target"`
	}
	if err := json.Unmarshal([]byte(text), &p); err != nil {
		return "", nil, err
	}
	for _, line := range strings.Split(p.Target, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			roots = append(roots, t)
		}
	}
	return p.Name, roots, nil
}

func ssPageAll(ctx context.Context, cl ssSyncClient, ssType string, filter map[string]any, pageSize int) (items []json.RawMessage, truncated bool, err error) {
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return items, truncated, err
		}
		args := map[string]any{"asset_type": ssType, "pageIndex": page, "pageSize": pageSize}
		if len(filter) > 0 {
			args["filter"] = filter
		}
		text, cerr := cl.Call(ctx, "list_assets", args)
		if cerr != nil {
			return items, truncated, cerr
		}
		var env struct {
			List []json.RawMessage `json:"list"`
		}
		if uerr := json.Unmarshal([]byte(text), &env); uerr != nil {
			return items, truncated, uerr
		}
		if len(env.List) == 0 {
			break
		}
		items = append(items, env.List...)
		if len(items) >= syncMaxPerType {
			items = items[:syncMaxPerType]
			truncated = true
			break
		}
		if len(env.List) < pageSize {
			break
		}
	}
	return items, truncated, nil
}
