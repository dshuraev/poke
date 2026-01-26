package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"poke/internal/server/dispatch"
	"poke/internal/server/executor"
	"poke/internal/server/listener"
	"poke/internal/server/request"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reqCh := make(chan request.CommandRequest, 16)

	echoID := "11111111-1111-1111-1111-111111111111"
	registry := dispatch.NewCommandRegistry(map[string]executor.Command{
		echoID: {
			Name:        "echo",
			Description: "Print a greeting to stdout",
			Args:        []string{"echo", "hello from poke"},
			Executor:    "bin",
			Env: executor.Env{
				Strategy: executor.EnvStrategyInherit,
				Vals:     executor.EnvMap{},
			},
			Timeout: 5 * time.Second,
		},
	})

	httpCfg := listener.HTTPListenerConfig{
		Host:         "127.0.0.1",
		Port:         8080,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	httpListener := &listener.HTTPListener{}
	httpListener.Listen(ctx, httpCfg, reqCh)

	dispatcher, err := dispatch.NewSyncDispatcher(ctx, registry, []string{"bin"}, reqCh)
	if err != nil {
		log.Fatalf("dispatch setup failed: %v", err)
	}

	go dispatcher.Run()

	log.Printf("server listening on http://%s:%d", httpCfg.Host, httpCfg.Port)
	log.Printf("sample command registered: %s", echoID)

	<-ctx.Done()
	close(reqCh)
	log.Printf("server shutting down")
}
