package pubsub

import (
	"context"

	gopubsub "github.com/go-zoox/pubsub"
)

// PubSub ...
type PubSub interface {
	Publish(topic string, message *gopubsub.Message) error
	Subscribe(topic string, handler gopubsub.Handler) error
}

type pubsub struct {
	ctx context.Context
	ps  gopubsub.PubSub
}

// New creates a pubsub at the given context.
func New(ctx context.Context, ps gopubsub.PubSub) PubSub {
	return &pubsub{
		ctx: ctx,
		ps:  ps,
	}
}

func (p *pubsub) Publish(topic string, message *gopubsub.Message) error {
	return p.ps.Publish(p.ctx, message)
}

func (p *pubsub) Subscribe(topic string, handler gopubsub.Handler) error {
	return p.ps.Subscribe(p.ctx, topic, handler)
}
