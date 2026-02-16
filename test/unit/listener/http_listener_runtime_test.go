package listener_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"poke/internal/server/auth"
	"poke/internal/server/listener"
	"poke/internal/server/request"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

type allowValidator struct{}

func (allowValidator) Validate(_ *auth.AuthContext) error {
	return nil
}

func TestHTTPListenerListenWithTLSRejectsInvalidKeyPair(t *testing.T) {
	port := reserveTCPPort(t)
	dir := t.TempDir()
	certFile := writeTestFile(t, dir, "cert.pem", "not-a-cert")
	keyFile := writeTestFile(t, dir, "key.pem", "not-a-key")

	cfg := mustHTTPListenerTLSConfigWithToken(t, port, "secret-token", certFile, keyFile)
	reqCh := make(chan request.CommandRequest, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var l listener.HTTPListener
	if err := l.Listen(ctx, cfg, reqCh); err == nil {
		t.Fatalf("expected tls keypair load error")
	}
}

func TestHTTPListenerListenWithTLSAcceptsAuthenticatedRequest(t *testing.T) {
	port := reserveTCPPort(t)
	certFile, keyFile := writeSelfSignedTLSFiles(t, t.TempDir())
	cfg := mustHTTPListenerTLSConfigWithToken(t, port, "secret-token", certFile, keyFile)

	reqCh := make(chan request.CommandRequest, 1)
	cancel := startHTTPListenerWithCancel(t, cfg, reqCh)
	defer cancel()

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true}, // #nosec G402 -- self-signed test cert.
		},
	}

	resp, err := requestWithRetry(
		client,
		http.MethodPut,
		fmt.Sprintf("https://127.0.0.1:%d/", port),
		`{"command_id":"uptime"}`,
		map[string]string{
			"Content-Type":       "application/json",
			"X-Poke-Auth-Method": "api_token",
			"X-Poke-API-Token":   "secret-token",
		},
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusAccepted)
	}

	select {
	case got := <-reqCh:
		if got.CommandID != "uptime" {
			t.Fatalf("command_id: got %q want %q", got.CommandID, "uptime")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected command to be enqueued")
	}
}

func TestHTTPListenerRequestRejectsUnknownConfiguredAuthContext(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := listener.HTTPListenerConfig{
		Host: "127.0.0.1",
		Port: port,
		Auth: &auth.Auth{
			Validators: map[string]auth.Validator{
				"custom_auth": allowValidator{},
			},
		},
	}

	reqCh := make(chan request.CommandRequest, 1)
	cancel := startHTTPListenerWithCancel(t, cfg, reqCh)
	defer cancel()

	resp, err := requestWithRetry(
		&http.Client{Timeout: 2 * time.Second},
		http.MethodPut,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{"command_id":"uptime"}`,
		map[string]string{
			"Content-Type":       "application/json",
			"X-Poke-Auth-Method": "custom_auth",
		},
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestHTTPListenerRequestRejectsUnknownAuthMethodHeader(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	cancel := startHTTPListenerWithCancel(t, cfg, reqCh)
	defer cancel()

	resp, err := requestWithRetry(
		&http.Client{Timeout: 2 * time.Second},
		http.MethodPut,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{"command_id":"uptime"}`,
		map[string]string{
			"Content-Type":       "application/json",
			"X-Poke-Auth-Method": "unknown",
			"X-Poke-API-Token":   "secret-token",
		},
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestHTTPListenerRequestRejectsInvalidJSON(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	cancel := startHTTPListenerWithCancel(t, cfg, reqCh)
	defer cancel()

	resp, err := requestWithRetry(
		&http.Client{Timeout: 2 * time.Second},
		http.MethodPut,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{"command_id":"uptime"`,
		map[string]string{
			"Content-Type": "application/json",
		},
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHTTPListenerRequestRejectsMissingCommandID(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	cancel := startHTTPListenerWithCancel(t, cfg, reqCh)
	defer cancel()

	resp, err := requestWithRetry(
		&http.Client{Timeout: 2 * time.Second},
		http.MethodPut,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{}`,
		map[string]string{
			"Content-Type": "application/json",
		},
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHTTPListenerRequestRejectsMethodNotAllowed(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	cancel := startHTTPListenerWithCancel(t, cfg, reqCh)
	defer cancel()

	resp, err := requestWithRetry(
		&http.Client{Timeout: 2 * time.Second},
		http.MethodGet,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		"",
		nil,
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHTTPListenerRequestReturnsServiceUnavailableWhenContextCanceled(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var l listener.HTTPListener
	if err := l.Listen(ctx, cfg, reqCh); err != nil {
		t.Fatalf("listen: %v", err)
	}

	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := requestWithRetry(
			&http.Client{Timeout: 2 * time.Second},
			http.MethodPut,
			fmt.Sprintf("http://127.0.0.1:%d/", port),
			`{"command_id":"uptime"}`,
			map[string]string{
				"Content-Type":       "application/json",
				"X-Poke-Auth-Method": "api_token",
				"X-Poke-API-Token":   "secret-token",
			},
			2*time.Second,
		)
		done <- result{resp: resp, err: err}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("request: %v", got.err)
		}
		defer func() { _ = got.resp.Body.Close() }()
		if got.resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("status: got %d want %d", got.resp.StatusCode, http.StatusServiceUnavailable)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for response")
	}
}

func TestHTTPListenerShutdownStopsAcceptingRequests(t *testing.T) {
	port := reserveTCPPort(t)
	cfg := mustHTTPListenerConfigWithToken(t, port, "secret-token")

	reqCh := make(chan request.CommandRequest, 1)
	cancel := startHTTPListenerWithCancel(t, cfg, reqCh)

	resp, err := requestWithRetry(
		&http.Client{Timeout: 2 * time.Second},
		http.MethodPut,
		fmt.Sprintf("http://127.0.0.1:%d/", port),
		`{"command_id":"uptime"}`,
		map[string]string{
			"Content-Type":       "application/json",
			"X-Poke-Auth-Method": "api_token",
			"X-Poke-API-Token":   "secret-token",
		},
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("request before shutdown: %v", err)
	}
	_ = resp.Body.Close()

	cancel()

	client := &http.Client{Timeout: 250 * time.Millisecond}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err = requestOnce(
			client,
			http.MethodPut,
			fmt.Sprintf("http://127.0.0.1:%d/", port),
			`{"command_id":"uptime"}`,
			map[string]string{
				"Content-Type":       "application/json",
				"X-Poke-Auth-Method": "api_token",
				"X-Poke-API-Token":   "secret-token",
			},
		)
		if err != nil {
			return
		}
		_ = resp.Body.Close()
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("expected listener to stop accepting requests after cancellation")
}

func startHTTPListenerWithCancel(t *testing.T, cfg listener.HTTPListenerConfig, reqCh chan<- request.CommandRequest) context.CancelFunc {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var l listener.HTTPListener
	if err := l.Listen(ctx, cfg, reqCh); err != nil {
		t.Fatalf("listen: %v", err)
	}

	return cancel
}

func mustHTTPListenerTLSConfigWithToken(t *testing.T, port int, token, certFile, keyFile string) listener.HTTPListenerConfig {
	t.Helper()

	input := fmt.Sprintf(`
host: 127.0.0.1
port: %d
tls:
  cert_file: %q
  key_file: %q
auth:
  api_token:
    token: %q
`, port, certFile, keyFile, token)

	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return cfg
}

func requestWithRetry(client *http.Client, method, url, body string, headers map[string]string, wait time.Duration) (*http.Response, error) {
	deadline := time.Now().Add(wait)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := requestOnce(client, method, url, body, headers)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		time.Sleep(25 * time.Millisecond)
	}

	return nil, lastErr
}

func requestOnce(client *http.Client, method, url, body string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}

func writeSelfSignedTLSFiles(t *testing.T, dir string) (certFile string, keyFile string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "127.0.0.1",
		},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(5 * time.Minute),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if certBlock == nil || keyBlock == nil {
		t.Fatalf("encode pem: got nil blocks")
	}

	certFile = writeTestFile(t, dir, "server.crt", string(certBlock))
	keyFile = writeTestFile(t, dir, "server.key", string(keyBlock))
	return certFile, keyFile
}

func writeTestFile(t *testing.T, dir, name, data string) string {
	t.Helper()

	path := fmt.Sprintf("%s/%s", dir, name)
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
	return path
}
