package trace

import (
	"github.com/bytecodealliance/wit-bindgen/wit_types"
	clock "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_clocks_wall_clock"
	tracing "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_otel_tracing"
	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_otel_types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func mapSpanData(sp tracesdk.ReadOnlySpan) tracing.SpanData {
	return tracing.SpanData{
		SpanContext:  mapSpanContext(sp.SpanContext()),
		ParentSpanId: sp.Parent().SpanID().String(),
		SpanKind:     mapSpanKind(sp.SpanKind()),
		Name:         sp.Name(),
		StartTime: clock.Datetime{
			Seconds:     uint64(sp.StartTime().Unix()),
			Nanoseconds: uint32(sp.StartTime().UnixNano()),
		},
		EndTime: clock.Datetime{
			Seconds:     uint64(sp.EndTime().Unix()),
			Nanoseconds: uint32(sp.EndTime().UnixNano()),
		},
		Attributes:           mapAttributes(sp.Attributes()),
		Events:               mapEvents(sp.Events()),
		Links:                mapLinks(sp.Links()),
		Status:               mapStatus(sp.Status()),
		InstrumentationScope: mapInstrumentationScope(sp.InstrumentationScope()),
		DroppedAttributes:    uint32(sp.DroppedAttributes()),
		DroppedEvents:        uint32(sp.DroppedEvents()),
		DroppedLinks:         uint32(sp.DroppedLinks()),
	}
}

func mapSpanContext(sc trace.SpanContext) tracing.SpanContext {
	return tracing.SpanContext{
		TraceId:    sc.TraceID().String(),
		SpanId:     sc.SpanID().String(),
		TraceFlags: tracing.TraceFlags(sc.TraceFlags()),
		IsRemote:   sc.IsRemote(),
		// TraceState: TODO,
	}
}

func mapSpanKind(sk trace.SpanKind) tracing.SpanKind {
	switch sk {
	case trace.SpanKindInternal:
		return tracing.SpanKindInternal
	case trace.SpanKindServer:
		return tracing.SpanKindServer
	case trace.SpanKindClient:
		return tracing.SpanKindClient
	case trace.SpanKindProducer:
		return tracing.SpanKindProducer
	case trace.SpanKindConsumer:
		return tracing.SpanKindConsumer
	default:
		// SpanKindUnspecified not in wit, as per otel docs this should default to SpanKindInternal
		return tracing.SpanKindInternal
	}
}

func mapAttributes(a []attribute.KeyValue) []types.KeyValue {
	attrs := make([]types.KeyValue, len(a))
	for i, attr := range a {
		attrs[i] = types.KeyValue{
			Key:   string(attr.Key),
			Value: attr.Value.AsString(),
		}
	}
	return attrs
}

func mapEvents(e []tracesdk.Event) []tracing.Event {
	evt := make([]tracing.Event, len(e))
	for i, event := range e {
		evt[i] = tracing.Event{
			Name: event.Name,
			Time: clock.Datetime{
				Seconds:     uint64(event.Time.Unix()),
				Nanoseconds: uint32(event.Time.UnixNano()),
			},
			Attributes: mapAttributes(event.Attributes),
		}
	}
	return evt
}

func mapLinks(l []tracesdk.Link) []tracing.Link {
	links := make([]tracing.Link, len(l))
	for i, link := range l {
		links[i] = tracing.Link{
			SpanContext: mapSpanContext(link.SpanContext),
			Attributes:  mapAttributes(link.Attributes),
		}
	}
	return links
}

func mapStatus(st tracesdk.Status) tracing.Status {
	switch st.Code {
	case codes.Error:
		return tracing.MakeStatusError(st.Description)
	case codes.Ok:
		return tracing.MakeStatusOk()
	default:
		return tracing.MakeStatusUnset()
	}
}

func mapInstrumentationScope(sc instrumentation.Scope) types.InstrumentationScope {
	scope := types.InstrumentationScope{
		Name:       sc.Name,
		Attributes: mapAttributes(sc.Attributes.ToSlice()),
	}

	if sc.Version != "" {
		scope.Version = wit_types.Some(sc.Version)
	}

	if sc.SchemaURL != "" {
		scope.SchemaUrl = wit_types.Some(sc.SchemaURL)
	}
	return scope
}
