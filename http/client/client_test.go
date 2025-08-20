package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jesse0michael/testhelpers/pkg/testserver"
)

func TestHttpClient(t *testing.T) {
	tests := []struct {
		name       string
		server     *testserver.Server
		wantErr    bool
		wantBody   string
		wantStatus int
	}{
		{
			name: "successful request",
			server: testserver.New(
				testserver.Handler{Path: "test", Status: http.StatusOK, Response: []byte(`{"status": "OK"}`)},
				testserver.Handler{Path: "test", Status: http.StatusTeapot, Response: []byte(`{"status": "Teapot"}`)},
			),
			wantErr:    false,
			wantBody:   `{"status": "OK"}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "closed server request",
			server: func() *testserver.Server {
				ts := testserver.New(
					testserver.Handler{Path: "test", Status: http.StatusTeapot, Response: []byte(`{"status": "Teapot"}`)},
				)
				ts.Close()
				return ts
			}(),
			wantErr:    true,
			wantBody:   ``,
			wantStatus: 0,
		},
		{
			name: "recovered failure request",
			server: testserver.New(
				testserver.Handler{Path: "test", Status: http.StatusInternalServerError, Response: []byte(`{"status": "InternalServerError"}`)},
				testserver.Handler{Path: "test", Status: http.StatusInternalServerError, Response: []byte(`{"status": "InternalServerError"}`)},
				testserver.Handler{Path: "test", Status: http.StatusInternalServerError, Response: []byte(`{"status": "InternalServerError"}`)},
				testserver.Handler{Path: "test", Status: http.StatusOK, Response: []byte(`{"status": "OK"}`)},
				testserver.Handler{Path: "test", Status: http.StatusTeapot, Response: []byte(`{"status": "Teapot"}`)},
			),
			wantErr:    false,
			wantBody:   `{"status": "OK"}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "exhausted failure request",
			server: testserver.New(
				testserver.Handler{Path: "test", Status: http.StatusInternalServerError, Response: []byte(`{"status": "InternalServerError"}`)},
				testserver.Handler{Path: "test", Status: http.StatusInternalServerError, Response: []byte(`{"status": "InternalServerError"}`)},
				testserver.Handler{Path: "test", Status: http.StatusInternalServerError, Response: []byte(`{"status": "InternalServerError"}`)},
				testserver.Handler{Path: "test", Status: http.StatusInternalServerError, Response: []byte(`{"status": "InternalServerError"}`)},
				testserver.Handler{Path: "test", Status: http.StatusTeapot, Response: []byte(`{"status": "Teapot"}`)},
			),
			wantErr:    false,
			wantBody:   `{"status": "InternalServerError"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "non-retryable request",
			server: testserver.New(
				testserver.Handler{Path: "test", Status: http.StatusBadRequest, Response: []byte(`{"status": "BadRequest"}`)},
				testserver.Handler{Path: "test", Status: http.StatusTeapot, Response: []byte(`{"status": "Teapot"}`)},
			),
			wantErr:    false,
			wantBody:   `{"status": "BadRequest"}`,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := HTTPClient()
			req, _ := http.NewRequest("GET", tt.server.URL+"/test", nil)
			resp, err := c.Do(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("HttpClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var b []byte
			if resp != nil {
				b, _ = io.ReadAll(resp.Body)
				_ = resp.Body.Close()
			}
			if !strings.Contains(string(b), tt.wantBody) {
				t.Errorf("HttpClient() resp.Body = %v, wantBody %v", string(b), tt.wantBody)
				return
			}

			var statusCode int
			if resp != nil {
				statusCode = resp.StatusCode
			}
			if statusCode != tt.wantStatus {
				t.Errorf("HttpClient() resp.Status = %v, wantStatus %v", err, tt.wantStatus)
				return
			}
			tt.server.Close()
		})
	}
}
