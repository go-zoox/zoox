package mq

import (
	"context"

	gomq "github.com/go-zoox/mq"
)

// MQ ...
type MQ interface {
	Send(topic string, message *gomq.Message) error
	Consume(ctx context.Context, topic string, group string, consumer string, start string, batchSize int, h gomq.Handler) error
}

type mq struct {
	ctx context.Context
	ps  gomq.MQ
}

// New creates a mq at the given context.
func New(ctx context.Context, ps gomq.MQ) MQ {
	return &mq{
		ctx: ctx,
		ps:  ps,
	}
}

func (p *mq) Send(topic string, message *gomq.Message) error {
	return p.ps.Send(p.ctx, message)
}

func (p *mq) Consume(ctx context.Context, topic string, group string, consumer string, start string, batchSize int, h gomq.Handler) error {
	return p.ps.Consume(p.ctx, topic, group, consumer, start, batchSize, h)
}
