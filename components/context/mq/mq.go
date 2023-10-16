package mq

import (
	"context"

	gomq "github.com/go-zoox/mq"
)

// MQ ...
type MQ interface {
	Send(topic string, message *gomq.Msg) error
	Consume(topic string, handler gomq.Handler) error
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

func (p *mq) Send(topic string, message *gomq.Msg) error {
	return p.ps.Send(p.ctx, message)
}

func (p *mq) Consume(topic string, handler gomq.Handler) error {
	return p.ps.Consume(p.ctx, topic, handler)
}
