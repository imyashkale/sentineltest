package validator

import (
	"testing"
	"time"
	"waf-tester/pkg/config"
	"waf-tester/pkg/executor"
)

func TestNewResponseValidator(t *testing.T) {
	validator := NewResponseValidator()
	if validator == nil {
		t.Fatal("NewResponseValidator() returned nil")
	}
}

func TestValidateStatusCode(t *testing.T) {
	validator := NewResponseValidator()

	tests := []struct {
		name           string
		response       *executor.Response
		expected       *config.Expected
		wantPassed     bool
		wantErrorCount int
		wantWarnCount  int
	}{
		{
			name: "status code match",
			response: &executor.Response{
				StatusCode: 200,
			},
			expected: &config.Expected{
				Status: []int{200, 201},
			},
			wantPassed:     true,
			wantErrorCount: 0,
			wantWarnCount:  0,
		},
		{
			name: "status code mismatch",
			response: &executor.Response{
				StatusCode: 404,
			},
			expected: &config.Expected{
				Status: []int{200, 201},
			},
			wantPassed:     false,
			wantErrorCount: 1,
			wantWarnCount:  0,
		},
		{
			name: "no expected status codes",
			response: &executor.Response{
				StatusCode: 200,
			},
			expected: &config.Expected{
				Status: []int{},
			},
			wantPassed:     true,
			wantErrorCount: 0,
			wantWarnCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.response, tt.expected, "test")

			if result.Passed != tt.wantPassed {
				t.Errorf("Validate() passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if len(result.Errors) != tt.wantErrorCount {
				t.Errorf("Validate() error count = %d, want %d", len(result.Errors), tt.wantErrorCount)
			}

			if len(result.Warnings) != tt.wantWarnCount {
				t.Errorf("Validate() warning count = %d, want %d", len(result.Warnings), tt.wantWarnCount)
			}
		})
	}
}

func TestValidateHeaders(t *testing.T) {
	validator := NewResponseValidator()

	tests := []struct {
		name           string
		response       *executor.Response
		expected       *config.Expected
		wantPassed     bool
		wantErrorCount int
	}{
		{
			name: "header match",
			response: &executor.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"X-Custom":     "test-value",
				},
			},
			expected: &config.Expected{
				Status: []int{200},
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
		{
			name: "header partial match",
			response: &executor.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json; charset=utf-8",
				},
			},
			expected: &config.Expected{
				Status: []int{200},
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
		{
			name: "missing header",
			response: &executor.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			expected: &config.Expected{
				Status: []int{200},
				Headers: map[string]string{
					"X-Missing": "value",
				},
			},
			wantPassed:     false,
			wantErrorCount: 1,
		},
		{
			name: "header value mismatch",
			response: &executor.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/html",
				},
			},
			expected: &config.Expected{
				Status: []int{200},
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantPassed:     false,
			wantErrorCount: 1,
		},
		{
			name: "no expected headers",
			response: &executor.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			expected: &config.Expected{
				Status: []int{200},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.response, tt.expected, "test")

			if result.Passed != tt.wantPassed {
				t.Errorf("Validate() passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if len(result.Errors) != tt.wantErrorCount {
				t.Errorf("Validate() error count = %d, want %d. Errors: %v", len(result.Errors), tt.wantErrorCount, result.Errors)
			}
		})
	}
}

func TestValidateBody(t *testing.T) {
	validator := NewResponseValidator()

	tests := []struct {
		name           string
		response       *executor.Response
		expected       *config.Expected
		wantPassed     bool
		wantErrorCount int
	}{
		{
			name: "body contains match",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "success", "data": "test"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Contains: []string{"success", "data"},
				},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
		{
			name: "body contains mismatch",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "error"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Contains: []string{"success"},
				},
			},
			wantPassed:     false,
			wantErrorCount: 1,
		},
		{
			name: "body not contains match",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "success"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					NotContains: []string{"error", "fail"},
				},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
		{
			name: "body not contains mismatch",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "error occurred"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					NotContains: []string{"error"},
				},
			},
			wantPassed:     false,
			wantErrorCount: 1,
		},
		{
			name: "body exact match",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"exact": "response"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Exact: `{"exact": "response"}`,
				},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
		{
			name: "body exact mismatch",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"actual": "response"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Exact: `{"expected": "response"}`,
				},
			},
			wantPassed:     false,
			wantErrorCount: 1,
		},
		{
			name: "body regex match",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "success", "id": 123}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Regex: `^\{.*"id":\s*\d+.*\}$`,
				},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
		{
			name: "body regex mismatch",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "success"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Regex: `^\{.*"id":\s*\d+.*\}$`,
				},
			},
			wantPassed:     false,
			wantErrorCount: 1,
		},
		{
			name: "invalid regex",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "success"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Regex: `[invalid regex`,
				},
			},
			wantPassed:     false,
			wantErrorCount: 1,
		},
		{
			name: "combined body validation",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "success", "status": "ok"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Contains:    []string{"success", "status"},
					NotContains: []string{"error", "fail"},
				},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
		{
			name: "combined body validation with errors",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"message": "error", "status": "fail"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Contains:    []string{"success"},
					NotContains: []string{"error"},
				},
			},
			wantPassed:     false,
			wantErrorCount: 2,
		},
		{
			name: "no body validation",
			response: &executor.Response{
				StatusCode: 200,
				Body:       `{"any": "content"}`,
			},
			expected: &config.Expected{
				Status: []int{200},
			},
			wantPassed:     true,
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.response, tt.expected, "test")

			if result.Passed != tt.wantPassed {
				t.Errorf("Validate() passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if len(result.Errors) != tt.wantErrorCount {
				t.Errorf("Validate() error count = %d, want %d. Errors: %v", len(result.Errors), tt.wantErrorCount, result.Errors)
			}
		})
	}
}

func TestValidateMultiple(t *testing.T) {
	validator := NewResponseValidator()

	responses := []*executor.Response{
		{
			StatusCode: 200,
			Body:       `{"status": "success"}`,
		},
		{
			StatusCode: 404,
			Body:       `{"status": "not found"}`,
		},
		{
			StatusCode: 200,
			Body:       `{"status": "ok"}`,
		},
	}

	expected := []*config.Expected{
		{
			Status: []int{200},
			Body: &config.BodyExpected{
				Contains: []string{"success"},
			},
		},
		{
			Status: []int{200},
			Body: &config.BodyExpected{
				Contains: []string{"success"},
			},
		},
		{
			Status: []int{200, 201},
			Body: &config.BodyExpected{
				Contains: []string{"ok"},
			},
		},
	}

	testNames := []string{"test1", "test2", "test3"}

	results := validator.ValidateMultiple(responses, expected, testNames)

	if len(results) != 3 {
		t.Fatalf("ValidateMultiple() returned %d results, want 3", len(results))
	}

	// First test should pass
	if !results[0].Passed {
		t.Error("ValidateMultiple() result[0] should pass")
	}

	// Second test should fail (status code mismatch)
	if results[1].Passed {
		t.Error("ValidateMultiple() result[1] should fail")
	}

	// Third test should pass
	if !results[2].Passed {
		t.Error("ValidateMultiple() result[2] should pass")
	}
}

func TestValidateMultiplePanic(t *testing.T) {
	validator := NewResponseValidator()

	defer func() {
		if r := recover(); r == nil {
			t.Error("ValidateMultiple() should panic with mismatched lengths")
		}
	}()

	responses := []*executor.Response{
		{StatusCode: 200},
	}
	expected := []*config.Expected{
		{Status: []int{200}},
		{Status: []int{201}},
	}
	testNames := []string{"test1"}

	validator.ValidateMultiple(responses, expected, testNames)
}

func TestValidationResultStructure(t *testing.T) {
	validator := NewResponseValidator()

	response := &executor.Response{
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body:     `{"error": "not found"}`,
		Duration: 100 * time.Millisecond,
	}

	expected := &config.Expected{
		Status: []int{200},
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Missing":    "value",
		},
		Body: &config.BodyExpected{
			Contains:    []string{"success"},
			NotContains: []string{"error"},
		},
	}

	result := validator.Validate(response, expected, "complex-test")

	// Should fail with multiple errors
	if result.Passed {
		t.Error("Validation should fail")
	}

	expectedErrors := 5 // status code + 2 headers + 2 body issues
	if len(result.Errors) != expectedErrors {
		t.Errorf("Expected %d errors, got %d: %v", expectedErrors, len(result.Errors), result.Errors)
	}

	// Check that all errors are meaningful
	for i, err := range result.Errors {
		if err == "" {
			t.Errorf("Error[%d] is empty", i)
		}
	}
}

func TestValidateWithComplexRegex(t *testing.T) {
	validator := NewResponseValidator()

	tests := []struct {
		name       string
		body       string
		regex      string
		wantPassed bool
	}{
		{
			name:       "email regex match",
			body:       `{"email": "test@example.com"}`,
			regex:      `"email":\s*"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}"`,
			wantPassed: true,
		},
		{
			name:       "email regex mismatch",
			body:       `{"email": "invalid-email"}`,
			regex:      `"email":\s*"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}"`,
			wantPassed: false,
		},
		{
			name:       "JSON structure regex",
			body:       `{"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]}`,
			regex:      `^\{.*"users":\s*\[.*\{.*"id":\s*\d+.*\}.*\].*\}$`,
			wantPassed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &executor.Response{
				StatusCode: 200,
				Body:       tt.body,
			}

			expected := &config.Expected{
				Status: []int{200},
				Body: &config.BodyExpected{
					Regex: tt.regex,
				},
			}

			result := validator.Validate(response, expected, tt.name)

			if result.Passed != tt.wantPassed {
				t.Errorf("Validate() passed = %v, want %v. Errors: %v", result.Passed, tt.wantPassed, result.Errors)
			}
		})
	}
}