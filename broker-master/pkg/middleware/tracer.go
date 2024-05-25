package middleware

import (
	"context"
	"errors"
	"io"
	"strconv"
	"therealbroker/config"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc/metadata"
)

var Tracer opentracing.Tracer

func NewJaegerObject(cfg config.Config, logger *logrus.Logger) (io.Closer, error) {
	jagerCfg := jaegercfg.Configuration{
		ServiceName: cfg.Jaeger.ServiceName,
		RPCMetrics:  true,
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: cfg.Jaeger.Host + ":" + strconv.Itoa(cfg.Jaeger.Port),
		},
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}
	tracer, closer, err := jagerCfg.NewTracer(jaegercfg.Logger(&LogrusJaegerAdapter{logger: logger}))
	Tracer = tracer
	return closer, err
}

func StartSpanFromGRPC(ctx context.Context, operationName string) (opentracing.Span, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("could not retrieve metadata from context")
	}

	spanContext, err := Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, err
	}

	return Tracer.StartSpan(operationName, ext.RPCServerOption(spanContext)), nil
}
