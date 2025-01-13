package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThatAuthorizationTokenIsSent(t *testing.T) {
	httpClientMock := newHTTPClientMock()
	httpClientMock.checkToken = true

	cl := NewClientWithTokent(httpClientMock, "myToken")

	resp, err := cl.DoRequest(context.Background(), "GET", "http://example.com", nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

func TestThatAuthorizationTokenIsNotSentWhenUsingSimpleConstructor(t *testing.T) {
	httpClientMock := newHTTPClientMock()

	cl := NewClient(httpClientMock)

	resp, err := cl.DoRequest(context.Background(), "GET", "http://example.com", nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

func TestThatErrorCodeDifferentFrom2xxReturnsError(t *testing.T) {
	httpClientMock := newHTTPClientMock()
	httpClientMock.status = 500

	cl := NewClient(httpClientMock)

	resp, err := cl.DoRequest(context.Background(), "GET", "http://example.com", nil)

	assert.Error(t, err)
	assert.Empty(t, resp)
}

func TestThatContentTypeIsSent(t *testing.T) {
	httpClientMock := newHTTPClientMock()
	httpClientMock.checkContentType = true

	cl := NewClient(httpClientMock)

	resp, err := cl.DoRequestWithContentType(context.Background(), "GET", "http://example.com", nil, "application/json")

	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

type HTTPClientMock struct {
	checkToken       bool
	checkContentType bool
	status           int
}

func newHTTPClientMock() *HTTPClientMock {
	return &HTTPClientMock{
		status:           200,
		checkToken:       false,
		checkContentType: false,
	}
}
func (h *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	if h.checkToken && req.Header.Get("Authorization") != "Bearer myToken" {
		return nil, http.ErrHandlerTimeout
	}

	if h.checkContentType && req.Header.Get("Content-Type") == "" {
		return nil, http.ErrHandlerTimeout
	}

	return &http.Response{
		Body:       io.NopCloser(strings.NewReader("Hello, client")),
		StatusCode: h.status,
	}, nil
}
