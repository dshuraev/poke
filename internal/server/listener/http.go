package listener

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"poke/internal/server/auth"
	"poke/internal/server/request"
	"strings"
	"time"
)

type HTTPListener struct {
	srv *http.Server
}

type httpCommandRequest struct {
	CommandID string `json:"command_id"`
}

// HTTPListenerTLSConfig defines TLS settings for the HTTP listener.
type HTTPListenerTLSConfig struct {
	CertFile string `yaml:"cert_file,omitempty"`
	KeyFile  string `yaml:"key_file,omitempty"`
}

// expandEnvStrict resolves environment variables and errors on missing entries.
func expandEnvStrict(input string) (string, error) {
	if input == "" {
		return input, nil
	}

	missing := map[string]struct{}{}
	expanded := os.Expand(input, func(name string) string {
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		missing[name] = struct{}{}
		return ""
	})

	if len(missing) == 0 {
		return expanded, nil
	}

	names := make([]string, 0, len(missing))
	for name := range missing {
		names = append(names, name)
	}

	return "", fmt.Errorf("missing environment variable(s): %s", strings.Join(names, ", "))
}

// UnmarshalYAML parses HTTP listener TLS config per docs/configuration/listener.md.
func (cfg *HTTPListenerTLSConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type httpListenerTLSConfigInput struct {
		CertFile *string `yaml:"cert_file"`
		KeyFile  *string `yaml:"key_file"`
	}

	*cfg = HTTPListenerTLSConfig{}

	var in httpListenerTLSConfigInput
	if err := unmarshal(&in); err != nil {
		return err
	}

	if in.CertFile != nil {
		certFile, err := expandEnvStrict(*in.CertFile)
		if err != nil {
			return fmt.Errorf("tls cert_file: %w", err)
		}
		cfg.CertFile = certFile
	}
	if in.KeyFile != nil {
		keyFile, err := expandEnvStrict(*in.KeyFile)
		if err != nil {
			return fmt.Errorf("tls key_file: %w", err)
		}
		cfg.KeyFile = keyFile
	}

	return nil
}

// validate enforces required HTTP listener TLS config invariants.
func (cfg HTTPListenerTLSConfig) validate() error {
	certFile := strings.TrimSpace(cfg.CertFile)
	keyFile := strings.TrimSpace(cfg.KeyFile)
	if certFile == "" || keyFile == "" {
		return fmt.Errorf("tls requires both cert_file and key_file")
	}
	if err := ensureReadableFile(certFile); err != nil {
		return fmt.Errorf("tls cert_file: %w", err)
	}
	if err := ensureReadableFile(keyFile); err != nil {
		return fmt.Errorf("tls key_file: %w", err)
	}
	return nil
}

// ensureReadableFile validates a path exists, is a file, and can be read.
func ensureReadableFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	file, err := os.Open(path) // #nosec G304 -- by design, comes from config
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck // Best-effort close; readability already established.

	return nil
}

// Config for HTTP listener.
type HTTPListenerConfig struct {
	Host         string                 `yaml:"host,omitempty"`
	Port         int                    `yaml:"port,omitempty"`
	ReadTimeout  time.Duration          `yaml:"read_timeout,omitempty"`
	WriteTimeout time.Duration          `yaml:"write_timeout,omitempty"`
	IdleTimeout  time.Duration          `yaml:"idle_timeout,omitempty"`
	TLS          *HTTPListenerTLSConfig `yaml:"tls,omitempty"`
	Auth         *auth.Auth             `yaml:"auth,omitempty"`
}

const (
	defaultHTTPListenerHost = "127.0.0.1"        // Default bind host when omitted.
	defaultHTTPListenerPort = 8008               // Default port when omitted.
	minHTTPListenerPort     = 1                  // Minimum allowed port value.
	maxHTTPListenerPort     = 65535              // Maximum allowed port value.
	httpShutdownTimeout     = 5 * time.Second    // Graceful shutdown timeout after context cancellation.
	httpListenerType        = "http"             // Listener type identifier used in auth contexts.
	httpAPITokenHeader      = "X-Poke-API-Token" // #nosec G101 -- Header key identifier, not a secret.
	httpAuthMethodHeader    = "X-Poke-Auth-Method"
)

// validateHTTPCommandAuth validates request-scoped auth when listener auth validators are configured.
func validateHTTPCommandAuth(cfg HTTPListenerConfig, headers http.Header) error {
	if cfg.Auth == nil || len(cfg.Auth.Validators) == 0 {
		return nil
	}

	method := strings.TrimSpace(headers.Get(httpAuthMethodHeader))
	if method == "" {
		return fmt.Errorf("auth method header %q is required", httpAuthMethodHeader)
	}

	validator, exists := cfg.Auth.Validators[method]
	if !exists {
		return fmt.Errorf("auth method %q is not configured", method)
	}

	authCtx, err := buildHTTPAuthContext(method, headers)
	if err != nil {
		return err
	}

	return validator.Validate(&authCtx)
}

// buildHTTPAuthContext maps a request auth method to its auth context.
func buildHTTPAuthContext(method string, headers http.Header) (auth.AuthContext, error) {
	switch method {
	case auth.AuthTypeAPIToken:
		token := strings.TrimSpace(headers.Get(httpAPITokenHeader))
		return auth.NewAPITokenContext(httpListenerType, token), nil
	default:
		return auth.AuthContext{}, fmt.Errorf("unsupported auth method %q", method)
	}
}

// UnmarshalYAML parses HTTP listener config per docs/configuration/listener.md.
func (cfg *HTTPListenerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type httpListenerConfigInput struct {
		Host         *string                `yaml:"host"`
		Port         *int                   `yaml:"port"`
		ReadTimeout  *time.Duration         `yaml:"read_timeout"`
		WriteTimeout *time.Duration         `yaml:"write_timeout"`
		IdleTimeout  *time.Duration         `yaml:"idle_timeout"`
		TLS          *HTTPListenerTLSConfig `yaml:"tls"`
		Auth         *auth.Auth             `yaml:"auth"`
	}

	*cfg = HTTPListenerConfig{
		Host: defaultHTTPListenerHost,
		Port: defaultHTTPListenerPort,
	}

	var in httpListenerConfigInput
	if err := unmarshal(&in); err != nil {
		return err
	}

	if in.Host != nil {
		cfg.Host = *in.Host
	}
	if in.Port != nil {
		cfg.Port = *in.Port
	}
	if in.ReadTimeout != nil {
		cfg.ReadTimeout = *in.ReadTimeout
	}
	if in.WriteTimeout != nil {
		cfg.WriteTimeout = *in.WriteTimeout
	}
	if in.IdleTimeout != nil {
		cfg.IdleTimeout = *in.IdleTimeout
	}
	if in.TLS != nil {
		cfg.TLS = in.TLS
	}
	if in.Auth != nil {
		cfg.Auth = in.Auth
	}

	return cfg.validate()
}

// validate enforces required HTTP listener config invariants.
func (cfg HTTPListenerConfig) validate() error {
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("host must not be empty")
	}
	if cfg.Port < minHTTPListenerPort || cfg.Port > maxHTTPListenerPort {
		return fmt.Errorf("port must be between %d and %d", minHTTPListenerPort, maxHTTPListenerPort)
	}
	if cfg.TLS != nil {
		if err := cfg.TLS.validate(); err != nil {
			return err
		}
	}
	if cfg.Auth == nil {
		return fmt.Errorf("auth is required for listener http")
	}
	if len(cfg.Auth.Validators) == 0 {
		return fmt.Errorf("auth must configure at least one method")
	}
	return nil
}

func (cfg HTTPListenerConfig) address() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

func (l *HTTPListener) Listen(ctx context.Context, cfg HTTPListenerConfig, ch chan<- request.CommandRequest) error {
	logger := slog.Default().With("component", "listener/http")

	l.srv = &http.Server{
		Addr:         cfg.address(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	logHTTPListenerStart(logger, cfg)
	l.srv.Handler = newHTTPHandler(ctx, cfg, ch)

	srvListener, err := buildHTTPServerListener(cfg)
	if err != nil {
		return err
	}

	startHTTPListenerShutdownLoop(ctx, l.srv, cfg.address(), logger)
	startHTTPServeLoop(l.srv, srvListener, cfg.address(), logger)

	return nil
}

func logHTTPListenerStart(logger *slog.Logger, cfg HTTPListenerConfig) {
	if cfg.TLS != nil {
		logger.Info("listener starting with tls", "event", "listener_starting_tls", "listener", "http", "address", cfg.address())
		return
	}
	logger.Info("listener starting without tls", "event", "listener_starting_plain", "listener", "http", "address", cfg.address())
}

func newHTTPHandler(ctx context.Context, cfg HTTPListenerConfig, ch chan<- request.CommandRequest) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPCommandRequest(ctx, cfg, ch, w, r)
	})
	return mux
}

func handleHTTPCommandRequest(ctx context.Context, cfg HTTPListenerConfig, ch chan<- request.CommandRequest, w http.ResponseWriter, r *http.Request) {
	logger := slog.Default().With("component", "listener/http")
	logger.Info("request received", "event", "request_received", "listener", "http", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req, err := decodeHTTPCommandRequest(r)
	if err != nil {
		logger.Warn("invalid json", "event", "request_invalid_json", "listener", "http", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.CommandID == "" {
		logger.Warn("missing command id", "event", "request_missing_command_id", "listener", "http", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := validateHTTPCommandAuth(cfg, r.Header); err != nil {
		logger.Warn("auth failed", "event", "request_auth_failed", "listener", "http", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr, "command_id", req.CommandID, "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !enqueueHTTPCommandRequest(ctx, ch, req.CommandID, logger) {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func decodeHTTPCommandRequest(r *http.Request) (httpCommandRequest, error) {
	var req httpCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpCommandRequest{}, err
	}
	return req, nil
}

func enqueueHTTPCommandRequest(ctx context.Context, ch chan<- request.CommandRequest, commandID string, logger *slog.Logger) bool {
	select {
	case <-ctx.Done():
		logger.Warn("context canceled before enqueue", "event", "request_enqueue_canceled", "listener", "http", "command_id", commandID)
		return false
	case ch <- request.CommandRequest{CommandID: commandID}:
		logger.Info("request enqueued", "event", "request_enqueued", "listener", "http", "command_id", commandID)
		return true
	}
}

func buildHTTPServerListener(cfg HTTPListenerConfig) (net.Listener, error) {
	rawListener, err := net.Listen("tcp", cfg.address())
	if err != nil {
		return nil, fmt.Errorf("start listener on %s: %w", cfg.address(), err)
	}
	if cfg.TLS == nil {
		return rawListener, nil
	}

	tlsCert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	if err != nil {
		_ = rawListener.Close()
		return nil, fmt.Errorf("load tls key pair: %w", err)
	}

	return tls.NewListener(rawListener, &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}), nil
}

func startHTTPListenerShutdownLoop(ctx context.Context, srv *http.Server, addr string, logger *slog.Logger) {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("listener shutdown failed", "event", "listener_shutdown_failed", "listener", "http", "address", addr, "error", err)
		}
	}()
}

func startHTTPServeLoop(srv *http.Server, listener net.Listener, addr string, logger *slog.Logger) {
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listener serve failed", "event", "listener_serve_failed", "listener", "http", "address", addr, "error", err)
		}
	}()
}
