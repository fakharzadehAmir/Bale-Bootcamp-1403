package database

import (
	"context"
	"therealbroker/pkg/broker"
)

type DB interface {
	AddMessage(ctx context.Context, msg broker.Message, subject string) (int, error)
	FetchMessage(ctx context.Context, id int, subject string) (broker.Message, error)
	DeleteMessage(subject string, id int)
	GetMessagesBySubject(ctx context.Context, subject string) ([]broker.Message, error)
	Close() error
}
