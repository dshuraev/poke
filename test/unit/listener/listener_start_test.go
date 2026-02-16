package listener_test

import (
	"context"
	"fmt"
	"net"
	"poke/internal/server/listener"
	"poke/internal/server/request"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestListenerStartAllReturnsErrorWhenHTTPPortInUse(t *testing.T) {
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

	input := []byte(fmt.Sprintf(`
http:
  host: 127.0.0.1
  port: %d
  auth:
    api_token:
      token: "secret"
`, addr.Port))

	var cfg listener.ListenerConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	requests := make(chan request.CommandRequest, 1)
	if _, err := cfg.StartAll(ctx, requests); err == nil {
		t.Fatalf("expected listener start error while port is occupied")
	}
}
