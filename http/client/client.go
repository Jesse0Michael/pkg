package client

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jesse0michael/pkg/http/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func HTTPClient() *http.Client {
	transport := cleanhttp.DefaultPooledTransport()
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = slog.Default()
	retryClient.RetryMax = 3
	retryClient.HTTPClient.Transport = otelhttp.NewTransport(transport)
	retryClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		if resp != nil && err != nil {
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			return resp, errors.NewError(resp.StatusCode, err.Error(), string(b))
		}
		return resp, err
	}
	return retryClient.StandardClient()
}
