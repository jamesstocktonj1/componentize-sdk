package httptypes

import (
	"net/http"
	"testing"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// mockFieldsSetter implements FieldsSetter for testing.
type mockFieldsSetter struct {
	headers map[string][][]uint8
	// errOnKey causes Set to return an error when this key is set.
	errOnKey string
}

func newMockFieldsSetter() *mockFieldsSetter {
	return &mockFieldsSetter{headers: make(map[string][][]uint8)}
}

func (m *mockFieldsSetter) Set(name string, value [][]uint8) witTypes.Result[witTypes.Unit, types.HeaderError] {
	if m.errOnKey != "" && name == m.errOnKey {
		return witTypes.Err[witTypes.Unit, types.HeaderError](types.MakeHeaderErrorInvalidSyntax())
	}
	m.headers[name] = value
	return witTypes.Ok[witTypes.Unit, types.HeaderError](witTypes.Unit{})
}

func TestMapHttpHeaderTo_EmptyHeader(t *testing.T) {
	mock := newMockFieldsSetter()
	if err := mapHttpHeaderTo(http.Header{}, mock); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.headers) != 0 {
		t.Errorf("expected no headers set, got %v", mock.headers)
	}
}

func TestMapHttpHeaderTo_SingleHeader(t *testing.T) {
	mock := newMockFieldsSetter()
	h := http.Header{"Content-Type": {"application/json"}}

	if err := mapHttpHeaderTo(h, mock); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vals, ok := mock.headers["Content-Type"]
	if !ok {
		t.Fatal("Content-Type header not set")
	}
	if len(vals) != 1 || string(vals[0]) != "application/json" {
		t.Errorf("unexpected Content-Type values: %v", vals)
	}
}

func TestMapHttpHeaderTo_MultipleValues(t *testing.T) {
	mock := newMockFieldsSetter()
	h := http.Header{"Accept": {"text/html", "application/json"}}

	if err := mapHttpHeaderTo(h, mock); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vals, ok := mock.headers["Accept"]
	if !ok {
		t.Fatal("Accept header not set")
	}
	if len(vals) != 2 {
		t.Errorf("expected 2 Accept values, got %d", len(vals))
	}
}

func TestMapHttpHeaderTo_MultipleHeaders(t *testing.T) {
	mock := newMockFieldsSetter()
	h := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer token"},
		"X-Request-Id":  {"abc-123"},
	}

	if err := mapHttpHeaderTo(h, mock); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(mock.headers))
	}
}

func TestMapHttpHeaderTo_ValuesAreByteSlices(t *testing.T) {
	mock := newMockFieldsSetter()
	h := http.Header{"X-Custom": {"hello"}}

	if err := mapHttpHeaderTo(h, mock); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vals := mock.headers["X-Custom"]
	if len(vals) != 1 || string(vals[0]) != "hello" {
		t.Errorf("expected []byte(\"hello\"), got %v", vals)
	}
}

func TestMapHttpHeaderTo_SetReturnsError(t *testing.T) {
	mock := &mockFieldsSetter{
		headers:  make(map[string][][]uint8),
		errOnKey: "X-Invalid",
	}
	h := http.Header{"X-Invalid": {"value"}}

	err := mapHttpHeaderTo(h, mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
