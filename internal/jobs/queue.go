package jobs

import "context"

type Handler func(ctx context.Context, msg Message) error

type Queue interface {
	Enqueue(ctx context.Context, msg Message) error
	Consume(ctx context.Context, handler Handler) error
	Ping(ctx context.Context) error
	Close() error
}
