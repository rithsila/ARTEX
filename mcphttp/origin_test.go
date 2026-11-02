package mcphttp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCredentialsNeverCrossOrigin(t *testing.T) {
	for _, mode := range []string{"endpoint", "sse-redirect", "http-redirect", "message-redirect"} {
		t.Run(mode, func(t *testing.T) {
			var received atomic.Int32
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received.Add(1)
				w.WriteHeader(http.StatusForbidden)
			}))
			defer target.Close()
			source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if mode == "http-redirect" || mode == "sse-redirect" || r.URL.Path == "/message" {
					http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
					return
				}
				endpoint := target.URL
				if mode == "message-redirect" {
					endpoint = "/message"
				}
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpoint)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			defer source.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			headers := map[string]string{"Authorization": "Bearer dummy", "X-API-Key": "dummy"}
			var err error
			if mode == "http-redirect" {
				_, err = New(ctx, "test", source.URL, headers, false)
			} else {
				_, err = NewSSE(ctx, "test", source.URL, headers, false)
			}
			if err == nil {
				t.Error("expected rejection")
			}
			if n := received.Load(); n != 0 {
				t.Errorf("cross-origin destination received %d requests with configured credentials", n)
			}
		})
	}
}

func TestSSEEndpointOriginConstraints(t *testing.T) {
	for _, tc := range []struct {
		endpoint string
		allowed  bool
	}{
		{"message", true}, {"/message", true}, {"https://example.com:443/message", true},
		{"https://EXAMPLE.com/message", true}, {"http://example.com/message", false},
		{"https://example.com:444/message", false}, {"//other.example/message", false},
		{"https://user:pass@example.com/message", false}, {"file:///message", false},
		{"https://example.com:0/message", false}, {"https://example.com:99999/message", false},
	} {
		t.Run(tc.endpoint, func(t *testing.T) {
			_, err := readSSEEndpoint(bufReader("data: "+tc.endpoint+"\n"), "https://example.com/dir/sse")
			if (err == nil) != tc.allowed {
				t.Fatalf("allowed=%v, error=%v", tc.allowed, err)
			}
		})
	}
}

func bufReader(s string) *bufio.Reader { return bufio.NewReader(strings.NewReader(s)) }

func TestHTTPSameOriginRedirectPreservesCredentials(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/rpc", http.StatusTemporaryRedirect)
			return
		}
		if r.Header.Get("Authorization") != "Bearer dummy" || r.Header.Get("X-API-Key") != "dummy" {
			t.Error("same-origin credentials lost")
		}
		calls.Add(1)
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if req.ID == 0 {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{}}`, req.ID)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c, err := New(ctx, "test", srv.URL+"/start", map[string]string{"Authorization": "Bearer dummy", "X-API-Key": "dummy"}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if calls.Load() != 2 {
		t.Fatalf("got %d calls, want initialize and notification", calls.Load())
	}
}

func TestRejectInvalidInitialEndpoints(t *testing.T) {
	for _, endpoint := range []string{"file:///tmp/mcp", "https://user:pass@example.com/mcp", "https://example.com:0", "https://example.com:65536", "https://example.com:"} {
		for _, constructor := range []func(context.Context, string, string, map[string]string, bool) (*Client, error){New, NewSSE} {
			if c, err := constructor(context.Background(), "test", endpoint, nil, false); err == nil {
				c.Close()
				t.Error("invalid initial endpoint accepted")
			}
		}
	}
}

func TestRedirectOriginConstraints(t *testing.T) {
	hc := &http.Client{}
	if err := restrictOrigin(hc, "https://example.com/start"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		target  string
		allowed bool
	}{
		{"https://EXAMPLE.com:443/next", true},
		{"http://example.com/next", false},
		{"https://example.com:444/next", false},
		{"https://other.example/next", false},
		{"https://user:pass@example.com/next", false},
		{"file:///next", false},
	} {
		u, err := url.Parse(tc.target)
		if err != nil {
			t.Fatal(err)
		}
		err = hc.CheckRedirect(&http.Request{URL: u}, nil)
		if (err == nil) != tc.allowed {
			t.Errorf("target %q allowed=%v error=%v", tc.target, tc.allowed, err)
		}
	}
}
