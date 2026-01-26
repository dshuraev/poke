package executor

import "context"

type ExecutorFn func(context.Context, Command) Result
