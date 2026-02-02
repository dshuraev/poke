package server

import (
	"context"
	"poke/internal/server/dispatch"
	"poke/internal/server/listener"
	"poke/internal/server/request"
)

const defaultRequestBuffer = 16 // Default buffer for inbound command requests.

// Runtime bundles running server components for lifecycle management.
type Runtime struct {
	RequestChannel chan request.CommandRequest
	Dispatcher     *dispatch.SyncDispatcher
	Listeners      []listener.Listener
}

// Start wires configuration into listeners and the dispatcher, then starts them.
func Start(ctx context.Context, cfg Config) (*Runtime, error) {
	reqCh := make(chan request.CommandRequest, defaultRequestBuffer)
	registry := &cfg.Commands

	startedListeners, err := cfg.Listeners.StartAll(ctx, reqCh)
	if err != nil {
		return nil, err
	}

	executors := registry.ExecutorNames()
	dispatcher, err := dispatch.NewSyncDispatcher(ctx, registry, executors, reqCh)
	if err != nil {
		return nil, err
	}

	go dispatcher.Run()

	return &Runtime{
		RequestChannel: reqCh,
		Dispatcher:     dispatcher,
		Listeners:      startedListeners,
	}, nil
}
