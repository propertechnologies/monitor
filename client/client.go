package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

const traceparent = "traceparent"

type (
	Client struct {
		client             HTTPClient
		authorizationToken string
	}

	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}
)

func NewClient(client HTTPClient) *Client {
	return &Client{client: client}
}

func NewClientWithTokent(client HTTPClient, token string) *Client {
	return &Client{client: client, authorizationToken: token}
}

func (c *Client) DoRequest(
	ctx context.Context,
	method, url string,
	body io.Reader,
) ([]byte, error) {
	request, err := c.setGenericHeaders(method, url, body, nil)
	if err != nil {
		return nil, err
	}

	return c.execute(ctx, request)
}

func (c *Client) DoRequestWithExtraHeaders(
	ctx context.Context,
	method, url string,
	body io.Reader,
	extraHeaders map[string]string) ([]byte, error) {
	request, err := c.setGenericHeaders(method, url, body, extraHeaders)
	if err != nil {
		return nil, err
	}

	return c.execute(ctx, request)
}

func SetAuthorizationHeader(request *http.Request, token string) {
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (c *Client) buildRequest(
	method, url string,
	body io.Reader,
	requestModifier func(*http.Request),
) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	if requestModifier != nil {
		requestModifier(req)
	}

	return req, nil
}

func (c *Client) execute(_ context.Context, req *http.Request) ([]byte, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		var err error

		if bodyBytes != nil {
			err = fmt.Errorf("status %d, message %s", res.StatusCode, string(bodyBytes))
		} else {
			err = fmt.Errorf("status %d", res.StatusCode)
		}

		return nil, err
	}

	return bodyBytes, nil
}

func (c *Client) BuildUrl(baseURL string, params map[string]string) string {
	u, _ := url.Parse(baseURL)

	queryParams := url.Values{}
	for param, value := range params {
		queryParams.Add(param, value)
	}

	u.RawQuery = queryParams.Encode()
	return u.String()
}

func (c *Client) DoRequestWithContentType(ctx context.Context, method, url string, body io.Reader, contentType string) ([]byte, error) {
	request, err := c.setGenericHeaders(method, url, body, nil)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	return c.execute(ctx, request)
}

func (c *Client) setGenericHeaders(method string, url string, body io.Reader, extraHeaders map[string]string) (*http.Request, error) {
	request, err := c.buildRequest(method, url, body, func(r *http.Request) {
		if c.authorizationToken != "" {
			SetAuthorizationHeader(r, c.authorizationToken)
		}
		SetFlowID(r)
		SetTraceparentHeader(r)

		if len(extraHeaders) != 0 {
			for headerName, headerValue := range extraHeaders {
				SetHeader(r, headerName, headerValue)
			}
		}
	})
	if err != nil {
		return nil, err
	}

	return request, nil
}

func SetHeader(request *http.Request, headerName, headerValue string) {
	request.Header.Set(headerName, headerValue)
}

func SetTraceparentHeader(request *http.Request) {
	request.Header.Set("proper-referer", GetTraceparent())
	request.Header.Set(traceparent, GetTraceparent())
}

func SetFlowID(request *http.Request) {
	request.Header.Set("X-Flow-Id", GetFlowID())
}

func GetTraceparent() string {
	return os.Getenv(traceparent)
}

func GetFlowID() string {
	return os.Getenv("FLOW")
}

// MultipartFile struct to encapsulate file data
type MultipartFile struct {
	FieldName string
	FileName  string
	Reader    io.Reader
}

func (c *Client) BuildMultipartFormRequest(
	ctx context.Context,
	method, url string,
	formData map[string]string,
	file *MultipartFile,
) (*http.Request, error) {
	// Create a buffer to hold the multipart form data
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Add form fields
	for key, value := range formData {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}

	// Add file
	fileWriter, err := writer.CreateFormFile(file.FieldName, file.FileName)
	if err != nil {
		return nil, err
	}

	// Copy the file content to the file writer
	if _, err = io.Copy(fileWriter, file.Reader); err != nil {
		return nil, err
	}

	// Close the writer to finalize the form data
	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Create a new request with the multipart form data
	req, err := http.NewRequest(method, url, &buffer)
	if err != nil {
		return nil, err
	}

	// Set the content type to multipart/form-data with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set additional headers if necessary
	if c.authorizationToken != "" {
		SetAuthorizationHeader(req, c.authorizationToken)
	}
	SetFlowID(req)
	SetTraceparentHeader(req)

	return req, nil
}

func (c *Client) DoRequestRaw(ctx context.Context, req *http.Request) (*http.Response, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return res, err
	}

	return res, nil
}
