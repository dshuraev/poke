package server_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"poke/internal/server"
)

func TestStartReturnsRuntimeForValidConfig(t *testing.T) {
	port := reserveFreePort(t)
	cfg := mustParseServerConfig(t, fmt.Sprintf(`
commands:
  ok: ["true"]
listeners:
  http:
    host: 127.0.0.1
    port: %d
    auth:
      api_token:
        token: "secret"
`, port))

	ctx, cancel := context.WithCancel(context.Background())
	runtime, err := server.Start(ctx, cfg)
	if err != nil {
		cancel()
		t.Fatalf("start: %v", err)
	}

	if runtime == nil {
		t.Fatalf("runtime: expected non-nil")
	}
	if runtime.Dispatcher == nil {
		t.Fatalf("dispatcher: expected non-nil")
	}
	if len(runtime.Listeners) != 1 {
		t.Fatalf("listeners: got %d want 1", len(runtime.Listeners))
	}

	cancel()
	close(runtime.RequestChannel)
}

func TestStartReturnsErrorWhenListenerCannotBind(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() {
		_ = occupied.Close()
	}()

	addr, ok := occupied.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("address type: got %T", occupied.Addr())
	}

	cfg := mustParseServerConfig(t, fmt.Sprintf(`
commands:
  ok: ["true"]
listeners:
  http:
    host: 127.0.0.1
    port: %d
    auth:
      api_token:
        token: "secret"
`, addr.Port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := server.Start(ctx, cfg); err == nil {
		t.Fatalf("expected start error when listener port is in use")
	}
}

func mustParseServerConfig(t *testing.T, input string) server.Config {
	t.Helper()

	cfg, err := server.Parse([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return cfg
}

func reserveFreePort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() {
		_ = ln.Close()
	}()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("address type: got %T", ln.Addr())
	}

	return addr.Port
}
