package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// Config for HTTP listener.
type HTTPListenerConfig struct {
	Host         string        `yaml:"host,omitempty"`
	Port         int           `yaml:"port,omitempty"`
	ReadTimeout  time.Duration `yaml:"read_timeout,omitempty"`
	WriteTimeout time.Duration `yaml:"write_timeout,omitempty"`
	IdleTimeout  time.Duration `yaml:"idle_timeout,omitempty"`
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
		Host         *string        `yaml:"host"`
		Port         *int           `yaml:"port"`
		ReadTimeout  *time.Duration `yaml:"read_timeout"`
		WriteTimeout *time.Duration `yaml:"write_timeout"`
		IdleTimeout  *time.Duration `yaml:"idle_timeout"`
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
		_ = l.srv.ListenAndServe()
	}()
}
