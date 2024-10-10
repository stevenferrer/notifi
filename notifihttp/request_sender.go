package notifihttp

import (
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type RequestSender interface {
	SendRequest(ctx context.Context, urlStr string,
		body io.Reader, headers map[string]string) error
}

// DefaultRequestSender is the default HTTP request sender
type DefaultRequestSender struct {
	httpClient *http.Client
}

var _ RequestSender = (*DefaultRequestSender)(nil)

// NewDefaultRequestSender returns a new DefaultRequestSender
func NewDefaultRequestSender() *DefaultRequestSender {
	return &DefaultRequestSender{
		httpClient: http.DefaultClient,
	}
}

// WithHTTPClient overrides the default HTTP client
func (rs *DefaultRequestSender) WithHTTPClient(httpClient *http.Client) *DefaultRequestSender {
	rs.httpClient = httpClient
	return rs
}

// SendRequest builds and sends the HTTP request
func (rs *DefaultRequestSender) SendRequest(ctx context.Context,
	urlStr string, body io.Reader, headers map[string]string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, body)
	if err != nil {
		return errors.Wrap(err, "new http request")
	}
	httpReq.Header.Add("content-type", "application/json")

	// Additional headers
	for key, value := range headers {
		httpReq.Header.Add(key, value)
	}

	var httpResp *http.Response
	httpResp, err = rs.httpClient.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "send http request")
	}

	if httpResp.StatusCode != http.StatusOK {
		return errors.Errorf("expecting status 200 but got %v", httpResp.StatusCode)
	}

	return nil
}
