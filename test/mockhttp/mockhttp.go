package mockhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type RequestMatcher func(*http.Request) bool
type ResponseBuilder func(*http.Request) *http.Response

type expectation struct {
	matchers []RequestMatcher
	response ResponseBuilder
	called   int
	times    int // expected call count; 0 means at least once
	desc     string
}

type MockHTTP struct {
	server       *httptest.Server
	tlsServer    *httptest.Server
	mu           sync.Mutex
	expectations []*expectation
	requests     []*http.Request
	URL          string
	Client       *http.Client
}

// NewMockHTTP creates a new mock HTTP server that handles both HTTP and HTTPS requests.
func NewMockHTTP() *MockHTTP {
	m := &MockHTTP{}
	handler := http.HandlerFunc(m.handle)
	m.server = httptest.NewServer(handler)
	m.tlsServer = httptest.NewTLSServer(handler)
	m.URL = m.server.URL

	httpAddr := m.server.Listener.Addr().String()
	tlsAddr := m.tlsServer.Listener.Addr().String()

	transport := m.tlsServer.Client().Transport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, httpAddr)
	}
	transport.DialTLSContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return tls.DialWithDialer(&net.Dialer{}, network, tlsAddr, &tls.Config{InsecureSkipVerify: true})
	}
	m.Client = &http.Client{Transport: transport}
	return m
}

// Expect sets up a new expectation
func (m *MockHTTP) Expect(matchers ...RequestMatcher) *expectation {
	e := &expectation{matchers: matchers}
	m.mu.Lock()
	m.expectations = append(m.expectations, e)
	m.mu.Unlock()
	return e
}

// Returns sets the response for an expectation.
func (e *expectation) Returns(opts ...ResponseOption) *expectation {
	e.response = buildResponse(opts...)
	return e
}

// Times sets the exact number of times this expectation should be matched.
func (e *expectation) Times(n int) *expectation {
	e.times = n
	return e
}

// handle matches requests and returns responses.
func (m *MockHTTP) handle(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requests = append(m.requests, r)
	for i, e := range m.expectations {
		if e.times > 0 && e.called >= e.times {
			continue
		}
		if matchAll(e.matchers, r) {
			e.called++
			resp := e.response(r)
			writeResponse(w, resp)
			// Move matched expectation to end so earlier-registered expectations get priority.
			m.expectations = append(append(m.expectations[:i], m.expectations[i+1:]...), e)
			m.mu.Unlock()
			return
		}
	}
	// Default: 501 Not Implemented
	w.WriteHeader(http.StatusNotImplemented)
	m.mu.Unlock()
}

// Validate checks that all expectations were called the expected number of times.
func (m *MockHTTP) Validate(t *testing.T) {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.expectations {
		name := e.desc
		if name == "" {
			name = "expectation"
		}
		switch {
		case e.times > 0 && e.called != e.times:
			t.Errorf("mockhttp: %s: expected %d call(s), got %d", name, e.times, e.called)
		case e.times == 0 && e.called == 0:
			t.Errorf("mockhttp: %s: expected at least 1 call, got 0", name)
		}
	}
}

// Close shuts down the test servers
func (m *MockHTTP) Close() {
	m.server.Close()
	m.tlsServer.Close()
}

// --- Helper types and functions ---

type ResponseOption func(*http.Response)

func buildResponse(opts ...ResponseOption) ResponseBuilder {
	return func(r *http.Request) *http.Response {
		resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
		for _, opt := range opts {
			opt(resp)
		}
		return resp
	}
}

func writeResponse(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		_, _ = io.Copy(w, resp.Body)
	}
}

func matchAll(matchers []RequestMatcher, r *http.Request) bool {
	for _, m := range matchers {
		if !m(r) {
			return false
		}
	}
	return true
}

// --- Matcher and ResponseOption helpers ---

func Method(method string) RequestMatcher {
	return func(r *http.Request) bool { return r.Method == method }
}
func Path(path string) RequestMatcher {
	return func(r *http.Request) bool { return r.URL.Path == path }
}
func Header(key, value string) RequestMatcher {
	return func(r *http.Request) bool { return r.Header.Get(key) == value }
}
func Body(expected string) RequestMatcher {
	return func(r *http.Request) bool {
		b, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(b))
		return string(b) == expected
	}
}
func Query(key, value string) RequestMatcher {
	return func(r *http.Request) bool { return r.URL.Query().Get(key) == value }
}

func Status(code int) ResponseOption {
	return func(resp *http.Response) { resp.StatusCode = code }
}

func BodyResp(body string) ResponseOption {
	return func(resp *http.Response) { resp.Body = io.NopCloser(strings.NewReader(body)) }
}

func HeaderResp(key, value string) ResponseOption {
	return func(resp *http.Response) { resp.Header.Add(key, value) }
}
