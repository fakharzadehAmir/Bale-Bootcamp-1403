package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"
	"time"

	"github.com/opentracing/opentracing-go"
)

type Queue struct {
	sync.RWMutex
	expiredMsgId   chan int
	Subs           map[int]*Subscriber
	newSub         chan *Subscriber
	publishedMsg   chan *broker.Message
	msgList        map[int]*broker.Message
	expiredMsgList map[int]bool
}

func CreateQueue(subject string, ctx context.Context) *Queue {
	span, _ := opentracing.StartSpanFromContext(ctx, "CreateQueue")
	defer span.Finish()

	queue := &Queue{
		publishedMsg:   make(chan *broker.Message, 1),
		newSub:         make(chan *Subscriber, 1),
		Subs:           make(map[int]*Subscriber),
		msgList:        make(map[int]*broker.Message),
		expiredMsgList: make(map[int]bool),
		expiredMsgId:   make(chan int),
	}
	go queue.handleNewSignals()
	return queue
}

func (q *Queue) handleNewSignals() {
	for {
		select {
		case expiredMsgId := <-q.expiredMsgId:
			q.Lock()
			q.expiredMsgList[expiredMsgId] = true
			q.Unlock()

		case addedSubscriber := <-q.newSub:
			waitGroup := &sync.WaitGroup{}
			q.Lock()
			newId := addedSubscriber.id
			q.Subs[newId] = addedSubscriber
			q.Unlock()
			q.RLock()
			for id, msg := range q.msgList {
				_, exist := q.expiredMsgList[id]
				if !exist {
					waitGroup.Add(1)
					go func(msg broker.Message) {
						defer waitGroup.Done()
						addedSubscriber.chanMsg <- msg
					}(*msg)
				}
			}
			waitGroup.Wait()
			q.RUnlock()
		case newPublishedMsg := <-q.publishedMsg:
			waitGroup := &sync.WaitGroup{}
			q.RLock()
			for _, sub := range q.Subs {
				waitGroup.Add(1)
				go func(sub *Subscriber, pubMsg broker.Message) {
					defer waitGroup.Done()
					sub.chanMsg <- pubMsg
				}(sub, *newPublishedMsg)
			}
			waitGroup.Wait()
			q.RUnlock()
		}
	}
}

func (q *Queue) AddSubscriber(ctx context.Context) chan broker.Message {
	span, _ := opentracing.StartSpanFromContext(ctx, "AddSubscriber method in Queue")
	defer span.Finish()

	newChan := make(chan broker.Message, 100)
	newSub := &Subscriber{
		id:      idGenerator.NewID(),
		chanMsg: newChan,
	}
	q.newSub <- newSub
	return newChan
}

func (q *Queue) PublishMessage(msg *broker.Message, ctx context.Context) int {
	span, _ := opentracing.StartSpanFromContext(ctx, "PublishMessage method in Queue")
	defer span.Finish()
	var newMsgId = idGenerator.NewID()
	if msg.Expiration != 0 {
		q.Lock()
		q.msgList[newMsgId] = msg
		q.Unlock()
		go func(expiredMsg *broker.Message, id int) {
			afterTime := time.After(expiredMsg.Expiration * time.Second)
			for {
				select {
				case <-afterTime:
					q.expiredMsgId <- id
					return
				}
			}
		}(msg, newMsgId)
	}
	q.publishedMsg <- msg
	return newMsgId
}

func (q *Queue) GetMessageByID(id int, ctx context.Context) (broker.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetMessageByID method in Queue")
	defer span.Finish()

	q.RLock()
	defer q.RUnlock()
	msg, exist := q.msgList[id]
	if !exist {
		return broker.Message{}, broker.ErrInvalidID
	}

	q.RLock()
	_, exist = q.expiredMsgList[id]
	if exist {
		return broker.Message{}, broker.ErrExpiredID
	}
	q.RUnlock()
	return *msg, nil
}
