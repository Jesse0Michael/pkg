package client

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	httperrors "github.com/jesse0michael/pkg/http/errors"
)

func TestNew(t *testing.T) {
	customClient := &http.Client{}
	customLogger := slog.New(slog.DiscardHandler)
	mustURL := func(s string) *url.URL { u, _ := url.Parse(s); return u }
	defaultLogger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name string
		opts []Option
		want *REST
	}{
		{
			name: "defaults",
			opts: nil,
			want: &REST{
				headers:  http.Header{},
				logger:   defaultLogger,
				maxBytes: DefaultMaxResponseBytes,
			},
		},
		{
			name: "WithHTTPClient",
			opts: []Option{WithHTTPClient(customClient)},
			want: &REST{
				client:   customClient,
				headers:  http.Header{},
				logger:   defaultLogger,
				maxBytes: DefaultMaxResponseBytes,
			},
		},
		{
			name: "WithBaseURL",
			opts: []Option{WithBaseURL("https://api.example.com/v1")},
			want: &REST{
				baseURL:  mustURL("https://api.example.com/v1"),
				headers:  http.Header{},
				logger:   defaultLogger,
				maxBytes: DefaultMaxResponseBytes,
			},
		},
		{
			name: "WithHeader",
			opts: []Option{WithHeader("X-Test", "value")},
			want: &REST{
				headers:  http.Header{"X-Test": []string{"value"}},
				logger:   defaultLogger,
				maxBytes: DefaultMaxResponseBytes,
			},
		},
		{
			name: "WithHeaders",
			opts: []Option{WithHeaders(http.Header{"X-Test": []string{"v1", "v2"}})},
			want: &REST{
				headers:  http.Header{"X-Test": []string{"v1", "v2"}},
				logger:   defaultLogger,
				maxBytes: DefaultMaxResponseBytes,
			},
		},
		{
			name: "WithLogger",
			opts: []Option{WithLogger(customLogger)},
			want: &REST{
				headers:  http.Header{},
				logger:   customLogger,
				maxBytes: DefaultMaxResponseBytes,
			},
		},
		{
			name: "WithMaxResponseBytes",
			opts: []Option{WithMaxResponseBytes(1024)},
			want: &REST{
				headers:  http.Header{},
				logger:   defaultLogger,
				maxBytes: 1024,
			},
		},
		{
			name: "WithMaxResponseBytes zero disables cap",
			opts: []Option{WithMaxResponseBytes(0)},
			want: &REST{
				headers:  http.Header{},
				logger:   defaultLogger,
				maxBytes: 0,
			},
		},
		{
			name: "all options together",
			opts: []Option{
				WithHTTPClient(customClient),
				WithBaseURL("https://api.example.com"),
				WithHeader("X-A", "a"),
				WithHeaders(http.Header{"X-B": []string{"b"}}),
				WithLogger(customLogger),
				WithMaxResponseBytes(2048),
			},
			want: &REST{
				client:   customClient,
				baseURL:  mustURL("https://api.example.com"),
				headers:  http.Header{"X-A": []string{"a"}, "X-B": []string{"b"}},
				logger:   customLogger,
				maxBytes: 2048,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.opts...)
			// HTTPClient() is non-deterministic; if a case doesn't override it,
			// adopt whatever default got produced so the rest can be DeepEqual'd.
			if tt.want.client == nil {
				if got.client == nil {
					t.Fatal("client should not be nil")
				}
				tt.want.client = got.client
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

type testThing struct {
	A string `json:"a" xml:"a" yaml:"a"`
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		out     any // pointer to destination, or nil
		want    any // expected value of *out, compared via reflect.DeepEqual
		wantErr bool
	}{
		{
			name: "json struct decode",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"a":"hello"}`))
			},
			out:  &testThing{},
			want: testThing{A: "hello"},
		},
		{
			name: "xml struct decode",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				_, _ = w.Write([]byte(`<testThing><a>hello</a></testThing>`))
			},
			out:  &testThing{},
			want: testThing{A: "hello"},
		},
		{
			name: "yaml struct decode",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/yaml")
				_, _ = w.Write([]byte("a: hello\n"))
			},
			out:  &testThing{},
			want: testThing{A: "hello"},
		},
		{
			name: "missing content-type falls back to json",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"a":"hello"}`))
			},
			out:  &testThing{},
			want: testThing{A: "hello"},
		},
		{
			name: "raw bytes",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte("hello"))
			},
			out:  &[]byte{},
			want: []byte("hello"),
		},
		{
			name: "raw string",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte("hello"))
			},
			out:  new(string),
			want: "hello",
		},
		{
			name: "nil out discards body",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"a":"hello"}`))
			},
			out:  nil,
			want: nil,
		},
		{
			name: "empty body with struct out is no-op",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			out:  &testThing{A: "untouched"},
			want: testThing{A: "untouched"},
		},
		{
			name: "non-2xx returns error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`oops`))
			},
			out:     &testThing{},
			want:    testThing{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)

			req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
			err := New().Process(t.Context(), req, tt.out)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Process() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Process() error = %v, want nil", err)
			}
			if tt.out == nil {
				return
			}
			got := reflect.ValueOf(tt.out).Elem().Interface()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() out = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessNonOKError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`oops`))
	}))
	t.Cleanup(srv.Close)

	c := New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	err := c.Process(t.Context(), req, nil)

	var herr *httperrors.Error
	if !errors.As(err, &herr) {
		t.Fatalf("Process() error = %v, want *httperrors.Error", err)
	}
	if herr.Code != http.StatusBadRequest {
		t.Errorf("Process() error code = %d, want %d", herr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(herr.Details, "oops") {
		t.Errorf("Process() error details = %q, want to contain %q", herr.Details, "oops")
	}
}

func TestProcessIOWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello"))
	}))
	t.Cleanup(srv.Close)

	var buf bytes.Buffer
	c := New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	if err := c.Process(t.Context(), req, &buf); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if got := buf.String(); got != "hello" {
		t.Errorf("buf = %q, want %q", got, "hello")
	}
}

func TestProcessHeaders(t *testing.T) {
	tests := []struct {
		name      string
		opts      []Option
		reqHeader http.Header
		want      string
	}{
		{
			name: "default header applied",
			opts: []Option{WithHeader("X-Test", "from-default")},
			want: "from-default",
		},
		{
			name:      "per-request header overrides default",
			opts:      []Option{WithHeader("X-Test", "from-default")},
			reqHeader: http.Header{"X-Test": []string{"from-request"}},
			want:      "from-request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"a":"` + r.Header.Get("X-Test") + `"}`))
			}))
			t.Cleanup(srv.Close)

			c := New(tt.opts...)
			req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
			for k, vs := range tt.reqHeader {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
			var got testThing
			if err := c.Process(t.Context(), req, &got); err != nil {
				t.Fatalf("Process() error = %v", err)
			}
			if got.A != tt.want {
				t.Errorf("got X-Test = %q, want %q", got.A, tt.want)
			}
		})
	}
}

func TestProcessBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"a":"` + r.URL.Path + `"}`))
	}))
	t.Cleanup(srv.Close)

	c := New(WithBaseURL(srv.URL + "/echo-path/"))
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "relative", nil)
	var got testThing
	if err := c.Process(t.Context(), req, &got); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if want := "/echo-path/relative"; got.A != want {
		t.Errorf("path = %q, want %q", got.A, want)
	}
}

func TestProcessMaxBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(bytes.Repeat([]byte("x"), 1024))
	}))
	t.Cleanup(srv.Close)

	c := New(WithMaxResponseBytes(512))
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	var b []byte
	err := c.Process(t.Context(), req, &b)
	if err == nil {
		t.Fatal("Process() error = nil, want overflow error")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("Process() error = %v, want overflow error", err)
	}
}

func TestProcessTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close() // closed before request

	c := New(WithHTTPClient(&http.Client{}))
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	err := c.Process(req.Context(), req, nil)
	if err == nil {
		t.Fatal("Process() error = nil, want transport error")
	}
	if !strings.Contains(err.Error(), "failed to do request") {
		t.Errorf("Process() error = %v, want wrapped transport error", err)
	}
}

func TestProcessContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	c := New(WithHTTPClient(&http.Client{}))
	ctx, cancel := context.WithCancel(t.Context())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	cancel()
	err := c.Process(ctx, req, nil)
	if err == nil {
		t.Fatal("Process() error = nil, want cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Process() error = %v, want context.Canceled", err)
	}
}

