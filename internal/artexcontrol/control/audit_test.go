package control_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Autumn-27/artex/internal/artexcontrol/control"
)

func TestFileAuditRejectsUnsafePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Directory path
	_, err := control.NewFileAudit(tmpDir)
	if err == nil {
		t.Errorf("expected error for directory path, got nil")
	}

	// 2. Relative path
	_, err = control.NewFileAudit("relative.log")
	if err == nil {
		t.Errorf("expected error for relative path, got nil")
	}

	// 3. Symlink path
	if runtime.GOOS != "windows" {
		target := filepath.Join(tmpDir, "target.log")
		_ = os.WriteFile(target, []byte{}, 0600)
		sym := filepath.Join(tmpDir, "symlink.log")
		_ = os.Symlink(target, sym)

		_, err = control.NewFileAudit(sym)
		if err == nil {
			t.Errorf("expected error for symlink audit path, got nil")
		}
	}
}

func TestFileAuditModeAndNoSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	auditPath := filepath.Join(tmpDir, "audit.jsonl")

	audit, err := control.NewFileAudit(auditPath)
	if err != nil {
		t.Fatalf("NewFileAudit failed: %v", err)
	}

	// Verify file mode 0600 on Unix
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(auditPath)
		if err != nil {
			t.Fatalf("stat audit file: %v", err)
		}
		if fi.Mode().Perm() != 0600 {
			t.Errorf("expected mode 0600, got %04o", fi.Mode().Perm())
		}
	}

	rawPrompt := "Super sensitive secret prompt with sk-12345"
	rawToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.dummy"

	ev := control.AuditEvent{
		Timestamp:     time.Now().UTC(),
		OperationID:   "op-123",
		SessionID:     "ses-999",
		Operation:     "agent_prompt",
		Target:        "planner",
		Key:           "",
		BeforeHash:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		AfterHash:     "ca978112ca1bbdcafac231b39a23dc4da786081496140970e4e4632dc9600a1e",
		DispatchState: "dispatched",
		Outcome:       "verified",
		HTTPStatus:    200,
	}

	if err := audit.Append(context.Background(), ev); err != nil {
		t.Fatalf("audit.Append failed: %v", err)
	}

	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("read audit file: %v", err)
	}

	content := string(data)
	if strings.Contains(content, rawPrompt) || strings.Contains(content, rawToken) {
		t.Errorf("audit file leaked sensitive raw text: %s", content)
	}

	var parsed control.AuditEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse JSONL line: %v", err)
	}
	if parsed.OperationID != "op-123" || parsed.Outcome != "verified" {
		t.Errorf("unexpected event content: %+v", parsed)
	}
}
