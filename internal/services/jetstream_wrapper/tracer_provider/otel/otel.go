package otel

import (
	"context"
	contracts_nats "natsauth/internal/contracts/nats"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	"github.com/nats-io/nats.go"
	otel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	otel_attribute "go.opentelemetry.io/otel/attribute"
	otel_trace "go.opentelemetry.io/otel/trace"
)

type (
	service struct {
		config *contracts_nats.TraceProviderConfig
	}
	spanCarrier struct {
		span otel_trace.Span
	}
)

var stemService = (*service)(nil)
var _ contracts_nats.ITracerProvider = (*service)(nil)
var _ contracts_nats.ISpan = (*spanCarrier)(nil)

func (s *service) Ctor(config *contracts_nats.TraceProviderConfig) (contracts_nats.ITracerProvider, error) {
	return &service{
		config: config,
	}, nil
}
func AddSingletonITracerProvider(builder di.ContainerBuilder) {
	di.AddSingleton[contracts_nats.ITracerProvider](
		builder,
		stemService.Ctor,
	)
}
func (s *spanCarrier) Finish() {
	if s.span != nil {
		s.span.End()
	}
}
func (s *service) StartNewSpan(ctx context.Context, request *contracts_nats.StartNewSpanRequest) (*contracts_nats.StartNewSpanResponse, error) {
	var span otel_trace.Span

	tracer := otel.Tracer(s.config.AppName)
	ctx, span = tracer.Start(ctx, request.OperationName)

	for k, v := range request.Tags {
		span.SetAttributes(otel_attribute.String(k, v))
	}
	if request.Msg.Header == nil {
		request.Msg.Header = make(nats.Header)
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Msg.Header))

	return &contracts_nats.StartNewSpanResponse{
		Span:    &spanCarrier{span: span},
		Context: ctx,
	}, nil
}
