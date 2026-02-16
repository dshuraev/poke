package dispatch

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"poke/internal/server/executor"
	"poke/internal/server/request"
)

type SyncDispatcher struct {
	ctx       context.Context                // executor context
	registry  *CommandRegistry               // command registry
	reqCh     <-chan request.CommandRequest  // request input stream, executor routes them
	executors map[string]executor.ExecutorFn // worker input channels
	logger    *slog.Logger                   // dispatcher logger
}

// NewSyncDispatcher constructs a synchronous dispatcher for configured executors.
//
// SyncDispatcher executes commands one at a time, taking new requests from
// reqCh only after the previous one completes.
//
// Note that SyncDispatcher does not own reqCh.
func NewSyncDispatcher(ctx context.Context, registry *CommandRegistry, executors []string, reqCh <-chan request.CommandRequest) (*SyncDispatcher, error) {
	workerChs := make(map[string]executor.ExecutorFn, len(executors))

	for _, e := range executors {
		switch e {
		case "bin":
			workerChs["bin"] = executor.ExecuteBinary
		default:
			return nil, fmt.Errorf("invalid executor: %s", e)
		}
	}

	logger := slog.Default()
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return &SyncDispatcher{
		ctx:       ctx,
		registry:  registry,
		reqCh:     reqCh,
		executors: workerChs,
		logger:    logger.With("component", "dispatcher"),
	}, nil
}

// Run consumes requests and executes commands serially until context or channel closure.
func (d *SyncDispatcher) Run() {
	d.logger.Info("sync loop started", "event", "loop_started")
	for {
		select {
		case <-d.ctx.Done():
			d.logger.Info("context canceled, stopping", "event", "context_canceled")
			return
		case req, ok := <-d.reqCh:
			if !ok {
				d.logger.Info("request channel closed, stopping", "event", "request_channel_closed")
				return
			}
			d.logger.Info("request received", "event", "request_received", "command_id", req.CommandID)
			cmd, err := d.registry.Get(req.CommandID)
			if err != nil {
				d.logger.Warn("command lookup failed", "event", "command_lookup_failed", "command_id", req.CommandID, "error", err)
				continue
			}
			cmd.ID = req.CommandID
			fn, exists := d.executors[cmd.Executor]
			if !exists {
				d.logger.Warn("unknown executor", "event", "unknown_executor", "executor", cmd.Executor, "command_id", cmd.ID, "command_name", cmd.Name)
				continue
			}
			d.logger.Info("executing command", "event", "command_execution_started", "executor", cmd.Executor, "command_id", cmd.ID, "command_name", cmd.Name)
			result := fn(d.ctx, cmd)
			if result.Error != nil {
				d.logger.Error("command execution failed", "event", "command_execution_failed", "command_id", cmd.ID, "command_name", cmd.Name, "exit_code", result.ExitCode, "error", result.Error)
				continue
			}
			d.logger.Info("command execution completed", "event", "command_execution_completed", "command_id", cmd.ID, "command_name", cmd.Name, "exit_code", result.ExitCode)
		}
	}
}
