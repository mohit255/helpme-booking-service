package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-helpme-booking/src/config"
	"go-helpme-booking/src/utils/logger"
	"go.uber.org/zap"
)

// redactedHeaderKeys are never logged in plain text.
var redactedHeaderKeys = map[string]bool{
	"authorization": true,
	"x-api-key":     true,
}

type ctxKey string

const extraHeadersCtxKey ctxKey = "http_client_extra_headers"

// WithHeaders attaches per-call headers (e.g. a forwarded Authorization token) to ctx.
// They're merged on top of the client's default headers for any HTTPClient call made
// with the returned context — the client instance itself (and its defaults) stay
// unmodified, so this is safe to use on a shared client across concurrent requests.
func WithHeaders(ctx context.Context, headers map[string]string) context.Context {
	return context.WithValue(ctx, extraHeadersCtxKey, headers)
}

// HTTPResponse is the result of a call made through HTTPClient.
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// HTTPClient is the shared HTTP client helper used by every
// service-specific client (UserServiceClient, and any sibling-service client
// added later). It sets default headers, retries failed calls, and logs the
// full request/response on any non-success status.
type HTTPClient struct {
	name       string
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// NewHTTPClient builds a base client for `name` (used in logs/errors) at baseURL,
// with default headers (Content-Type, Accept, X-Service-Name) already set. Callers
// append/override headers via WithHeader (e.g. Authorization, X-Request-ID) — those
// are merged on top of the defaults set here.
func NewHTTPClient(name, baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		name:    name,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		headers: map[string]string{
			config.HeaderContentType: "application/json",
			"Accept":                 "application/json",
			config.HeaderServiceName: config.Common.AppName,
		},
	}
}

// WithHeader appends/overrides a default header sent on every subsequent call.
func (c *HTTPClient) WithHeader(key, value string) *HTTPClient {
	c.headers[key] = value
	return c
}

// Get issues a GET call. retry is optional; it defaults to config.DefaultHTTPRetry (1 = no retry).
func (c *HTTPClient) Get(ctx context.Context, path string, dest interface{}, retry ...int) (*HTTPResponse, error) {
	return c.do(ctx, http.MethodGet, path, nil, dest, resolveRetry(retry))
}

// Post issues a POST call with a JSON body. retry is optional; it defaults to config.DefaultHTTPRetry (1 = no retry).
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, dest interface{}, retry ...int) (*HTTPResponse, error) {
	return c.do(ctx, http.MethodPost, path, body, dest, resolveRetry(retry))
}

func resolveRetry(retry []int) int {
	if len(retry) > 0 && retry[0] > 0 {
		return retry[0]
	}
	return config.DefaultHTTPRetry
}

// isSuccessStatus treats 200 as success for every method, plus 201 (Created) for POST.
func isSuccessStatus(method string, status int) bool {
	if status == http.StatusOK {
		return true
	}
	return method == http.MethodPost && status == http.StatusCreated
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body interface{}, dest interface{}, retry int) (*HTTPResponse, error) {
	var rawReqBody []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("%s: marshal request body: %w", c.name, err)
		}
		rawReqBody = b
	}

	url := c.baseURL + path

	var lastErr error
	var lastResp *HTTPResponse

	for attempt := 1; attempt <= retry; attempt++ {
		if attempt > 1 && ctx.Err() != nil {
			lastErr = ctx.Err()
			break
		}

		var reqBody io.Reader
		if rawReqBody != nil {
			reqBody = bytes.NewBuffer(rawReqBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("%s: build request: %w", c.name, err)
		}
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
		if extra, ok := ctx.Value(extraHeadersCtxKey).(map[string]string); ok {
			for k, v := range extra {
				req.Header.Set(k, v)
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%s: execute request: %w", c.name, err)
			logger.Warn(c.name+" request failed",
				zap.String("method", method),
				zap.String("url", url),
				zap.Int("attempt", attempt),
				zap.Int("max_retry", retry),
				zap.Error(err),
			)
			continue
		}

		rawRespBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("%s: read response: %w", c.name, readErr)
			logger.Warn(c.name+" failed to read response body",
				zap.String("method", method),
				zap.String("url", url),
				zap.Int("attempt", attempt),
				zap.Error(readErr),
			)
			continue
		}

		result := &HTTPResponse{StatusCode: resp.StatusCode, Body: rawRespBody, Headers: resp.Header}
		lastResp = result

		if !isSuccessStatus(method, resp.StatusCode) {
			logger.Error(c.name+" non-success response",
				zap.String("method", method),
				zap.String("url", url),
				zap.Int("attempt", attempt),
				zap.Int("max_retry", retry),
				zap.Int("status", resp.StatusCode),
				zap.Any("request_headers", redactHeaders(c.headers)),
				zap.ByteString("request_body", rawReqBody),
				zap.Any("response_headers", redactHTTPHeaders(resp.Header)),
				zap.ByteString("response_body", rawRespBody),
			)
			lastErr = fmt.Errorf("%s: upstream status %d: %s", c.name, resp.StatusCode, string(rawRespBody))
			continue
		}

		if dest != nil && len(rawRespBody) > 0 {
			if err := json.Unmarshal(rawRespBody, dest); err != nil {
				return result, fmt.Errorf("%s: decode response: %w", c.name, err)
			}
		}
		return result, nil
	}

	return lastResp, lastErr
}

func redactHeaders(headers map[string]string) map[string]string {
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		if redactedHeaderKeys[strings.ToLower(k)] {
			out[k] = "***redacted***"
			continue
		}
		out[k] = v
	}
	return out
}

func redactHTTPHeaders(headers http.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		if redactedHeaderKeys[strings.ToLower(k)] {
			out[k] = "***redacted***"
			continue
		}
		out[k] = strings.Join(v, ",")
	}
	return out
}
