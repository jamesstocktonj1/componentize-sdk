package httptypes

import (
	"net/http"
	"testing"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// mockFieldsSetter implements FieldsSetter for testing.
type mockFieldsSetter struct {
	headers  map[string][][]uint8
	errOnKey string // if non-empty, Set returns an error for this key
}

func (m *mockFieldsSetter) Set(name string, value [][]uint8) witTypes.Result[witTypes.Unit, types.HeaderError] {
	if m.errOnKey != "" && name == m.errOnKey {
		return witTypes.Err[witTypes.Unit, types.HeaderError](types.MakeHeaderErrorInvalidSyntax())
	}
	m.headers[name] = value
	return witTypes.Ok[witTypes.Unit, types.HeaderError](witTypes.Unit{})
}

func TestMapHttpHeaderTo(t *testing.T) {
	testMatrix := []struct {
		name     string
		input    http.Header
		errOnKey string
		wantErr  bool
		checkFn  func(t *testing.T, m *mockFieldsSetter)
	}{
		{
			name:  "empty header",
			input: http.Header{},
			checkFn: func(t *testing.T, m *mockFieldsSetter) {
				if len(m.headers) != 0 {
					t.Errorf("expected no headers set, got %v", m.headers)
				}
			},
		},
		{
			name:  "single header single value",
			input: http.Header{"Content-Type": {"application/json"}},
			checkFn: func(t *testing.T, m *mockFieldsSetter) {
				vals, ok := m.headers["Content-Type"]
				if !ok {
					t.Fatal("Content-Type not set")
				}
				if len(vals) != 1 || string(vals[0]) != "application/json" {
					t.Errorf("unexpected Content-Type values: %v", vals)
				}
			},
		},
		{
			name:  "single header multiple values",
			input: http.Header{"Accept": {"text/html", "application/json"}},
			checkFn: func(t *testing.T, m *mockFieldsSetter) {
				vals, ok := m.headers["Accept"]
				if !ok {
					t.Fatal("Accept not set")
				}
				if len(vals) != 2 {
					t.Errorf("expected 2 values, got %d", len(vals))
				}
			},
		},
		{
			name: "multiple headers",
			input: http.Header{
				"Content-Type":  {"application/json"},
				"Authorization": {"Bearer token"},
				"X-Request-Id":  {"abc-123"},
			},
			checkFn: func(t *testing.T, m *mockFieldsSetter) {
				if len(m.headers) != 3 {
					t.Errorf("expected 3 headers, got %d", len(m.headers))
				}
			},
		},
		{
			name:     "set error is propagated",
			input:    http.Header{"X-Invalid": {"value"}},
			errOnKey: "X-Invalid",
			wantErr:  true,
		},
	}

	for _, tt := range testMatrix {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFieldsSetter{
				headers:  make(map[string][][]uint8),
				errOnKey: tt.errOnKey,
			}
			err := mapHttpHeaderTo(tt.input, mock)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, mock)
			}
		})
	}
}
