package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
	"github.com/jamesstocktonj1/componentize-sdk/otel/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

var tracer = otel.Tracer("basic-example")

func init() {
	otel.SetTracerProvider(tracesdk.NewTracerProvider(
		tracesdk.WithSpanProcessor(trace.NewSpanProcessor()),
	))

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/echo", echo)
	mux.HandleFunc("/greet", greeting)

	wasihttp.Handle(mux)
}

func hello(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "hello")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.host", r.Host),
	)
	span.AddEvent("some event")

	otherFunc(ctx)

	fmt.Fprintln(w, "Hello, World!")
}

func otherFunc(ctx context.Context) {
	_, span := tracer.Start(ctx, "otherFunc")
	defer span.End()

	span.SetStatus(codes.Ok, "")
}

func echo(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "echo")
	defer span.End()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func greeting(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "greeting")
	defer span.End()

	request := struct {
		Name string `json:"name"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Hello, %s!", request.Name)
}

func main() {}
