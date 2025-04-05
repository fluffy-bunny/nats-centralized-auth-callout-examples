package consumer_wrapper

import (
	nats_jetstream "github.com/nats-io/nats.go/jetstream"
	otel_propagation "go.opentelemetry.io/otel/propagation"
	ddtrace_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type NATSMessageCarrier struct {
	msg nats_jetstream.Msg
}

var stemNATSMessageCarrier = (*NATSMessageCarrier)(nil)
var _ ddtrace_tracer.TextMapReader = stemNATSMessageCarrier
var _ ddtrace_tracer.TextMapWriter = stemNATSMessageCarrier
var _ otel_propagation.TextMapCarrier = stemNATSMessageCarrier

func NewNATSMessageCarrier(msg nats_jetstream.Msg) NATSMessageCarrier {
	return NATSMessageCarrier{msg: msg}
}

func (c NATSMessageCarrier) Get(key string) string {
	headers := c.msg.Headers()
	return headers.Get(key)
}

func (c NATSMessageCarrier) Set(key, val string) {
	headers := c.msg.Headers()

	headers.Set(key, val)
}

func (c NATSMessageCarrier) Keys() []string {
	headers := c.msg.Headers()

	keys := make([]string, len(headers))
	i := 0
	for k := range headers {
		keys[i] = k
		i++
	}
	return keys
}
func (c NATSMessageCarrier) ForeachKey(handler func(key string, val string) error) error {
	for _, key := range c.Keys() {
		headers := c.msg.Headers()
		value := headers.Get(key)
		if err := handler(key, value); err != nil {
			return err
		}
	}
	return nil
}
