package propagators

import (
	"context"

	uuid "github.com/google/uuid"
	otel_propagation "go.opentelemetry.io/otel/propagation"
)

// context key
type correlationIdContextKeyType int

const currentCorrelationIdContextKey correlationIdContextKeyType = iota
const currentRequestIdContextKey correlationIdContextKeyType = iota

type (
	CorrelationPropagator struct{}
)

func EnsureCorrelationId(ctx context.Context) context.Context {
	_, ok := ctx.Value(currentCorrelationIdContextKey).(string)
	if !ok {
		correlationID := uuid.New().String()
		ctx = context.WithValue(ctx, currentCorrelationIdContextKey, correlationID)
	}
	return ctx
}

func GetCorrelationID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(currentCorrelationIdContextKey).(string)
	return value, ok
}

func NewCorrelationPropagator() *CorrelationPropagator {
	return &CorrelationPropagator{}
}

func (c *CorrelationPropagator) Inject(ctx context.Context, carrier interface{}) {

	textMapCarrier := carrier.(otel_propagation.TextMapCarrier)

	// try extracting the correlation id from the context
	value, ok := ctx.Value(currentCorrelationIdContextKey).(string)
	if ok {
		textMapCarrier.Set(LogCorrelationIDName, value)
	}
	value, ok = ctx.Value(currentRequestIdContextKey).(string)
	if ok {
		textMapCarrier.Set(LogRequestIDName, value)
	}
}

func (c *CorrelationPropagator) Extract(ctx context.Context, carrier interface{}) context.Context {
	textMapCarrier := carrier.(otel_propagation.TextMapCarrier)

	// extract the correlation id from the carrier
	correlationId := textMapCarrier.Get(LogCorrelationIDName)
	requestId := textMapCarrier.Get(LogRequestIDName)

	// store the correlation id in the context
	ctx = context.WithValue(ctx, currentCorrelationIdContextKey, correlationId)
	ctx = context.WithValue(ctx, currentRequestIdContextKey, requestId)
	return ctx
}
