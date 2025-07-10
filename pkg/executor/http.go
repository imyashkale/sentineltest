package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"waf-tester/internal/logger"
	"waf-tester/pkg/config"

	"github.com/sirupsen/logrus"
)

type HTTPExecutor struct {
	client *http.Client
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	Duration   time.Duration
}

func NewHTTPExecutor(timeout time.Duration) *HTTPExecutor {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPExecutor{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (e *HTTPExecutor) ExecuteTest(test *config.Test, baseURL string) (*Response, error) {
	start := time.Now()
	
	logger.WithFields(logrus.Fields{
		"test_name": test.Name,
		"method":    test.Request.Method,
		"path":      test.Request.Path,
	}).Info("Executing test")

	fullURL, err := e.buildURL(baseURL, test.Request.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := e.createRequest(test.Request, fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	duration := time.Since(start)

	response := &Response{
		StatusCode: resp.StatusCode,
		Headers:    e.extractHeaders(resp.Header),
		Body:       string(body),
		Duration:   duration,
	}

	logger.WithFields(logrus.Fields{
		"test_name":   test.Name,
		"status_code": response.StatusCode,
		"duration":    duration.String(),
	}).Info("Test executed")

	return response, nil
}

func (e *HTTPExecutor) ExecuteTestWithContext(ctx context.Context, test *config.Test, baseURL string) (*Response, error) {
	start := time.Now()
	
	logger.WithFields(logrus.Fields{
		"test_name": test.Name,
		"method":    test.Request.Method,
		"path":      test.Request.Path,
	}).Info("Executing test with context")

	fullURL, err := e.buildURL(baseURL, test.Request.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := e.createRequest(test.Request, fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req = req.WithContext(ctx)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	duration := time.Since(start)

	response := &Response{
		StatusCode: resp.StatusCode,
		Headers:    e.extractHeaders(resp.Header),
		Body:       string(body),
		Duration:   duration,
	}

	logger.WithFields(logrus.Fields{
		"test_name":   test.Name,
		"status_code": response.StatusCode,
		"duration":    duration.String(),
	}).Info("Test executed with context")

	return response, nil
}

func (e *HTTPExecutor) buildURL(baseURL, path string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	rel, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	return base.ResolveReference(rel).String(), nil
}

func (e *HTTPExecutor) createRequest(reqConfig config.Request, url string) (*http.Request, error) {
	var body io.Reader
	if reqConfig.Body != "" {
		body = bytes.NewBufferString(reqConfig.Body)
	}

	req, err := http.NewRequest(reqConfig.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range reqConfig.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (e *HTTPExecutor) extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		result[key] = strings.Join(values, ", ")
	}
	return result
}