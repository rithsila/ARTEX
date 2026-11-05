package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	maxBodyBytes   = 8 * 1024 * 1024 // 8 MiB
)

// Options configures the ARTEX API client.
type Options struct {
	BaseURL   string
	TokenFile string
	Timeout   time.Duration
}

// Client is a fixed-origin HTTP client for ARTEX API.
type Client struct {
	baseURL    *url.URL
	tokenFile  string
	httpClient *http.Client
}

// isLoopbackOrPrivateHost checks whether a hostname/ip string represents loopback or private LAN IP.
func isLoopbackOrPrivateHost(hostPort string) bool {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		host = hostPort
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		return true
	}
	return false
}

// New creates and validates a new Client.
func New(opts Options) (*Client, error) {
	if opts.BaseURL == "" {
		return nil, errors.New("base URL is required")
	}
	if opts.TokenFile == "" {
		return nil, errors.New("token file path is required")
	}

	u, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("scheme must be http or https, got %q", u.Scheme)
	}

	if u.Host == "" {
		return nil, errors.New("base URL missing host")
	}

	if u.User != nil {
		return nil, errors.New("base URL must not contain credentials")
	}

	if u.RawQuery != "" {
		return nil, errors.New("base URL must not contain query parameters")
	}

	if u.Fragment != "" {
		return nil, errors.New("base URL must not contain a fragment")
	}

	if u.Scheme == "http" && !isLoopbackOrPrivateHost(u.Host) {
		return nil, errors.New("plain HTTP is permitted only for loopback or private network addresses")
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	// Reject if token file is obviously a directory at creation time
	if fi, err := readTokenSecurely(opts.TokenFile); err != nil && strings.Contains(err.Error(), "must be a regular file") {
		return nil, err
	} else {
		_ = fi
	}

	// Direct transport, strictly ignoring environment proxies (Proxy: nil)
	transport := &http.Transport{
		Proxy:                 nil,
		DisableKeepAlives:     false,
		ResponseHeaderTimeout: timeout,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("redirects are not allowed")
		},
	}

	return &Client{
		baseURL:    u,
		tokenFile:  opts.TokenFile,
		httpClient: httpClient,
	}, nil
}

// Do performs a request against the fixed ARTEX API origin.
func (c *Client) Do(ctx context.Context, method, reqPath string, body any, out any) error {
	isMutation := method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete || method == http.MethodPatch
	dispatched := false

	// Path safety checks
	cleanPath := path.Clean(reqPath)
	if strings.Contains(reqPath, "..") || (!strings.HasPrefix(cleanPath, "/api") && !strings.HasPrefix(cleanPath, "api")) {
		return &APIError{
			Code:       "invalid_path",
			Status:     http.StatusBadRequest,
			Dispatched: false,
			Message:    "request path must begin with /api and cannot traverse directories",
		}
	}

	// Read current token (re-read per request to support rotation)
	token, err := readTokenSecurely(c.tokenFile)
	if err != nil {
		return &APIError{
			Code:       "token_read_failed",
			Status:     0,
			Dispatched: false,
			Message:    fmt.Sprintf("failed to read token: %v", err),
		}
	}

	// Construct target URL keeping the fixed origin and configured prefix
	target := *c.baseURL
	combinedPath := path.Join(target.Path, cleanPath)
	if !strings.HasPrefix(combinedPath, "/") {
		combinedPath = "/" + combinedPath
	}
	target.Path = combinedPath

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return &APIError{
				Code:       "marshal_error",
				Status:     0,
				Dispatched: false,
				Message:    "failed to marshal request body",
			}
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, target.String(), reqBody)
	if err != nil {
		return &APIError{
			Code:       "request_creation_failed",
			Status:     0,
			Dispatched: false,
			Message:    "failed to create request",
		}
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	if isMutation {
		dispatched = true
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return &APIError{
				Code:       "canceled",
				Status:     0,
				Dispatched: dispatched,
				Message:    "request was canceled",
			}
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return &APIError{
				Code:       "timeout",
				Status:     0,
				Dispatched: dispatched,
				Message:    "request timed out",
			}
		}
		return &APIError{
			Code:       "network_error",
			Status:     0,
			Dispatched: dispatched,
			Message:    "network request failed",
		}
	}
	defer resp.Body.Close()

	// Mutation reached server and got a response
	if !dispatched {
		dispatched = true
	}

	// Bounded reading up to maxBodyBytes + 1
	lr := io.LimitReader(resp.Body, maxBodyBytes+1)
	respBytes, err := io.ReadAll(lr)
	if err != nil {
		return &APIError{
			Code:       "read_error",
			Status:     resp.StatusCode,
			Dispatched: dispatched,
			Message:    "failed reading response body",
		}
	}

	if len(respBytes) > maxBodyBytes {
		return &APIError{
			Code:       "response_too_large",
			Status:     resp.StatusCode,
			Dispatched: dispatched,
			Message:    "response body exceeded 8 MiB limit",
		}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return &APIError{
			Code:       "auth_expired_or_invalid",
			Status:     http.StatusUnauthorized,
			Dispatched: dispatched,
			Message:    "authentication expired or invalid; update token file",
		}
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			Code:       "http_error",
			Status:     resp.StatusCode,
			Dispatched: dispatched,
			Message:    fmt.Sprintf("upstream returned HTTP %d", resp.StatusCode),
		}
	}

	if out != nil {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return &APIError{
				Code:       "invalid_response",
				Status:     resp.StatusCode,
				Dispatched: dispatched,
				Message:    "failed to decode JSON response",
			}
		}
	}

	return nil
}
