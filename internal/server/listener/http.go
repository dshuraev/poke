package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
}

const (
	defaultHTTPListenerHost = "127.0.0.1" // Default bind host when omitted.
	defaultHTTPListenerPort = 8008        // Default port when omitted.
	minHTTPListenerPort     = 1           // Minimum allowed port value.
	maxHTTPListenerPort     = 65535       // Maximum allowed port value.
)

// UnmarshalYAML parses HTTP listener config per docs/configuration/listener.md.
func (cfg *HTTPListenerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type httpListenerConfigInput struct {
		Host         *string                `yaml:"host"`
		Port         *int                   `yaml:"port"`
		ReadTimeout  *time.Duration         `yaml:"read_timeout"`
		WriteTimeout *time.Duration         `yaml:"write_timeout"`
		IdleTimeout  *time.Duration         `yaml:"idle_timeout"`
		TLS          *HTTPListenerTLSConfig `yaml:"tls"`
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
	return nil
}

func (cfg HTTPListenerConfig) address() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

func (l *HTTPListener) Listen(ctx context.Context, cfg HTTPListenerConfig, ch chan<- request.CommandRequest) {
	l.srv = &http.Server{
		Addr:         cfg.address(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if cfg.TLS != nil {
		log.Printf("http listener: starting with tls addr=%s", cfg.address())
	} else {
		log.Printf("http listener: starting without tls addr=%s", cfg.address())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http request: method=%s path=%s remote=%s", r.Method, r.URL.Path, r.RemoteAddr)
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req httpCommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("http request: invalid json: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.CommandID == "" {
			log.Printf("http request: missing command_id")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		select {
		case <-ctx.Done():
			log.Printf("http request: context canceled before enqueue command_id=%s", req.CommandID)
			w.WriteHeader(http.StatusServiceUnavailable)
		case ch <- request.CommandRequest{CommandID: req.CommandID}:
			log.Printf("http request: enqueued command_id=%s", req.CommandID)
			w.WriteHeader(http.StatusAccepted)
		}
	})

	l.srv.Handler = mux
	go func() {
		if cfg.TLS != nil {
			_ = l.srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			return
		}
		_ = l.srv.ListenAndServe()
	}()
}
