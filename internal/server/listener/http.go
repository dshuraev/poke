package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"poke/internal/server/request"
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
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
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
