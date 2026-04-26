package client

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/jesse0michael/pkg/http/errors"
	jsoniter "github.com/json-iterator/go"
	"go.yaml.in/yaml/v3"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// DefaultMaxResponseBytes caps response body reads. Set via WithMaxResponseBytes;
// pass 0 to disable the cap entirely.
const DefaultMaxResponseBytes int64 = 32 << 20 // 32 MiB

// REST is an embeddable HTTP client for REST APIs. Concrete clients embed *REST
// and expose typed methods that build an *http.Request and call Process.
type REST struct {
	client   *http.Client
	baseURL  *url.URL
	headers  http.Header
	logger   *slog.Logger
	maxBytes int64
}

// Option configures a REST client.
type Option func(*REST)

// New creates a REST client with the given options. By default it uses HTTPClient()
// (retrying, OTel-instrumented), an empty header set, a discard logger, and a
// DefaultMaxResponseBytes response cap.
func New(opts ...Option) *REST {
	r := &REST{
		client:   HTTPClient(),
		headers:  http.Header{},
		logger:   slog.New(slog.DiscardHandler),
		maxBytes: DefaultMaxResponseBytes,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// WithHTTPClient overrides the underlying *http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(r *REST) { r.client = c }
}

// WithBaseURL sets a base URL that relative request URLs are resolved against.
// Absolute request URLs are passed through unchanged.
func WithBaseURL(s string) Option {
	return func(r *REST) {
		if u, err := url.Parse(s); err == nil {
			r.baseURL = u
		}
	}
}

// WithHeader sets a default header applied to every request. Per-request headers
// already set on the *http.Request take precedence.
func WithHeader(k, v string) Option {
	return func(r *REST) { r.headers.Set(k, v) }
}

// WithHeaders merges the given headers into the default header set.
func WithHeaders(h http.Header) Option {
	return func(r *REST) {
		for k, vs := range h {
			for _, v := range vs {
				r.headers.Add(k, v)
			}
		}
	}
}

// WithLogger overrides the slog.Logger used for request logging.
func WithLogger(l *slog.Logger) Option {
	return func(r *REST) { r.logger = l }
}

// WithMaxResponseBytes caps the number of response body bytes read into memory.
// Pass 0 to disable the cap.
func WithMaxResponseBytes(n int64) Option {
	return func(r *REST) { r.maxBytes = n }
}

// Process sends req and writes the response into out, choosing how based on the
// type of out:
//
//   - nil: response body is discarded.
//   - *[]byte: raw bytes.
//   - *string: raw string.
//   - io.Writer: streamed via io.Copy.
//   - anything else: JSON-decoded.
//
// Non-2xx responses return an *errors.Error carrying the status code and body.
// Response body reads are capped by WithMaxResponseBytes (0 disables the cap).
// The response body is always closed.
func (c *REST) Process(ctx context.Context, req *http.Request, out any) error {
	req = req.WithContext(ctx)
	if c.baseURL != nil && req.URL != nil {
		req.URL = c.baseURL.ResolveReference(req.URL)
	}
	for k, vs := range c.headers {
		if _, ok := req.Header[k]; ok {
			continue
		}
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.LogAttrs(ctx, slog.LevelWarn, "http request failed",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.Any("err", err),
		)
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := c.readBody(resp.Body)
		return errors.NewError(resp.StatusCode, http.StatusText(resp.StatusCode), string(b))
	}

	switch v := out.(type) {
	case nil:
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	case io.Writer:
		_, err := io.Copy(v, c.limit(resp.Body))
		if err != nil {
			return fmt.Errorf("failed to copy response: %w", err)
		}
		return nil
	case *[]byte:
		b, err := c.readBody(resp.Body)
		if err != nil {
			return err
		}
		*v = b
		return nil
	case *string:
		b, err := c.readBody(resp.Body)
		if err != nil {
			return err
		}
		*v = string(b)
		return nil
	default:
		b, err := c.readBody(resp.Body)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			return nil
		}
		if err := unmarshal(resp.Header.Get("Content-Type"), b, out); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return nil
	}
}

// unmarshal decodes b into out, choosing a format from the Content-Type header.
// XML and YAML media types use their respective decoders; everything else
// (including missing or unrecognized Content-Type) falls back to JSON.
func unmarshal(contentType string, b []byte, out any) error {
	media, _, _ := mime.ParseMediaType(contentType)
	switch {
	case media == "application/xml" || media == "text/xml" || strings.HasSuffix(media, "+xml"):
		return xml.Unmarshal(b, out)
	case media == "application/yaml" || media == "application/x-yaml" ||
		media == "text/yaml" || media == "text/x-yaml" || strings.HasSuffix(media, "+yaml"):
		return yaml.Unmarshal(b, out)
	default:
		return json.Unmarshal(b, out)
	}
}

// limit wraps r in an io.LimitReader honoring c.maxBytes (0 = unlimited).
// Reads one extra byte so the caller can detect overflow.
func (c *REST) limit(r io.Reader) io.Reader {
	if c.maxBytes <= 0 {
		return r
	}
	return io.LimitReader(r, c.maxBytes+1)
}

// readBody reads up to maxBytes from r, erroring if the body exceeds the cap.
func (c *REST) readBody(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(c.limit(r))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if c.maxBytes > 0 && int64(len(b)) > c.maxBytes {
		return nil, fmt.Errorf("response body exceeds %d bytes", c.maxBytes)
	}
	return b, nil
}
