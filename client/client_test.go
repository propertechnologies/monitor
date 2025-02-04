package client

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
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

type MockFile struct {
	content string
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	copy(p, m.content)
	return len(m.content), io.EOF
}

func TestBuildMultipartFormRequest(t *testing.T) {
	client := &Client{authorizationToken: "myToken"}
	formData := map[string]string{"field1": "value1", "field2": "value2"}
	mockFile := &MockFile{content: "file content"}
	file := &MultipartFile{FieldName: "file", FileName: "test.txt", Reader: mockFile}

	req, err := client.BuildMultipartFormRequest(context.Background(), "POST", "http://example.com", formData, file)

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "POST", req.Method)
	assert.Contains(t, req.Header.Get("Content-Type"), "multipart/form-data")

	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(req.Body)
	assert.NoError(t, err)

	boundary := req.Header.Get("Content-Type")[30:]

	reader := multipart.NewReader(buffer, boundary)

	var fileFound bool

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)

		if part.FormName() == "file" {
			assert.Equal(t, "test.txt", part.FileName())
			fileFound = true
			break
		}
	}

	assert.True(t, fileFound, "File part not found in multipart request")
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
