package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"syscall"
)

// maxTokenFileSize limits token file read size (64 KiB is far larger than standard JWTs)
const maxTokenFileSize = 64 * 1024

// readTokenSecurely reads and validates a token file from the local filesystem.
func readTokenSecurely(path string) (string, error) {
	if path == "" {
		return "", errors.New("token file path cannot be empty")
	}

	// 1. Lstat to ensure it is not a symlink and is a regular file
	lfi, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("token file stat failed: %w", err)
	}

	if lfi.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("token file cannot be a symlink")
	}

	if !lfi.Mode().IsRegular() {
		return "", errors.New("token file must be a regular file")
	}

	// 2. Ownership and permissions check on Unix platforms
	if runtime.GOOS != "windows" {
		if lfi.Mode().Perm()&0077 != 0 {
			return "", fmt.Errorf("token file permissions (%04o) too open; must be 0600 or stricter", lfi.Mode().Perm())
		}

		if stat, ok := lfi.Sys().(*syscall.Stat_t); ok {
			if int(stat.Uid) != os.Getuid() {
				return "", errors.New("token file must be owned by the current user")
			}
		}
	}

	// 3. Open file descriptor and verify descriptor attributes
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open token file: %w", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("cannot stat open token file descriptor: %w", err)
	}

	if !fi.Mode().IsRegular() {
		return "", errors.New("token file descriptor is not a regular file")
	}

	if runtime.GOOS != "windows" {
		if fi.Mode().Perm()&0077 != 0 {
			return "", fmt.Errorf("token file permissions (%04o) too open", fi.Mode().Perm())
		}
		if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
			if int(stat.Uid) != os.Getuid() {
				return "", errors.New("token file must be owned by the current user")
			}
		}
	}

	// 4. Bounded read
	lr := io.LimitReader(f, maxTokenFileSize+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return "", fmt.Errorf("failed reading token file: %w", err)
	}

	if len(data) > maxTokenFileSize {
		return "", errors.New("token file exceeds size limit")
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.New("token file is empty")
	}

	return token, nil
}
