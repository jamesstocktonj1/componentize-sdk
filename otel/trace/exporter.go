package trace

import (
	"context"

	tracing "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_otel_tracing"
	"go.opentelemetry.io/otel/sdk/trace"
)

func NewSpanProcessor() trace.SpanProcessor {
	return &wasiSpanProcessor{}
}

type wasiSpanProcessor struct {
}

var _ trace.SpanProcessor = (*wasiSpanProcessor)(nil)

func (sp *wasiSpanProcessor) OnStart(parent context.Context, s trace.ReadWriteSpan) {
	var sc tracing.SpanContext
	if s == nil {
		sc = tracing.OuterSpanContext()
	} else {
		sc = mapSpanContext(s.SpanContext())
	}
	tracing.OnStart(sc)
}

func (sp *wasiSpanProcessor) OnEnd(s trace.ReadOnlySpan) {
	tracing.OnEnd(mapSpanData(s))
}

func (sp *wasiSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

func (sp *wasiSpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}
