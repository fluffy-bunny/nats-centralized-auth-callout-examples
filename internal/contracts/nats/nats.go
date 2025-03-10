package nats

import (
	"context"

	nats_go "github.com/nats-io/nats.go"
)

type (
	NATSConnConfig struct {
		NatsUrl  string
		Username string
		Password string
	}
	INATSConnection interface {
		Conn(ctx context.Context) (*nats_go.Conn, error)
	}
)
