package otel

import (
	"context"
	contracts_nats "natsauth/internal/contracts/nats"
	propagators "natsauth/internal/propagators"

	services_consumer_wrapper "natsauth/internal/services/consumer_wrapper"
	services_jetstream_wrapper "natsauth/internal/services/jetstream_wrapper"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	nats "github.com/nats-io/nats.go"
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
	zerolog "github.com/rs/zerolog"
	otel "go.opentelemetry.io/otel"
	otel_attribute "go.opentelemetry.io/otel/attribute"
	otel_trace "go.opentelemetry.io/otel/trace"
)

type (
	service struct {
		config                *contracts_nats.TraceProviderConfig
		correlationPropagator *propagators.CorrelationPropagator
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
		config:                config,
		correlationPropagator: propagators.NewCorrelationPropagator(),
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

func (s *service) ContextFromMessage(msg nats_jetstream.Msg) context.Context {
	propagator := otel.GetTextMapPropagator()
	ctx := context.Background()
	carrier := services_consumer_wrapper.NewNATSMessageCarrier(msg)
	ctx = propagator.Extract(ctx, carrier)
	ctx = s.correlationPropagator.Extract(ctx, carrier)
	// Get the current span context
	spanContext := otel_trace.SpanContextFromContext(ctx)

	logContext := zerolog.Ctx(ctx).With().
		Str("trace_id", spanContext.TraceID().String()).
		Str("span_id", spanContext.SpanID().String())

	correlationID, ok := propagators.GetCorrelationID(ctx)
	if ok {
		logContext = logContext.Str("correlation_id", correlationID)
	}
	log := logContext.Logger()

	ctx = log.WithContext(ctx)
	return ctx
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
	carrier := services_jetstream_wrapper.NewNATSMessageCarrier(request.Msg)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	s.correlationPropagator.Inject(ctx, carrier)

	// Get the current span context
	spanContext := otel_trace.SpanContextFromContext(ctx)

	log := zerolog.Ctx(ctx).With().
		Str("trace_id", spanContext.TraceID().String()).
		Str("span_id", spanContext.SpanID().String()).
		Logger()
	ctx = log.WithContext(ctx)
	return &contracts_nats.StartNewSpanResponse{
		Span:    &spanCarrier{span: span},
		Context: ctx,
	}, nil
}
