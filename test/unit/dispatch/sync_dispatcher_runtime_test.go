package dispatch_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"poke/internal/server/dispatch"
	"poke/internal/server/executor"
	"poke/internal/server/request"
)

func TestNewSyncDispatcherRejectsUnknownExecutor(t *testing.T) {
	d := dispatch.NewCommandRegistry(map[string]executor.Command{})
	reqCh := make(chan request.CommandRequest)

	if _, err := dispatch.NewSyncDispatcher(context.Background(), d, []string{"unknown"}, reqCh); err == nil {
		t.Fatalf("expected error for unknown executor")
	}
}

func TestSyncDispatcherRunStopsWhenContextIsCanceled(t *testing.T) {
	reg := dispatch.NewCommandRegistry(map[string]executor.Command{})
	reqCh := make(chan request.CommandRequest)
	ctx, cancel := context.WithCancel(context.Background())

	d, err := dispatch.NewSyncDispatcher(ctx, reg, nil, reqCh)
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}

	logs := captureLogs(t, func() {
		done := make(chan struct{})
		go func() {
			d.Run()
			close(done)
		}()

		cancel()
		waitDone(t, done)
	})

	if !strings.Contains(logs, "event=context_canceled") {
		t.Fatalf("expected context canceled log, got %q", logs)
	}
}

func TestSyncDispatcherRunStopsWhenRequestChannelCloses(t *testing.T) {
	reg := dispatch.NewCommandRegistry(map[string]executor.Command{})
	reqCh := make(chan request.CommandRequest)

	d, err := dispatch.NewSyncDispatcher(context.Background(), reg, nil, reqCh)
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}

	logs := captureLogs(t, func() {
		done := make(chan struct{})
		go func() {
			d.Run()
			close(done)
		}()

		close(reqCh)
		waitDone(t, done)
	})

	if !strings.Contains(logs, "event=request_channel_closed") {
		t.Fatalf("expected channel-closed log, got %q", logs)
	}
}

func TestSyncDispatcherRunSkipsUnknownCommand(t *testing.T) {
	reg := dispatch.NewCommandRegistry(map[string]executor.Command{})
	reqCh := make(chan request.CommandRequest, 1)

	d, err := dispatch.NewSyncDispatcher(context.Background(), reg, nil, reqCh)
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}

	logs := captureLogs(t, func() {
		done := make(chan struct{})
		go func() {
			d.Run()
			close(done)
		}()

		reqCh <- request.CommandRequest{CommandID: "missing"}
		close(reqCh)
		waitDone(t, done)
	})

	if !strings.Contains(logs, "event=command_lookup_failed") || !strings.Contains(logs, "command_id=missing") {
		t.Fatalf("expected missing command log, got %q", logs)
	}
}

func TestSyncDispatcherRunHandlesUnknownExecutor(t *testing.T) {
	reg := dispatch.NewCommandRegistry(map[string]executor.Command{
		"bad": {
			Args: []string{"true"},
			Env:  executor.NewEnvDefault(),
		},
	})
	reqCh := make(chan request.CommandRequest, 1)

	d, err := dispatch.NewSyncDispatcher(context.Background(), reg, reg.ExecutorNames(), reqCh)
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}

	logs := captureLogs(t, func() {
		done := make(chan struct{})
		go func() {
			d.Run()
			close(done)
		}()

		reqCh <- request.CommandRequest{CommandID: "bad"}
		close(reqCh)
		waitDone(t, done)
	})

	if !strings.Contains(logs, "event=unknown_executor") {
		t.Fatalf("expected unknown executor log, got %q", logs)
	}
}

func TestSyncDispatcherRunExecutesCommand(t *testing.T) {
	reg := dispatch.NewCommandRegistry(map[string]executor.Command{
		"ok": {
			Name:     "ok",
			Args:     []string{"true"},
			Env:      executor.NewEnvDefault(),
			Executor: "bin",
		},
	})
	reqCh := make(chan request.CommandRequest, 1)

	d, err := dispatch.NewSyncDispatcher(context.Background(), reg, reg.ExecutorNames(), reqCh)
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}

	logs := captureLogs(t, func() {
		done := make(chan struct{})
		go func() {
			d.Run()
			close(done)
		}()

		reqCh <- request.CommandRequest{CommandID: "ok"}
		close(reqCh)
		waitDone(t, done)
	})

	if !strings.Contains(logs, "event=command_execution_completed") || !strings.Contains(logs, "exit_code=0") {
		t.Fatalf("expected successful execution log, got %q", logs)
	}
}

func captureLogs(t *testing.T, run func()) string {
	t.Helper()

	var buf bytes.Buffer
	oldDefault := slog.Default()
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
	})
	slog.SetDefault(slog.New(handler))
	defer func() {
		slog.SetDefault(oldDefault)
	}()

	run()
	return buf.String()
}

func waitDone(t *testing.T, done <-chan struct{}) {
	t.Helper()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for dispatcher")
	}
}
