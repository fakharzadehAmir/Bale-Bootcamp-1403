package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"

	"github.com/opentracing/opentracing-go"
)

type Module struct {
	// TODO: Add required fields
	queueList map[string]*Queue
	closed    bool
	sync.RWMutex
}

func NewModule() broker.Broker {
	return &Module{
		queueList: make(map[string]*Queue),
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
	span, _ := opentracing.StartSpanFromContext(ctx, "Publish method in Broker Module")
	spanCtx := opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	if m.closed {
		return -1, broker.ErrUnavailable
	}
	queue, exist := m.getQueue(subject)
	if !exist {
		queue = CreateQueue(subject, spanCtx)
		m.Lock()
		m.queueList[subject] = queue
		m.Unlock()

	}
	id := queue.PublishMessage(&msg, spanCtx)
	return id, nil

}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Subscribe method in Broker Module")
	spanCtx := opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	if m.closed {
		return nil, broker.ErrUnavailable
	}
	queue, exist := m.getQueue(subject)
	if !exist {
		queue = CreateQueue(subject, spanCtx)
		m.Lock()
		m.queueList[subject] = queue
		m.Unlock()
	}
	ch := queue.AddSubscriber(spanCtx)
	return ch, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Fetch method in Broker Module")
	spanCtx := opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	if m.closed {
		return broker.Message{}, broker.ErrUnavailable
	}
	queue, exist := m.getQueue(subject)
	if !exist {
		queue = CreateQueue(subject, spanCtx)
		m.Lock()
		m.queueList[subject] = queue
		m.Unlock()
	}

	retrievedMsg, err := queue.GetMessageByID(id, spanCtx)
	if err != nil {
		return broker.Message{}, err
	}
	return retrievedMsg, nil
}

func (m *Module) getQueue(subject string) (*Queue, bool) {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()
	queue, exist := m.queueList[subject]
	return queue, exist
}

// func (m *Module) signalBasedProcessing(queue *Queue) {
// 	for {
// 		select {
// 		case expiredMsgId := <-queue.expiredMsgId:
// 			queue.Lock()
// 			queue.expiredMsgList[expiredMsgId] = true
// 			queue.Unlock()

// 		case addedSubscriber := <-queue.newSub:
// 			waitGroup := &sync.WaitGroup{}
// 			queue.Lock()
// 			newId := addedSubscriber.id
// 			queue.Subs[newId] = addedSubscriber
// 			queue.Unlock()
// 			queue.RLock()
// 			for _, msg := range queue.msgList {
// 				waitGroup.Add(1)
// 				go func(msg broker.Message) {
// 					defer waitGroup.Done()
// 					addedSubscriber.chanMsg <- msg
// 				}(*msg)
// 			}
// 			waitGroup.Wait()
// 			queue.RUnlock()
// 		case newPublishedMsg := <-queue.publishedMsg:
// 			waitGroup := &sync.WaitGroup{}
// 			queue.RLock()
// 			for _, sub := range queue.Subs {
// 				waitGroup.Add(1)
// 				go func(sub *Subscriber, pubMsg broker.Message) {
// 					defer waitGroup.Done()
// 					sub.chanMsg <- pubMsg
// 				}(sub, *newPublishedMsg)
// 			}
// 			waitGroup.Wait()
// 			queue.RUnlock()
// 		}
// 	}
// }
