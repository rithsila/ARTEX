// Package scopesentry checks integration artifacts without importing server/agent/db
// test runners. It never opens config, a database, or a network connection.
package scopesentry

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Autumn-27/norma/skill"
)

func root(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSkillLinksExistingMCP(t *testing.T) {
	reg, err := skill.LoadDir(filepath.Join(root(t), "skills"))
	if err != nil {
		t.Fatal(err)
	}
	s, ok := reg.Get("scopesentry")
	if !ok || !reflect.DeepEqual(s.MCPs, []string{"ScopeSentry"}) {
		t.Fatalf("missing exact MCP linkage: %+v", s)
	}
}

// Static documentation lint only: this does not execute CLI commands or an agent.
func TestDocumentedCLIContract(t *testing.T) {
	for _, path := range []string{"skills/scopesentry/SKILL.md", "SCOPESENTRY_GUIDE.md"} {
		doc := read(t, filepath.Join(root(t), path))
		for _, bad := range []string{"--key ", "--args ", "all 17", "All 17", `--target "example.com\n`} {
			if strings.Contains(doc, bad) {
				t.Errorf("%s advertises unsupported contract %q", path, bad)
			}
		}
		for _, want := range []string{"--api-key", "--input", "task stop", "task resume", "stop_scan_task", "resume_scan_task", "19"} {
			if !strings.Contains(doc, want) {
				t.Errorf("%s missing %q", path, want)
			}
		}
	}
}

func TestStagingValidationAndArchitecture(t *testing.T) {
	script := filepath.Join(root(t), "scripts/build-scopesentry-cli.sh")
	for _, args := range [][]string{{"--arch", "386"}, {"--os", "plan9"}, {"--arch"}, {"--source", "/does-not-exist"}, {"--unknown"}} {
		cmd := exec.Command("bash", append([]string{script}, args...)...)
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("invalid args accepted: %v %s", args, out)
		}
	}
	source := t.TempDir()
	bin := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "cmd/scopesentry"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module github.com/Autumn-27/ScopeSentry\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// Fake only the compiler: run the real argument/path/environment handling.
	goFake := `#!/bin/sh
if [ "$1" = env ]; then
  case "$2" in GOHOSTOS) echo darwin;; GOHOSTARCH) echo arm64;; GOOS) echo linux;; GOARCH) echo amd64;; esac
  exit 0
fi
[ "$CGO_ENABLED:$GOOS:$GOARCH:$GOPROXY:$GOSUMDB:$GOTOOLCHAIN" = "0:$EXPECT_OS:$EXPECT_ARCH:off:off:local" ] || exit 41
[ "$1" = build ] || exit 42
shift
while [ "$#" -gt 0 ]; do
  if [ "$1" = -o ]; then shift; printf staged > "$1"; fi
  shift
done
`
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(goFake), 0700); err != nil {
		t.Fatal(err)
	}
	for _, target := range []struct{ os, arch string }{{"linux", "amd64"}, {"linux", "arm64"}, {"native", "native"}} {
		outdir := t.TempDir()
		wantOS, wantArch := target.os, target.arch
		if wantOS == "native" {
			wantOS = "darwin"
		}
		if wantArch == "native" {
			wantArch = "arm64"
		}
		cmd := exec.Command("bash", script, "--source", source, "--os", target.os, "--arch", target.arch, "--output-dir", outdir)
		cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"), "EXPECT_OS="+wantOS, "EXPECT_ARCH="+wantArch)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("stage %v: %s %v", target, out, err)
		}
		path := filepath.Join(outdir, "scopesentry-"+wantOS+"-"+wantArch)
		if read(t, path) != "staged" {
			t.Fatalf("wrong artifact: %s", path)
		}
		info, err := os.Stat(path)
		if err != nil || info.Mode()&0111 == 0 {
			t.Fatal("artifact not executable")
		}
	}
}

func TestOptionalDockerCLIContract(t *testing.T) {
	docker := read(t, filepath.Join(root(t), "Dockerfile"))
	if !strings.Contains(docker, "COPY tools/bin/ /tmp/scopesentry-cli/") || !strings.Contains(docker, "ARG INSTALL_SCOPESENTRY_CLI=false") {
		t.Fatal("Docker must copy an existing staging directory and default to MCP-only")
	}
	// Extract and execute the real RUN block, with paths redirected to a temp root.
	start := strings.Index(docker, "RUN case \"${TARGETARCH}\"")
	if start < 0 {
		t.Fatal("missing architecture-checked optional install")
	}
	block := strings.Split(docker[start:], "\n# end optional ScopeSentry CLI")[0]
	block = strings.TrimPrefix(block, "RUN ")
	for _, tc := range []struct {
		arch, install     string
		present, want, ok bool
	}{{"amd64", "false", false, false, true}, {"arm64", "true", false, false, true}, {"arm64", "true", true, true, true}, {"amd64", "true", true, true, true}, {"386", "true", true, false, false}} {
		dir := t.TempDir()
		stage := filepath.Join(dir, "stage")
		dest := filepath.Join(dir, "scopesentry")
		if err := os.Mkdir(stage, 0700); err != nil {
			t.Fatal(err)
		}
		if tc.present {
			if err := os.WriteFile(filepath.Join(stage, "scopesentry-linux-"+tc.arch), []byte("fake-cli"), 0700); err != nil {
				t.Fatal(err)
			}
		}
		command := strings.ReplaceAll(block, "/tmp/scopesentry-cli", stage)
		command = strings.ReplaceAll(command, "/usr/local/bin/scopesentry", dest)
		cmd := exec.Command("sh", "-ec", command)
		cmd.Env = append(os.Environ(), "TARGETARCH="+tc.arch, "INSTALL_SCOPESENTRY_CLI="+tc.install)
		out, err := cmd.CombinedOutput()
		if (err == nil) != tc.ok {
			t.Fatalf("%+v: %s %v", tc, out, err)
		}
		_, err = os.Stat(dest)
		if (err == nil) != tc.want {
			t.Fatalf("%+v: wrong installation result", tc)
		}
	}
}

type workflowStep struct {
	Tool   string         `json:"tool"`
	Input  map[string]any `json:"input"`
	Result map[string]any `json:"result"`
}
type workflowFixture struct {
	Steps       []workflowStep `json:"steps"`
	FailedReads []struct {
		Error        string   `json:"error"`
		Complete     bool     `json:"complete"`
		AllowedCalls []string `json:"allowed_calls"`
	} `json:"failed_reads"`
}

// Checked illustrative payloads, NOT runtime orchestration or LLM enforcement.
func TestCheckedWorkflowPreservesFullScopeAndSeparatesSync(t *testing.T) {
	var f workflowFixture
	if err := json.Unmarshal([]byte(read(t, "testdata/workflow.json")), &f); err != nil {
		t.Fatal(err)
	}
	wantOrder := []string{"list_projects_data", "get_project", "create_project", "list_nodes", "list_scan_templates", "get_scan_template", "create_scan_task", "get_task", "list_assets", "sync_scopesentry_assets"}
	var order []string
	for _, s := range f.Steps {
		order = append(order, s.Tool)
	}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("workflow order: %v", order)
	}
	current := f.Steps[1].Result
	if current["complete"] != true {
		t.Fatal("must read COMPLETE project scope before replacement")
	}
	if current["target"] != "example.com\n192.0.2.0/24" || f.Steps[2].Input["target"] != "example.com\n192.0.2.0/24\nthird.example" {
		t.Fatal("two old targets must survive adding the third target")
	}
	if f.Steps[2].Input["name"] != current["name"] {
		t.Fatal("update must use selected existing name")
	}
	nodes := f.Steps[3].Result["online_names"].([]any)
	templates := f.Steps[4].Result["ids"].([]any)
	create := f.Steps[6].Input
	if !reflect.DeepEqual(create["node"], nodes) || create["template"] != templates[0] || f.Steps[5].Result["id"] != create["template"] {
		t.Fatal("scan must select validated online node and existing template")
	}
	if f.Steps[7].Input["id"] != f.Steps[6].Result["id"] {
		t.Fatal("status must use returned task ID")
	}
	filter := f.Steps[8].Input["filter"].(map[string]any)
	if !reflect.DeepEqual(filter["task"], []any{"recon-example"}) || !reflect.DeepEqual(filter["project"], []any{"507f1f77bcf86cd799439011"}) {
		t.Fatal("results require task NAME and project ObjectID")
	}
	if f.Steps[9].Input["dimension"] != "task" || !reflect.DeepEqual(f.Steps[9].Input["targets"], []any{"recon-example"}) {
		t.Fatal("separate sync uses task name")
	}
	if len(f.FailedReads) != 2 {
		t.Fatal("must cover failed and incomplete scope reads")
	}
	for _, failure := range f.FailedReads {
		if (failure.Error == "" && failure.Complete) || len(failure.AllowedCalls) != 0 {
			t.Fatal("failed/incomplete scope read must stop, not replace project scope")
		}
	}
}
