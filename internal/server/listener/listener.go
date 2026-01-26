package listener

import (
	"context"
	"poke/internal/server/request"
)

type RequestSource[T any] interface {
	Listen(ctx context.Context, cfg T, ch chan<- request.CommandRequest)
}
