package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// FileAudit implements Audit by appending JSONL events to a secure local file.
type FileAudit struct {
	mu   sync.Mutex
	path string
	file *os.File
}

// NewFileAudit creates or opens an audit file at path.
// The file path must be absolute and not a symlink/directory, with mode 0600 on Unix.
func NewFileAudit(path string) (*FileAudit, error) {
	if path == "" {
		return nil, errors.New("audit file path cannot be empty")
	}
	if !filepath.IsAbs(path) {
		return nil, errors.New("audit file path must be absolute")
	}

	fi, err := os.Lstat(path)
	if err == nil {
		if fi.IsDir() {
			return nil, errors.New("audit path cannot be a directory")
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("audit path cannot be a symlink")
		}
		if runtime.GOOS != "windows" && fi.Mode().Perm() != 0600 {
			return nil, fmt.Errorf("audit file mode must be 0600, got %04o", fi.Mode().Perm())
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}

	dfi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if dfi.IsDir() {
		_ = f.Close()
		return nil, errors.New("audit path is a directory")
	}
	if runtime.GOOS != "windows" && dfi.Mode().Perm() != 0600 {
		_ = f.Close()
		return nil, fmt.Errorf("audit file descriptor mode must be 0600, got %04o", dfi.Mode().Perm())
	}

	return &FileAudit{
		path: path,
		file: f,
	}, nil
}

// Append writes a single AuditEvent line to the audit file.
func (a *FileAudit) Append(ctx context.Context, event AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}
	line = append(line, '\n')

	if _, err := a.file.Write(line); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}
	return nil
}
