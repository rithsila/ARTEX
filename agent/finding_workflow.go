package agent

import (
	"encoding/json"
	"fmt"

	"github.com/Autumn-27/artex/db"
	actool "github.com/Autumn-27/norma/tool"
)

const findingIDGuidance = "\n\n**Finding ID Convention**: finding_id is the standalone finding record ID; finding_node_id is the exploration node ID. The id in list_findings / list_task_findings / node_detail / get_task_node_detail is reserved for exploration node ID; read the standalone ID from finding_id in the same response. get_finding_traffic / bind_finding_traffic use standalone finding_id. The legacy update_finding_report finding_id parameter still takes finding_node_id. Do not use the number on the first line of report_finding for evidence tools, and do not guess other numbers when encountering ID errors."

// The server supplies the persisted setting. A missing setting/host is off.
// Consulted at assembly and again on writes so an already-running session
// cannot keep binding after the user switches the feature off.
var FindingTrafficBindingEnabled func() bool

func findingTrafficBindingEnabled() bool {
	return FindingTrafficBindingEnabled != nil && FindingTrafficBindingEnabled()
}

// Applied after ToolResolve: user descriptions and prompts remain intact, while
// all actual reporters (including Planner and custom chat agents) see the same
// API contract. Disabled/unbound tools are never reintroduced here.
func findingWorkflowTools(agentKey string, tools []actool.CoreTool) ([]actool.CoreTool, string) {
	if !findingTrafficBindingEnabled() {
		out := make([]actool.CoreTool, 0, len(tools))
		for _, tool := range tools {
			if tool.Name() == "bind_finding_traffic" {
				continue
			}
			if agentKey == "reporter" && (tool.Name() == "traffic_search" || tool.Name() == "traffic_get" || tool.Name() == "traffic_blob") {
				continue
			}
			switch tool.Name() {
			case "report_finding", "add_hint", "add_task_hint":
				// Work on a copy: toggling back on must restore the original schema.
				raw, _ := json.Marshal(tool.InputSchema())
				var schema map[string]any
				if json.Unmarshal(raw, &schema) == nil {
					stripTrafficParameters(schema)
					tool = DecorateTool(tool, tool.Description(), schema)
				}
			}
			out = append(out, tool)
		}
		return out, ""
	}
	out := append([]actool.CoreTool(nil), tools...)
	has := map[string]bool{}
	for i, tool := range out {
		has[tool.Name()] = true
		note := ""
		switch tool.Name() {
		case "report_finding":
			note = "\nBy default, the reporter agent verifies and binds traffic before writing the report. The submitter should retain verification commands, key output, existing real traffic IDs, and their purpose in evidence, for the reporter agent to verify against execution records; no extra packet captures needed for binding. Explicit immediate binding is supported: traffic_refs or evidence_hint_id can submit verified references, where the latter reads structured references from the specified hint of this task; if either is invalid, the entire submission fails. TCP or packetless findings do not require these optional parameters. Returns finding_id and finding_node_id representing standalone record and exploration node respectively."
		case "add_hint", "add_task_hint":
			note = "\nWhen handing off confirmed findings, retain the verified traffic IDs, roles, notes, and order in traffic_refs of the corresponding hint (top level for single hint, corresponding hints array for batch), and explain the specific vulnerability it proves in text. Callers must not hand off only text while dropping existing traffic references. Unverified candidates cannot be passed as evidence."
		case "get_finding_traffic", "bind_finding_traffic", "list_findings", "list_task_findings", "node_detail", "get_task_node_detail", "update_finding_report":
			note = findingIDGuidance
		}
		if note != "" {
			out[i] = DecorateTool(tool, tool.Description()+note, tool.InputSchema())
		}
	}
	guidance := ""
	if has["report_finding"] || has["add_task_hint"] || has["add_hint"] {
		guidance = "\n\n**Traffic Evidence Handoff (Optional)**: Automatic binding is performed by the reporter agent after finding storage and before report writing. Submitters should retain verification commands, key outputs, existing real traffic IDs and their purposes in evidence, with intent_id in task context, to facilitate tracing by the reporter agent; no need to inspect extra packets solely for binding. When Auto / Planner reports on behalf of an executor, do not discard existing references from the executor. add_hint / add_task_hint can use traffic_refs for handoff; explicit immediate binding remains compatible with traffic_refs / evidence_hint_id in report_finding. For TCP or packetless findings, record normally without guessing IDs or repeating scans solely to capture packets."
		if has["add_task_hint"] && !has["add_hint"] {
			guidance += "\nWhen platform chat lacks task context, do not call report_finding directly; hand off to existing tasks via add_task_hint, to be recorded by task agents, and check results with list_task_findings."
		}
		if has["prove_goal"] || has["goal_met"] {
			guidance += "\nBefore marking a goal met, complete reporting/handoff of all existing evidence. Do not end tasks or cancel workers solely because text findings are recorded when evidence handoff is not yet complete; waiting or forced packet capture is not required for packetless findings."
		}
	}
	if has["update_finding_report"] && has["bind_finding_traffic"] && has["get_finding_traffic"] {
		guidance += "\n\n**Automatic Traffic Association Before Reporting (Enabled)**: You are responsible for verifying and binding traffic for the vulnerability triggered this run before writing the report. First obtain explicit finding_id and finding_node_id from report_finding JSON return or get_task_node_detail / list_task_findings. Read finding details, execution records for the corresponding intent, and existing evidence list, prioritizing real IDs handed off by the submitter. If verification was HTTP and traffic tools are available, use traffic_search to filter candidates, then traffic_get to verify request/response actually supports the vulnerability; domain and time are for filtering only and do not prove attribution. Associate confirmed evidence in reproduction order using bind_finding_traffic(finding_id, traffic_refs), selecting baseline / proof / verification / supporting and explaining purpose. Only operate on this vulnerability, without creating duplicate findings or re-probing targets. After successful binding, call get_finding_traffic again to get the latest version, read required content, and pass the version actually read as evidence_version to update_finding_report (whose finding_id parameter still takes finding_node_id). Do not append existing bindings repeatedly. When TCP, uncollected, tools unavailable, or no exact match, skip automatic binding and write report normally based on text/command evidence explaining why, without guessing to collect traffic. Do not claim success if binding fails; retain existing evidence and explain unassociated reason in report."
	}
	if guidance != "" || has["get_finding_traffic"] || has["update_finding_report"] {
		guidance += findingIDGuidance
	}
	return out, guidance
}

func stripTrafficParameters(schema map[string]any) {
	props, _ := schema["properties"].(map[string]any)
	delete(props, "traffic_refs")
	delete(props, "evidence_hint_id")
	if required, ok := schema["required"].([]any); ok {
		kept := required[:0]
		for _, key := range required {
			if key != "traffic_refs" && key != "evidence_hint_id" {
				kept = append(kept, key)
			}
		}
		schema["required"] = kept
	}
	if hints, ok := props["hints"].(map[string]any); ok {
		if items, ok := hints["items"].(map[string]any); ok {
			stripTrafficParameters(items)
		}
	}
}

// HintTrafficSchema is shared by the task-local and cross-task hint tools.
func HintTrafficSchema() map[string]any {
	return map[string]any{"type": "array", "description": "Optional: Verified traffic references corresponding to specific vulnerabilities in this hint, preserving order; after handoff, report_finding can pass evidence_hint_id to carry these references.", "items": obj(map[string]any{"traffic_id": str("Real traffic ID"), "role": str("baseline / proof / verification / supporting"), "note": str("What conclusion this traffic supports")}, "traffic_id")}
}

func (t *ToolSet) findingRefsFromHint(hintID int64, explicit []db.TrafficRef) ([]db.TrafficRef, error) {
	if hintID <= 0 {
		return db.NormalizeTrafficRefs(explicit)
	}
	n, err := t.ts.GetNode(hintID) // local store only: inherited hints cannot supply evidence
	if err != nil {
		return nil, err
	}
	if n == nil || n.Kind != db.KindHint {
		return nil, fmt.Errorf("evidence_hint_id=%d must be a hint node of this task (inherited hints cannot be used directly for binding)", hintID)
	}
	var payload struct {
		Refs []db.TrafficRef `json:"traffic_refs"`
	}
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return nil, err
	}
	return db.NormalizeTrafficRefs(append(append([]db.TrafficRef{}, explicit...), payload.Refs...))
}
