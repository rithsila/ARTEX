package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Autumn-27/artex/internal/artexcontrol/client"
)

func writeSecureTokenFile(t *testing.T, dir, filename, token string, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(dir, filename)
	if err := os.WriteFile(p, []byte(token), mode); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}
	return p
}

func TestNewValidatesEndpoint(t *testing.T) {
	tmpDir := t.TempDir()
	validTokenFile := writeSecureTokenFile(t, tmpDir, "token.jwt", "jwt-token-123\n", 0600)

	cases := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid loopback http",
			baseURL: "http://127.0.0.1:8787",
			wantErr: false,
		},
		{
			name:    "valid loopback localhost http",
			baseURL: "http://localhost:8787/subpath",
			wantErr: false,
		},
		{
			name:    "valid ipv6 loopback http",
			baseURL: "http://[::1]:8787",
			wantErr: false,
		},
		{
			name:    "valid https non-loopback",
			baseURL: "https://artex.example.com",
			wantErr: false,
		},
		{
			name:    "reject plain http non-loopback",
			baseURL: "http://example.com:8787",
			wantErr: true,
		},
		{
			name:    "reject userinfo/credentials",
			baseURL: "https://user:pass@artex.example.com",
			wantErr: true,
		},
		{
			name:    "reject query params",
			baseURL: "https://artex.example.com?foo=bar",
			wantErr: true,
		},
		{
			name:    "reject fragment",
			baseURL: "https://artex.example.com#fragment",
			wantErr: true,
		},
		{
			name:    "reject missing host",
			baseURL: "https:///path",
			wantErr: true,
		},
		{
			name:    "reject invalid scheme",
			baseURL: "ftp://127.0.0.1:8787",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.New(client.Options{
				BaseURL:   tc.baseURL,
				TokenFile: validTokenFile,
			})
			if (err != nil) != tc.wantErr {
				t.Fatalf("New(%q) error = %v, wantErr = %v", tc.baseURL, err, tc.wantErr)
			}
		})
	}
}

func TestDoUsesFixedOriginAndPrefix(t *testing.T) {
	var requestedURI string
	var authHeader string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedURI = r.RequestURI
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "secret-token-val", 0600)

	baseURL := srv.URL + "/artex"
	c, err := client.New(client.Options{
		BaseURL:   baseURL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	type Resp struct {
		Status string `json:"status"`
	}
	var out Resp

	err = c.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if requestedURI != "/artex/api/health" {
		t.Errorf("expected URI /artex/api/health, got %q", requestedURI)
	}
	if authHeader != "Bearer secret-token-val" {
		t.Errorf("expected Bearer secret-token-val, got %q", authHeader)
	}
	if out.Status != "ok" {
		t.Errorf("expected status ok, got %q", out.Status)
	}

	err = c.Do(context.Background(), http.MethodGet, "../../../evil", nil, &out)
	if err == nil {
		t.Fatalf("expected error for path traversal, got nil")
	}
}

func TestTokenRotation(t *testing.T) {
	var seenTokens []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenTokens = append(seenTokens, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.jwt")
	if err := os.WriteFile(tokenPath, []byte("token-1\n"), 0600); err != nil {
		t.Fatalf("write token 1: %v", err)
	}

	c, err := client.New(client.Options{
		BaseURL:   srv.URL,
		TokenFile: tokenPath,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	var dummy map[string]any
	if err := c.Do(context.Background(), http.MethodGet, "/api/health", nil, &dummy); err != nil {
		t.Fatalf("Do 1 failed: %v", err)
	}

	if err := os.WriteFile(tokenPath, []byte("token-2\r\n"), 0600); err != nil {
		t.Fatalf("write token 2: %v", err)
	}

	if err := c.Do(context.Background(), http.MethodGet, "/api/health", nil, &dummy); err != nil {
		t.Fatalf("Do 2 failed: %v", err)
	}

	if len(seenTokens) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(seenTokens))
	}
	if seenTokens[0] != "Bearer token-1" {
		t.Errorf("expected first token Bearer token-1, got %q", seenTokens[0])
	}
	if seenTokens[1] != "Bearer token-2" {
		t.Errorf("expected second token Bearer token-2, got %q", seenTokens[1])
	}
}

func TestTokenFileRejectsUnsafeInputs(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Directory instead of regular file
	subDir := filepath.Join(tmpDir, "adir")
	_ = os.Mkdir(subDir, 0700)

	_, err := client.New(client.Options{
		BaseURL:   "http://127.0.0.1:8787",
		TokenFile: subDir,
	})
	if err == nil {
		t.Errorf("expected error for directory as token file, got nil")
	}

	// 2. Empty file
	emptyFile := filepath.Join(tmpDir, "empty.jwt")
	_ = os.WriteFile(emptyFile, []byte("   \n\t"), 0600)
	c, err := client.New(client.Options{
		BaseURL:   "http://127.0.0.1:8787",
		TokenFile: emptyFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	var out map[string]any
	err = c.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
	if err == nil || !strings.Contains(err.Error(), "token") {
		t.Errorf("expected empty token error, got: %v", err)
	}

	// 3. Permissive mode (Unix only)
	if runtime.GOOS != "windows" {
		permFile := filepath.Join(tmpDir, "perm.jwt")
		_ = os.WriteFile(permFile, []byte("valid-token"), 0644)
		cPerm, errPerm := client.New(client.Options{
			BaseURL:   "http://127.0.0.1:8787",
			TokenFile: permFile,
		})
		if errPerm != nil {
			t.Logf("New rejected permissive mode directly: %v", errPerm)
		} else {
			err = cPerm.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
			if err == nil {
				t.Errorf("expected error for permissive token file permissions 0644")
			}
		}

		// 4. Symlink
		targetFile := writeSecureTokenFile(t, tmpDir, "target.jwt", "valid-token", 0600)
		symlinkFile := filepath.Join(tmpDir, "symlink.jwt")
		_ = os.Symlink(targetFile, symlinkFile)

		cSym, errSym := client.New(client.Options{
			BaseURL:   "http://127.0.0.1:8787",
			TokenFile: symlinkFile,
		})
		if errSym == nil {
			err = cSym.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
			if err == nil {
				t.Errorf("expected error for symlink token file")
			}
		}
	}
}

func TestRedirectNeverForwardsToken(t *testing.T) {
	var targetSawAuth bool
	targetSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			targetSawAuth = true
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer targetSrv.Close()

	redirectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetSrv.URL, http.StatusFound)
	}))
	defer redirectSrv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "secret-token", 0600)
	c, err := client.New(client.Options{
		BaseURL:   redirectSrv.URL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	var out map[string]any
	err = c.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
	if err == nil {
		t.Fatalf("expected error on redirect, got nil")
	}
	if targetSawAuth {
		t.Errorf("target server saw authorization header despite redirect policy!")
	}
}

func TestEnvironmentProxyIgnored(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://192.0.2.1:8080")
	t.Setenv("HTTPS_PROXY", "http://192.0.2.1:8080")
	t.Setenv("http_proxy", "http://192.0.2.1:8080")
	t.Setenv("https_proxy", "http://192.0.2.1:8080")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "secret-token", 0600)
	c, err := client.New(client.Options{
		BaseURL:   srv.URL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	var out map[string]any
	err = c.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
	if err != nil {
		t.Fatalf("Do failed with environment proxy set: %v", err)
	}
}

func Test401Sanitized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"upstream internal jwt debug info: token expired at 12345"}`))
	}))
	defer srv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "expired-token", 0600)
	c, err := client.New(client.Options{
		BaseURL:   srv.URL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	var out map[string]any
	err = c.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
	if err == nil {
		t.Fatalf("expected error on 401, got nil")
	}

	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *client.APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "auth_expired_or_invalid" {
		t.Errorf("expected code auth_expired_or_invalid, got %q", apiErr.Code)
	}
	if apiErr.Status != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.Status)
	}
	if strings.Contains(err.Error(), "upstream internal jwt debug") || strings.Contains(err.Error(), "expired-token") {
		t.Errorf("error leaked sensitive upstream details: %v", err)
	}
}

func TestCancellation(t *testing.T) {
	reqReceived := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(reqReceived)
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "valid-token", 0600)
	c, err := client.New(client.Options{
		BaseURL:   srv.URL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-reqReceived
		cancel()
	}()

	var out map[string]any
	err = c.Do(ctx, http.MethodGet, "/api/health", nil, &out)
	if err == nil {
		t.Fatalf("expected cancellation error, got nil")
	}
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && apiErr.Code != "canceled" {
		t.Errorf("expected canceled error, got: %v", err)
	}
}

func TestOversizeBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"`))
		chunk := strings.Repeat("A", 1024*1024)
		for i := 0; i < 9; i++ {
			_, _ = w.Write([]byte(chunk))
		}
		_, _ = w.Write([]byte(`"}`))
	}))
	defer srv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "valid-token", 0600)
	c, err := client.New(client.Options{
		BaseURL:   srv.URL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	var out map[string]any
	err = c.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
	if err == nil {
		t.Fatalf("expected oversize error, got nil")
	}
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code != "response_too_large" {
			t.Errorf("expected code response_too_large, got %q", apiErr.Code)
		}
	} else if !strings.Contains(err.Error(), "too large") && !strings.Contains(err.Error(), "limit") {
		t.Errorf("expected error indicating response too large, got: %v", err)
	}
}

func TestNonJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body>Bad Gateway</body></html>`))
	}))
	defer srv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "valid-token", 0600)
	c, err := client.New(client.Options{
		BaseURL:   srv.URL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	var out map[string]any
	err = c.Do(context.Background(), http.MethodGet, "/api/health", nil, &out)
	if err == nil {
		t.Fatalf("expected non-json error, got nil")
	}
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code != "invalid_response" {
			t.Errorf("expected code invalid_response, got %q", apiErr.Code)
		}
	}
}

func TestWriteNotRetried(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"database failure"}`))
	}))
	defer srv.Close()

	tokenFile := writeSecureTokenFile(t, t.TempDir(), "token.jwt", "valid-token", 0600)
	c, err := client.New(client.Options{
		BaseURL:   srv.URL,
		TokenFile: tokenFile,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	var out map[string]any
	payload := map[string]any{"workers": 4}
	err = c.Do(context.Background(), http.MethodPut, "/api/settings", payload, &out)
	if err == nil {
		t.Fatalf("expected error from 500 server, got nil")
	}

	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *client.APIError, got %T: %v", err, err)
	}
	if !apiErr.Dispatched {
		t.Errorf("expected Dispatched to be true for write reaching server")
	}
	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected exactly 1 call without retries, got %d", count)
	}
}
