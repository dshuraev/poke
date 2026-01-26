package dispatch

import (
	"context"
	"fmt"
	"log"
	"poke/internal/server/executor"
	"poke/internal/server/request"
)

type SyncDispatcher struct {
	ctx       context.Context                // executor context
	registry  *CommandRegistry               // command registry
	reqCh     <-chan request.CommandRequest  // request input stream, executor routes them
	executors map[string]executor.ExecutorFn // worker input channels
}

// Create new synchronout dispatcher for specified executor types.
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

	return &SyncDispatcher{
		ctx:       ctx,
		registry:  registry,
		reqCh:     reqCh,
		executors: workerChs,
	}, nil
}

func (d *SyncDispatcher) Run() {
	log.Printf("dispatcher: sync loop started")
	for {
		select {
		case <-d.ctx.Done():
			log.Printf("dispatcher: context canceled, stopping")
			return
		case req, ok := <-d.reqCh:
			if !ok {
				log.Printf("dispatcher: request channel closed, stopping")
				return
			}
			log.Printf("dispatcher: request received command_id=%s", req.CommandID)
			cmd, err := d.registry.Get(req.CommandID)
			if err != nil {
				log.Printf("executor: request for command %s failed: %v", req.CommandID, err)
				continue
			}
			cmd.ID = req.CommandID
			fn, exists := d.executors[cmd.Executor]
			if !exists {
				log.Printf("executor: unknown executor %q for command %s[%s]", cmd.Executor, cmd.ID, cmd.Name)
				continue
			}
			log.Printf("executor: executing command %s[%s] via %s", cmd.ID, cmd.Name, cmd.Executor)
			result := fn(d.ctx, cmd)
			if result.Error != nil {
				log.Printf("executor: command %s[%s] failed with exit %d: %v", cmd.ID, cmd.Name, result.ExitCode, result.Error)
				continue
			}
			log.Printf("executor: command %s[%s] completed with exit %d", cmd.ID, cmd.Name, result.ExitCode)
		}
	}
}
