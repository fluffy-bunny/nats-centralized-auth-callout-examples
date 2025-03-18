package jetstream_wrapper

import (
	nats "github.com/nats-io/nats.go"
	otel_propagation "go.opentelemetry.io/otel/propagation"
	ddtrace_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type NATSMessageCarrier struct {
	msg *nats.Msg
}

var stemNATSMessageCarrier = (*NATSMessageCarrier)(nil)
var _ ddtrace_tracer.TextMapReader = stemNATSMessageCarrier
var _ ddtrace_tracer.TextMapWriter = stemNATSMessageCarrier
var _ otel_propagation.TextMapCarrier = stemNATSMessageCarrier

func NewNATSMessageCarrier(msg *nats.Msg) NATSMessageCarrier {
	return NATSMessageCarrier{msg: msg}
}

func (c NATSMessageCarrier) Get(key string) string {
	return c.msg.Header.Get(key)
}

func (c NATSMessageCarrier) Set(key, val string) {
	c.msg.Header.Set(key, val)
}

func (c NATSMessageCarrier) Keys() []string {
	keys := make([]string, len(c.msg.Header))
	i := 0
	for k := range c.msg.Header {
		keys[i] = k
		i++
	}
	return keys
}
func (c NATSMessageCarrier) ForeachKey(handler func(key string, val string) error) error {
	for _, key := range c.Keys() {
		value := c.msg.Header.Get(key)
		if err := handler(key, value); err != nil {
			return err
		}
	}
	return nil
}
