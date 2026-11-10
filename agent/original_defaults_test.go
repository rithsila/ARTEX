package agent

import (
	"os"
	"strings"
	"testing"

	"github.com/Autumn-27/norma/harness"
)

func originalFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/original_defaults/" + name + ".txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func originalRoles() map[string]string {
	return map[string]string{"goals": goalsDefaultTmpl, "planner": plannerDefaultTmpl, "worker": workerDefaultTmpl, "mainagent": mainAgentDefaultTmpl, "auto": autoDefaultTmpl, "pentest": pentestDefaultTmpl, "reporter": ReporterDefaultPrompt, "retester": RetesterDefaultPrompt}
}

func TestOriginalConstantBodies(t *testing.T) {
	bodies := map[string]string{"goalsDefaultTmpl": goalsDefaultTmpl, "goalsScopeTail": goalsScopeTail, "plannerDefaultTmpl": plannerDefaultTmpl, "workerDefaultTmpl": workerDefaultTmpl, "mainAgentDefaultTmpl": mainAgentDefaultTmpl, "autoDefaultTmpl": autoDefaultTmpl, "pentestDefaultTmpl": pentestDefaultTmpl, "ReporterDefaultPrompt": ReporterDefaultPrompt, "RetesterDefaultPrompt": RetesterDefaultPrompt, "settleWrapUpPrompt": settleWrapUpPrompt, "plannerWrapUpDefault": plannerWrapUpDefault, "mainAgentWrapUpDefault": mainAgentWrapUpDefault, "genericWrapUpDefault": genericWrapUpDefault, "workerTaskTimeoutDefault": workerTaskTimeoutDefault, "plannerTaskTimeoutDefault": plannerTaskTimeoutDefault}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			if body != originalFixture(t, name) {
				t.Errorf("%s differs from approved historical body", name)
			}
		})
	}
}

func requireEnglishDefault(t *testing.T, got, body string) {
	t.Helper()
	if !strings.HasPrefix(got, body) {
		t.Error("original rules not preserved")
	}
	suffix := strings.TrimPrefix(got, body)
	for _, want := range []string{"English", "supersedes", "Chinese", "Tool names", "code", "quoted evidence"} {
		if !strings.Contains(suffix, want) {
			t.Errorf("default language directive missing %q", want)
		}
	}
}

func TestOriginalDefaultSeedsAndRendering(t *testing.T) {
	old := PromptOverride
	t.Cleanup(func() { PromptOverride = old })
	vars := PlannerVars{Goal: "test-goal", DataDir: "/data", Now: "now"}
	seeds := BuiltinPromptSeeds()
	for key, body := range originalRoles() {
		t.Run(key, func(t *testing.T) {
			requireEnglishDefault(t, seeds[key], body)
			rendered, err := renderTmpl(body, vars)
			if err != nil {
				t.Fatal(err)
			}
			PromptOverride = nil
			requireEnglishDefault(t, renderSystem(key, body, vars), rendered)
			for _, bad := range []string{"{{.Invalid}}", "{{"} {
				PromptOverride = func(string) (string, bool) { return bad, true }
				requireEnglishDefault(t, renderSystem(key, body, vars), rendered)
			}
			PromptOverride = func(string) (string, bool) { return seeds[key], true }
			requireEnglishDefault(t, renderSystem(key, body, vars), rendered)
			custom := "  你是自定义助手，中文回答。渗透 {{.Goal}}\n"
			PromptOverride = func(string) (string, bool) { return custom, true }
			if got := renderSystem(key, body, vars); got != "  你是自定义助手，中文回答。渗透 test-goal\n" {
				t.Errorf("custom override changed: %q", got)
			}
		})
	}
	PromptOverride = nil
	if got := renderSystem("custom", DefaultAssistantPrompt, vars); got != DefaultAssistantPrompt {
		t.Fatal("custom fallback changed")
	}
	if got := renderSystem("planner", "DEFAULT", vars); got != "DEFAULT" {
		t.Fatal("non-default body modified")
	}
	if got := renderSystem("planner", plannerDefaultTmpl, struct{}{}); !strings.HasPrefix(got, plannerDefaultTmpl) || !strings.Contains(got, "supersedes") {
		t.Fatal("raw fallback lost default language")
	}
}

func TestOriginalWrapupDefaultsAndOverrides(t *testing.T) {
	old, oldTT := WrapupOverride, WrapupTaskTimeoutOverride
	t.Cleanup(func() { WrapupOverride = old; WrapupTaskTimeoutOverride = oldTT })
	WrapupOverride = nil
	WrapupTaskTimeoutOverride = nil
	for key, body := range map[string]string{"worker": settleWrapUpPrompt, "planner": plannerWrapUpDefault, "mainagent": mainAgentWrapUpDefault} {
		requireEnglishDefault(t, WrapupDefault(key), body)
		requireEnglishDefault(t, resolveWrapup(key), body)
	}
	if WrapupDefault("custom") != "You are about to be terminated due to budget exhaustion. Please write back any uncommitted results first, then summarize what you did and key conclusions in a single line of plain text (this sentence will be displayed as the result of this run)." {
		t.Fatal("custom generic wrapup changed")
	}
	for key, body := range map[string]string{"worker": workerTaskTimeoutDefault, "planner": plannerTaskTimeoutDefault} {
		requireEnglishDefault(t, TaskTimeoutWrapupDefault(key), body)
		st := wrapupSettlementForTask(key, []string{"exec"}, true)
		requireEnglishDefault(t, st.PromptByReason[harness.ReasonTimeout], body)
		if st.PromptByReason[harness.ReasonMaxTurns] != WrapupDefault(key) {
			t.Fatal("turn exhaustion incorrectly ends whole task")
		}
		if wrapupSettlementForTask(key, nil, false).PromptByReason != nil {
			t.Fatal("unclamped task received timeout prompt")
		}
	}
	for _, key := range []string{"custom", "mainagent", "chat"} {
		if TaskTimeoutWrapupDefault(key) != "" {
			t.Fatal("unsupported timeout default is not empty")
		}
	}
	custom := "  你是收尾助手，渗透结果用中文。\n"
	WrapupOverride = func(string) (string, bool) { return custom, true }
	WrapupTaskTimeoutOverride = WrapupOverride
	for _, key := range []string{"worker", "planner", "mainagent", "custom"} {
		if resolveWrapup(key) != custom || resolveTaskTimeoutWrapup(key) != custom {
			t.Fatal("custom wrapup changed")
		}
	}
	WrapupOverride = func(string) (string, bool) { return " \n", true }
	WrapupTaskTimeoutOverride = WrapupOverride
	requireEnglishDefault(t, resolveWrapup("planner"), plannerWrapUpDefault)
	requireEnglishDefault(t, resolveTaskTimeoutWrapup("planner"), plannerTaskTimeoutDefault)
}

func TestOriginalChatFallbacks(t *testing.T) {
	old := PromptOverride
	t.Cleanup(func() { PromptOverride = old })
	for key, body := range originalRoles() {
		t.Run(key, func(t *testing.T) {
			want, err := renderTmpl(body, chatVars{Goal: "(none — interactive chat session, no task attached)", DataDir: "/data"})
			if err != nil {
				t.Fatal(err)
			}
			for _, override := range []string{"", "{{.Invalid}}", "{{"} {
				PromptOverride = func(string) (string, bool) { return override, override != "" }
				got := strings.TrimSuffix(chatSystem(key, "/data", "/work"), chatWorkDirSpec("/work"))
				requireEnglishDefault(t, got, want)
			}
			custom := "  你是自定义助手，中文回答。\n"
			PromptOverride = func(string) (string, bool) { return custom, true }
			if got := chatSystem(key, "/data", "/work"); got != custom+chatWorkDirSpec("/work") {
				t.Fatalf("custom chat changed: %q", got)
			}
		})
	}
	for _, override := range []string{"", "{{.Invalid}}"} {
		PromptOverride = func(string) (string, bool) { return override, override != "" }
		if got := chatSystem("custom", "/data", "/work"); got != DefaultAssistantPrompt+chatWorkDirSpec("/work") {
			t.Fatal("custom fallback changed")
		}
	}
}
