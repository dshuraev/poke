package listener_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"poke/internal/server/auth"
	"poke/internal/server/listener"
	"poke/internal/server/request"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

func TestHTTPListenerRequestWithAuthAcceptsValidAPIToken(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	startHTTPListener(t, cfg, reqCh)

	resp := putJSONRequestWithRetry(
		t,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{"command_id":"uptime"}`,
		map[string]string{
			"Content-Type":       "application/json",
			"X-Poke-Auth-Method": "api_token",
			"X-Poke-API-Token":   "secret-token",
		},
	)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusAccepted)
	}

	select {
	case got := <-reqCh:
		if got.CommandID != "uptime" {
			t.Fatalf("command_id: got %q", got.CommandID)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected command to be enqueued")
	}
}

func TestHTTPListenerRequestWithAuthRejectsInvalidAPIToken(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	startHTTPListener(t, cfg, reqCh)

	resp := putJSONRequestWithRetry(
		t,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{"command_id":"uptime"}`,
		map[string]string{
			"Content-Type":       "application/json",
			"X-Poke-Auth-Method": "api_token",
			"X-Poke-API-Token":   "wrong-token",
		},
	)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}

	select {
	case got := <-reqCh:
		t.Fatalf("unexpected command enqueued: %#v", got)
	default:
	}
}

func TestHTTPListenerRequestWithoutAuthMethodHeaderRejectsRequest(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	startHTTPListener(t, cfg, reqCh)

	resp := putJSONRequestWithRetry(
		t,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{"command_id":"uptime"}`,
		map[string]string{
			"Content-Type":     "application/json",
			"X-Poke-API-Token": "secret-token",
		},
	)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}

	select {
	case got := <-reqCh:
		t.Fatalf("unexpected command enqueued: %#v", got)
	default:
	}
}

func startHTTPListener(t *testing.T, cfg listener.HTTPListenerConfig, reqCh chan<- request.CommandRequest) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var httpListener listener.HTTPListener
	if err := httpListener.Listen(ctx, cfg, reqCh); err != nil {
		t.Fatalf("listen: %v", err)
	}
}

func mustHTTPListenerConfigWithToken(t *testing.T, port int, token string) listener.HTTPListenerConfig {
	t.Helper()

	input := fmt.Sprintf(`
host: 127.0.0.1
port: %d
auth:
  %s:
    token: %q
`, port, auth.AuthTypeAPIToken, token)

	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return cfg
}

func reserveTCPPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() {
		if closeErr := ln.Close(); closeErr != nil {
			t.Fatalf("close listener: %v", closeErr)
		}
	}()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("address type: got %T", ln.Addr())
	}

	return addr.Port
}

func putJSONRequestWithRetry(t *testing.T, url string, body string, headers map[string]string) *http.Response {
	t.Helper()

	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(2 * time.Second)
	var lastErr error

	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err == nil {
			return resp
		}
		lastErr = err
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("request did not succeed before timeout: %v", lastErr)
	return nil
}
