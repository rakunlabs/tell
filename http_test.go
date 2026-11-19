package tell

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestNewHTTP(t *testing.T) {
	var (
		mu    sync.Mutex
		paths = map[string]int{}
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths[r.URL.Path]++
		mu.Unlock()

		if ct := r.Header.Get("Content-Type"); ct != "application/x-protobuf" {
			t.Errorf("content type = %q", ct)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv(envEndpoint, "")
	t.Setenv(envProtocol, ProtocolHTTPProtobuf)

	c, err := New(t.Context(), Config{Collector: srv.URL + "/otlp"})
	if err != nil {
		t.Fatal(err)
	}

	if c.IsMetricNoop() || c.IsTraceNoop() {
		t.Fatal("providers should not be noop")
	}

	if c.Conn != nil {
		t.Error("grpc connection should not be created")
	}

	_, span := otel.Tracer("").Start(context.Background(), "test")
	span.End()

	// metric interval is 5s, so metrics are only sent by the flush in shutdown
	if err := c.Shutdown(); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()

	for _, p := range []string{"/otlp/v1/traces", "/otlp/v1/metrics"} {
		if paths[p] == 0 {
			t.Errorf("no export to %s, got %v", p, paths)
		}
	}
}

func TestNewUnsupportedProtocol(t *testing.T) {
	_, err := New(t.Context(), Config{Collector: "localhost:4317", Protocol: "http/json"})
	if err == nil || !strings.Contains(err.Error(), "unsupported protocol") {
		t.Fatalf("err = %v", err)
	}
}
