package broker

import (
	"context"
	"errors"
	"sync"
	"therealbroker/config"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/database"
	"time"

	"github.com/opentracing/opentracing-go"
)

const (
	POSTGRES   = "POSTGRES"
	CASSANDRA  = "CASSANDRA"
	GOLANG_MAP = "NOT_PERSISTED"
)

type Subscriber struct {
	channMsg chan broker.Message
}

type Queue struct {
	queueName string
	subs      []Subscriber
}

type Module struct {
	// TODO: Add required fields
	queue       map[string]Queue
	messages    map[int]broker.Message
	closed      bool
	pgDb        database.DB
	cassandraDb database.DB
	storageType string
	sync.RWMutex
}

func NewModule() broker.Broker {
	return &Module{
		queue:       make(map[string]Queue),
		messages:    make(map[int]broker.Message),
		storageType: config.GetConfigInstance().Broker.StorageType,
		pgDb:        database.GetDatabaseInstance(),
		cassandraDb: database.GetCassandraInstance(),
	}
}

func (m *Module) Close() error {
	if m.closed {
		return broker.ErrUnavailable
	}
	m.closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if m.closed {
		return -1, broker.ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return -1, ctx.Err()
	default:

		//	Send new published message to subscribers
		sendSpan, _ := opentracing.StartSpanFromContext(ctx, "Send Published Message to Subscribers")
		for _, sub := range m.queue[subject].subs {
			if len(sub.channMsg) != cap(sub.channMsg) {
				sub.channMsg <- msg
			}
		}
		sendSpan.Finish()

		//	Store new message
		storeSpan, storeCtx := opentracing.StartSpanFromContext(ctx, "Store Published Message")
		var newMsgId int
		var errInsertMsg error
		switch m.storageType {
		case GOLANG_MAP:
			m.Lock()
			newMsgId = len(m.messages)
			m.messages[newMsgId] = msg
			m.Unlock()
		case POSTGRES:
			newMsgId, errInsertMsg = m.pgDb.AddMessage(storeCtx, msg, subject)
			if errInsertMsg != nil {
				return -1, errInsertMsg
			}
		case CASSANDRA:
			newMsgId, _ = m.cassandraDb.AddMessage(storeCtx, msg, subject)
		}
		storeSpan.Finish()

		//	Check Expiration
		if msg.Expiration != 0 {
			go func(msgId int, subjct string, ticker time.Duration) {

				expireTime := time.NewTicker(ticker * time.Second)
				defer expireTime.Stop()

				<-expireTime.C
				switch m.storageType {
				case GOLANG_MAP:
					m.Lock()
					delete(m.messages, msgId)
					m.Unlock()
				case POSTGRES:
					m.pgDb.DeleteMessage(subjct, msgId)
				case CASSANDRA:
					m.cassandraDb.DeleteMessage(subjct, msgId)
				}

			}(newMsgId, subject, msg.Expiration)
		}

		return newMsgId, nil
	}

}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {

	if m.closed {
		return nil, broker.ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		subSpan, _ := opentracing.StartSpanFromContext(ctx, "Add new Subscriber")
		chanMsg := make(chan broker.Message, 200)
		m.Lock()
		m.queue[subject] = Queue{
			queueName: subject,
			subs:      append(m.queue[subject].subs, Subscriber{channMsg: chanMsg}),
		}
		m.Unlock()
		subSpan.Finish()

		go func(ctx context.Context, channel chan broker.Message, subj string) {
			switch m.storageType {
			case POSTGRES:
				messages, _ := m.pgDb.GetMessagesBySubject(ctx, subj)
				for _, msg := range messages {
					channel <- msg
				}
			case CASSANDRA:
				messages, _ := m.cassandraDb.GetMessagesBySubject(ctx, subj)
				for _, msg := range messages {
					channel <- msg
				}
			}
		}(ctx, chanMsg, subject)

		return chanMsg, nil
	}
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	if m.closed {
		return broker.Message{}, broker.ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return broker.Message{}, ctx.Err()
	default:

		retrieveSpan, retrieveCtx := opentracing.StartSpanFromContext(ctx, "Retrieve message in fetch method Broker Module")

		var msg broker.Message
		var errRetrieving error
		var messageExist bool

		switch m.storageType {

		case GOLANG_MAP:
			m.RLock()
			msg, messageExist = m.messages[id]
			m.RUnlock()

			if !messageExist {
				return broker.Message{}, errors.New(broker.ErrInvalidID.Error() + " or " + broker.ErrExpiredID.Error())
			}
		case POSTGRES:
			msg, errRetrieving = m.pgDb.FetchMessage(retrieveCtx, id, subject)
			if errRetrieving != nil {
				return broker.Message{}, errRetrieving
			}
		case CASSANDRA:
			msg, errRetrieving = m.cassandraDb.FetchMessage(retrieveCtx, id, subject)
			if errRetrieving != nil {
				return broker.Message{}, errRetrieving
			}
		}
		retrieveSpan.Finish()

		return msg, nil

	}
}
