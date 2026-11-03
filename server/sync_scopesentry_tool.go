package server

import (
	"context"
	"encoding/json"

	"github.com/Autumn-27/artex/agent"
	actool "github.com/Autumn-27/norma/tool"
)

func newSSSyncTool(run func(context.Context, ssSyncReq) (ssSyncResponse, error)) actool.CoreTool {
	return wrTool("sync_scopesentry_assets",
		"Import existing ScopeSentry assets into ARTEX (does not start a scan). Same import as Asset Sync: subdomains, services, apps and related IPs; at most 5000 per target/type. Project targets are ObjectIDs; task targets are task names, not IDs. Company creation applies only to projects. Check warnings/errors for partial imports. No vulnerability findings are imported.",
		objSchema(map[string]any{
			"dimension":      map[string]any{"type": "string", "enum": []string{"project", "task"}},
			"targets":        map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "string", "minLength": 1}, "description": "Project ObjectIDs or exact task names, according to dimension"},
			"asset_types":    map[string]any{"type": "array", "items": map[string]any{"type": "string", "enum": []string{"subdomain", "service", "app"}}, "description": "Empty or omitted imports all three types"},
			"create_company": map[string]any{"type": "boolean", "description": "Create/update company scopes from project roots (project dimension only)"},
			"page_size":      map[string]any{"type": "integer", "description": "MCP page size; 1-500, otherwise defaults to 100"},
		}, "dimension", "targets"),
		func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
			var req ssSyncReq
			if err := json.Unmarshal(in, &req); err != nil {
				return actool.Errorf("invalid JSON: " + err.Error()), nil
			}
			if err := validateSSSync(req); err != nil {
				return actool.Errorf(err.Error()), nil
			}
			resp, err := run(ctx, req)
			if err != nil {
				return actool.Errorf(err.Error()), nil
			}
			return jsonResult(resp)
		})
}

func (s *Server) toolSyncScopeSentry() actool.CoreTool { return newSSSyncTool(s.syncScopeSentry) }

// Host tools live in server, not the per-task agent ToolSet. Reuse ToolSeed
// metadata while keeping this catalog entry explicitly unbound for every agent.
func (s *Server) scopeSentryToolSeeds() []agent.ToolSeed {
	t := s.toolSyncScopeSentry()
	return []agent.ToolSeed{{Key: t.Name(), Desc: t.Description(), Schema: t.InputSchema(), Agents: []string{}}}
}
