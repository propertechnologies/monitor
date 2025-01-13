package contador

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const traceparent = "traceparent"

type Client struct {
	client             *http.Client
	authorizationToken *string
}

func NewClient(client *http.Client) *Client {
	return &Client{client: client}
}

func (c *Client) SetAuthorizationToken(token string) {
	c.authorizationToken = &token
}

func (c *Client) DoRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	request, err := c.buildRequest(method, url, body, func(r *http.Request) {
		c.SetAuthorizationheader(r)
		SetFlowID(r)
		SetTraceparentHeader(r)
	})
	if err != nil {
		return nil, err
	}

	return c.execute(ctx, request)
}

func (c *Client) SetAuthorizationheader(request *http.Request) {
	if c.authorizationToken != nil {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *c.authorizationToken))
	} else {
		c.SetAuthorizationheader(request)
	}
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

func (c *Client) execute(ctx context.Context, req *http.Request) ([]byte, error) {
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
	request, err := c.buildRequest(method, url, body, func(r *http.Request) {
		c.SetAuthorizationheader(r)
		SetFlowID(r)
		SetTraceparentHeader(r)
	})

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	if err != nil {
		return nil, err
	}

	return c.execute(ctx, request)
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
