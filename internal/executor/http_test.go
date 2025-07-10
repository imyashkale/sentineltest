package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"wafguard/internal/core/config"
)

func TestNewHTTPExecutor(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:           "default timeout",
			timeout:        0,
			expectedTimeout: 30 * time.Second,
		},
		{
			name:           "custom timeout",
			timeout:        10 * time.Second,
			expectedTimeout: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewHTTPExecutor(tt.timeout)
			if executor == nil {
				t.Fatal("NewHTTPExecutor() returned nil")
			}
			if executor.client == nil {
				t.Fatal("HTTPExecutor client is nil")
			}
			if executor.client.Timeout != tt.expectedTimeout {
				t.Errorf("HTTPExecutor timeout = %v, want %v", executor.client.Timeout, tt.expectedTimeout)
			}
		})
	}
}

func TestBuildURL(t *testing.T) {
	executor := NewHTTPExecutor(30 * time.Second)

	tests := []struct {
		name    string
		baseURL string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "simple path",
			baseURL: "https://example.com",
			path:    "/test",
			want:    "https://example.com/test",
			wantErr: false,
		},
		{
			name:    "path with query params",
			baseURL: "https://example.com",
			path:    "/test?param=value",
			want:    "https://example.com/test?param=value",
			wantErr: false,
		},
		{
			name:    "absolute path",
			baseURL: "https://example.com/api",
			path:    "/test",
			want:    "https://example.com/test",
			wantErr: false,
		},
		{
			name:    "relative path",
			baseURL: "https://example.com/api/",
			path:    "test",
			want:    "https://example.com/api/test",
			wantErr: false,
		},
		{
			name:    "invalid base URL",
			baseURL: "://not-a-url",
			path:    "/test",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid path",
			baseURL: "https://example.com",
			path:    "://invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executor.buildURL(tt.baseURL, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateRequest(t *testing.T) {
	executor := NewHTTPExecutor(30 * time.Second)

	tests := []struct {
		name       string
		reqConfig  config.Request
		url        string
		wantMethod string
		wantURL    string
		wantErr    bool
	}{
		{
			name: "GET request without body",
			reqConfig: config.Request{
				Method: "GET",
				Path:   "/test",
				Headers: map[string]string{
					"User-Agent": "test-agent",
				},
			},
			url:        "https://example.com/test",
			wantMethod: "GET",
			wantURL:    "https://example.com/test",
			wantErr:    false,
		},
		{
			name: "POST request with body",
			reqConfig: config.Request{
				Method: "POST",
				Path:   "/test",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: `{"key": "value"}`,
			},
			url:        "https://example.com/test",
			wantMethod: "POST",
			wantURL:    "https://example.com/test",
			wantErr:    false,
		},
		{
			name: "invalid method",
			reqConfig: config.Request{
				Method: "INVALID METHOD",
				Path:   "/test",
			},
			url:     "https://example.com/test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := executor.createRequest(tt.reqConfig, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("createRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if req.Method != tt.wantMethod {
					t.Errorf("createRequest() method = %v, want %v", req.Method, tt.wantMethod)
				}
				if req.URL.String() != tt.wantURL {
					t.Errorf("createRequest() URL = %v, want %v", req.URL.String(), tt.wantURL)
				}

				// Check headers
				for key, value := range tt.reqConfig.Headers {
					if req.Header.Get(key) != value {
						t.Errorf("createRequest() header %s = %v, want %v", key, req.Header.Get(key), value)
					}
				}
			}
		})
	}
}

func TestExtractHeaders(t *testing.T) {
	executor := NewHTTPExecutor(30 * time.Second)

	headers := http.Header{
		"Content-Type": []string{"application/json"},
		"Set-Cookie":   []string{"session=abc123", "user=test"},
		"X-Custom":     []string{"value1"},
	}

	result := executor.extractHeaders(headers)

	expected := map[string]string{
		"Content-Type": "application/json",
		"Set-Cookie":   "session=abc123, user=test",
		"X-Custom":     "value1",
	}

	for key, value := range expected {
		if result[key] != value {
			t.Errorf("extractHeaders() %s = %v, want %v", key, result[key], value)
		}
	}
}

func TestExecuteTest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success", "method": "` + r.Method + `"}`))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(30 * time.Second)

	test := &config.Test{
		Name: "test-request",
		Request: config.Request{
			Method: "GET",
			Path:   "/test",
			Headers: map[string]string{
				"User-Agent": "test-agent",
			},
		},
	}

	response, err := executor.ExecuteTest(test, server.URL)
	if err != nil {
		t.Fatalf("ExecuteTest() failed: %v", err)
	}

	if response == nil {
		t.Fatal("ExecuteTest() returned nil response")
	}

	if response.StatusCode != 200 {
		t.Errorf("ExecuteTest() status code = %d, want 200", response.StatusCode)
	}

	if response.Headers["Content-Type"] != "application/json" {
		t.Errorf("ExecuteTest() Content-Type = %s, want application/json", response.Headers["Content-Type"])
	}

	if response.Headers["X-Test-Header"] != "test-value" {
		t.Errorf("ExecuteTest() X-Test-Header = %s, want test-value", response.Headers["X-Test-Header"])
	}

	if !contains(response.Body, "success") {
		t.Errorf("ExecuteTest() body should contain 'success', got: %s", response.Body)
	}

	if response.Duration <= 0 {
		t.Error("ExecuteTest() duration should be positive")
	}
}

func TestExecuteTestWithContext(t *testing.T) {
	// Create a test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Small delay
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(30 * time.Second)

	test := &config.Test{
		Name: "test-request",
		Request: config.Request{
			Method: "GET",
			Path:   "/test",
		},
	}

	// Test with normal context
	ctx := context.Background()
	response, err := executor.ExecuteTestWithContext(ctx, test, server.URL)
	if err != nil {
		t.Fatalf("ExecuteTestWithContext() failed: %v", err)
	}

	if response.StatusCode != 200 {
		t.Errorf("ExecuteTestWithContext() status code = %d, want 200", response.StatusCode)
	}

	// Test with cancelled context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = executor.ExecuteTestWithContext(cancelCtx, test, server.URL)
	if err == nil {
		t.Error("ExecuteTestWithContext() should fail with cancelled context")
	}
}

func TestExecuteTestServerError(t *testing.T) {
	// Create a test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(30 * time.Second)

	test := &config.Test{
		Name: "error-test",
		Request: config.Request{
			Method: "GET",
			Path:   "/error",
		},
	}

	response, err := executor.ExecuteTest(test, server.URL)
	if err != nil {
		t.Fatalf("ExecuteTest() failed: %v", err)
	}

	if response.StatusCode != 500 {
		t.Errorf("ExecuteTest() status code = %d, want 500", response.StatusCode)
	}

	if !contains(response.Body, "error") {
		t.Errorf("ExecuteTest() body should contain 'error', got: %s", response.Body)
	}
}

func TestExecuteTestWithBody(t *testing.T) {
	// Create a test server that echoes the request body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"received": "` + string(body) + `"}`))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(30 * time.Second)

	requestBody := `{"test": "data"}`
	test := &config.Test{
		Name: "post-test",
		Request: config.Request{
			Method: "POST",
			Path:   "/echo",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: requestBody,
		},
	}

	response, err := executor.ExecuteTest(test, server.URL)
	if err != nil {
		t.Fatalf("ExecuteTest() failed: %v", err)
	}

	if response.StatusCode != 200 {
		t.Errorf("ExecuteTest() status code = %d, want 200", response.StatusCode)
	}

	if !contains(response.Body, requestBody) {
		t.Errorf("ExecuteTest() body should contain request data, got: %s", response.Body)
	}
}

func TestExecuteTestInvalidURL(t *testing.T) {
	executor := NewHTTPExecutor(30 * time.Second)

	test := &config.Test{
		Name: "invalid-url-test",
		Request: config.Request{
			Method: "GET",
			Path:   "/test",
		},
	}

	// Test with invalid base URL
	_, err := executor.ExecuteTest(test, "invalid-url")
	if err == nil {
		t.Error("ExecuteTest() should fail with invalid base URL")
	}

	// Test with unreachable URL
	_, err = executor.ExecuteTest(test, "http://localhost:99999")
	if err == nil {
		t.Error("ExecuteTest() should fail with unreachable URL")
	}
}

func TestExecuteTestTimeout(t *testing.T) {
	// Create a test server with long delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than our timeout
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Create executor with very short timeout
	executor := NewHTTPExecutor(100 * time.Millisecond)

	test := &config.Test{
		Name: "timeout-test",
		Request: config.Request{
			Method: "GET",
			Path:   "/slow",
		},
	}

	_, err := executor.ExecuteTest(test, server.URL)
	if err == nil {
		t.Error("ExecuteTest() should fail with timeout")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 indexContains(s, substr) >= 0)))
}

func indexContains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}