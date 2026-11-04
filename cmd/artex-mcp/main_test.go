package main_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	main "github.com/Autumn-27/artex/cmd/artex-mcp"
)

var (
	builtBinaryOnce sync.Once
	builtBinaryPath string
	builtBinaryErr  error
)

func getBuiltBinary(t *testing.T) string {
	t.Helper()
	builtBinaryOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "artex-mcp-test-bin-*")
		if err != nil {
			builtBinaryErr = err
			return
		}
		binPath := filepath.Join(tmpDir, "artex-mcp")
		cmd := exec.Command("go", "build", "-o", binPath, ".")
		cmd.Dir = "."
		out, err := cmd.CombinedOutput()
		if err != nil {
			builtBinaryErr = fmt.Errorf("build failed: %w: %s", err, string(out))
			return
		}
		builtBinaryPath = binPath
	})
	if builtBinaryErr != nil {
		t.Fatalf("failed to build test binary: %v", builtBinaryErr)
	}
	return builtBinaryPath
}

func TestStartupConfig(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.jwt")
	_ = os.WriteFile(tokenPath, []byte("valid-token"), 0600)
	auditPath := filepath.Join(tmpDir, "audit.jsonl")

	cases := []struct {
		name      string
		env       map[string]string
		wantError string
	}{
		{
			name: "missing base url",
			env: map[string]string{
				"ARTEX_TOKEN_FILE": tokenPath,
			},
			wantError: "ARTEX_BASE_URL is required",
		},
		{
			name: "missing token file",
			env: map[string]string{
				"ARTEX_BASE_URL": "http://127.0.0.1:8787",
			},
			wantError: "ARTEX_TOKEN_FILE is required",
		},
		{
			name: "invalid write enabled toggle",
			env: map[string]string{
				"ARTEX_BASE_URL":          "http://127.0.0.1:8787",
				"ARTEX_TOKEN_FILE":        tokenPath,
				"ARTEX_MCP_WRITE_ENABLED": "yes",
			},
			wantError: "ARTEX_MCP_WRITE_ENABLED must be 'true' or 'false'",
		},
		{
			name: "write enabled without audit file",
			env: map[string]string{
				"ARTEX_BASE_URL":          "http://127.0.0.1:8787",
				"ARTEX_TOKEN_FILE":        tokenPath,
				"ARTEX_MCP_WRITE_ENABLED": "true",
			},
			wantError: "ARTEX_MCP_AUDIT_FILE is required when write mode is enabled",
		},
		{
			name: "valid read-only config",
			env: map[string]string{
				"ARTEX_BASE_URL":   "http://127.0.0.1:8787",
				"ARTEX_TOKEN_FILE": tokenPath,
			},
			wantError: "",
		},
		{
			name: "valid write config",
			env: map[string]string{
				"ARTEX_BASE_URL":          "http://127.0.0.1:8787",
				"ARTEX_TOKEN_FILE":        tokenPath,
				"ARTEX_MCP_WRITE_ENABLED": "true",
				"ARTEX_MCP_AUDIT_FILE":    auditPath,
			},
			wantError: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := main.LoadConfigFromEnv(func(k string) string {
				return tc.env[k]
			})
			if tc.wantError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantError)
				}
				if err.Error() == "" || !testingContains(err.Error(), tc.wantError) {
					t.Errorf("expected error containing %q, got %v", tc.wantError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.BaseURL != tc.env["ARTEX_BASE_URL"] {
					t.Errorf("expected BaseURL %s, got %s", tc.env["ARTEX_BASE_URL"], cfg.BaseURL)
				}
			}
		})
	}
}

func testingContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || filepath.Base(s) == substr || (len(substr) > 0 && substrIn(s, substr)))
}

func substrIn(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestStdoutProtocolOnly(t *testing.T) {
	bin := getBuiltBinary(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/health" {
			_, _ = w.Write([]byte(`{"ok":true,"service":"artex"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.jwt")
	if err := os.WriteFile(tokenPath, []byte("dummy-token"), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"ARTEX_BASE_URL="+ts.URL,
		"ARTEX_TOKEN_FILE="+tokenPath,
		"ARTEX_MCP_WRITE_ENABLED=false",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start companion process: %v", err)
	}

	// Send JSON-RPC initialize request
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-runner",
				"version": "1.0",
			},
		},
	}
	reqData, _ := json.Marshal(initReq)
	_, _ = stdin.Write(append(reqData, '\n'))

	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatalf("expected stdout line, got EOF. Stderr was: %s", stderrBuf.String())
	}
	line := scanner.Bytes()

	var rpcResp struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   any             `json:"error"`
	}
	if err := json.Unmarshal(line, &rpcResp); err != nil {
		t.Fatalf("stdout contained non-JSON-RPC text: %q (err: %v)", string(line), err)
	}
	if rpcResp.JSONRPC != "2.0" {
		t.Fatalf("expected JSON-RPC 2.0 on stdout, got: %s", string(line))
	}

	_ = stdin.Close()
	_ = cmd.Wait()

	// Verify stderr got log messages (e.g., "starting ARTEX control MCP companion")
	if !strings.Contains(stderrBuf.String(), "starting ARTEX control MCP companion") {
		t.Errorf("expected startup log on stderr, got: %s", stderrBuf.String())
	}
}

func TestProcessExitOnEOF(t *testing.T) {
	bin := getBuiltBinary(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"service":"artex"}`))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.jwt")
	_ = os.WriteFile(tokenPath, []byte("dummy-token"), 0600)

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"ARTEX_BASE_URL="+ts.URL,
		"ARTEX_TOKEN_FILE="+tokenPath,
		"ARTEX_MCP_WRITE_ENABLED=false",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}
	_, err = cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start companion process: %v", err)
	}

	// Immediately close stdin to simulate EOF
	_ = stdin.Close()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			// Some platforms or exit codes are non-zero when interrupted or closed, check
			t.Logf("process exited on EOF: %v", err)
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("process did not exit on stdin EOF within 5 seconds")
	}
}
