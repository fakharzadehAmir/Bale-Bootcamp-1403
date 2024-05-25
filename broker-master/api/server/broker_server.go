package server

import (
	"context"
	"sync"
	"therealbroker/api/proto"
	brokerModule "therealbroker/internal/broker"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/middleware"
	"time"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImplementedBrokerServer struct {
	proto.UnimplementedBrokerServer
	broker broker.Broker
}

func NewImplementedServer() proto.BrokerServer {
	return &ImplementedBrokerServer{
		broker: brokerModule.NewModule(),
	}
}

func (s ImplementedBrokerServer) Publish(ctx context.Context, request *proto.PublishRequest) (*proto.PublishResponse, error) {

	span, err := middleware.StartSpanFromGRPC(ctx, "Publish gRPC Broker Server")
	if err != nil {
		return nil, err
	}
	spanCtx := opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()
	startTime := time.Now()
	defer middleware.MethodDuration.WithLabelValues("publish").Observe(float64(time.Since(startTime).Microseconds()))
	publishedMessage := broker.Message{
		Body:       string(request.GetBody()),
		Expiration: time.Duration(request.GetExpirationSeconds()),
	}

	msgId, err := s.broker.Publish(spanCtx, request.GetSubject(), publishedMessage)
	if err != nil {

		middleware.MethodCount.WithLabelValues("publish", "failed").Observe(float64(time.Since(startTime)))
		return nil, status.Errorf(codes.Unavailable, "Broker is closed")

	}

	reponse := &proto.PublishResponse{Id: int32(msgId)}
	middleware.MethodCount.WithLabelValues("publish", "successful").Observe(float64(time.Since(startTime)))

	return reponse, nil
}

func (s ImplementedBrokerServer) Subscribe(request *proto.SubscribeRequest, stream proto.Broker_SubscribeServer) error {
	span, err := middleware.StartSpanFromGRPC(stream.Context(), "Subscribe gRPC Broker Server")
	if err != nil {
		return err
	}
	spanCtx := opentracing.ContextWithSpan(stream.Context(), span)
	defer span.Finish()

	startTime := time.Now()
	defer middleware.MethodDuration.WithLabelValues("subscirbe").Observe(float64(time.Since(startTime).Microseconds()))

	var subErr error
	middleware.ActiveSubscribers.Inc()

	messageChan, err := s.broker.Subscribe(spanCtx, request.GetSubject())
	if err != nil {
		middleware.MethodCount.WithLabelValues("subscribe", "failed").Observe(float64(time.Since(startTime)))
		return status.Errorf(codes.Unavailable, "Broker is closed ")
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(ctx context.Context) {
		defer func() {
			middleware.ActiveSubscribers.Dec()
			wg.Done()
		}()

		for {
			select {
			case msg, ok := <-messageChan:
				if !ok {
					return
				}
				go func(m broker.Message) {
					if err := stream.Send(&(proto.MessageResponse{Body: []byte(m.Body)})); err != nil {
						subErr = err
					}
				}(msg)
			case <-ctx.Done():
				return
			}
		}
	}(stream.Context())
	wg.Wait()
	if subErr != nil {
		middleware.MethodCount.WithLabelValues("subscribe", "failed").Observe(float64(time.Since(startTime)))
	} else {
		middleware.MethodCount.WithLabelValues("subscribe", "successful").Observe(float64(time.Since(startTime)))
	}

	return subErr
}

func (s ImplementedBrokerServer) Fetch(ctx context.Context, request *proto.FetchRequest) (*proto.MessageResponse, error) {
	span, err := middleware.StartSpanFromGRPC(ctx, "Fetch gRPC Broker Server")
	if err != nil {
		return nil, err
	}
	spanCtx := opentracing.ContextWithSpan(ctx, span)

	defer span.Finish()

	startTime := time.Now()
	defer middleware.MethodDuration.WithLabelValues("fetch").Observe(float64(time.Since(startTime).Microseconds()))

	message, err := s.broker.Fetch(spanCtx, request.GetSubject(), int(request.GetId()))
	if err != nil {
		middleware.MethodCount.WithLabelValues("fetch", "failed").Observe(float64(time.Since(startTime)))
		switch err {
		case broker.ErrUnavailable:
			return nil, status.Errorf(codes.Unavailable, "Broker is closed")
		case broker.ErrExpiredID:
			return nil, status.Errorf(codes.InvalidArgument, "Expired Message")
		case broker.ErrInvalidID:
			return nil, status.Errorf(codes.InvalidArgument, "Invalid ID")
		}
	}
	response := &proto.MessageResponse{Body: []byte(message.Body)}

	middleware.MethodCount.WithLabelValues("fetch", "successful").Observe(float64(time.Since(startTime)))
	return response, nil

}
