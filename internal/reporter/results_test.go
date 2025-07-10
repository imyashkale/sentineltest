package reporter

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
	"wafguard/internal/core/config"
	"wafguard/internal/executor"
	"wafguard/internal/validator"
)

func TestNewReporter(t *testing.T) {
	tests := []struct {
		name   string
		format string
		output string
	}{
		{
			name:   "json format",
			format: "json",
			output: "test.json",
		},
		{
			name:   "text format",
			format: "text",
			output: "test.txt",
		},
		{
			name:   "no output file",
			format: "json",
			output: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewReporter(tt.format, tt.output)
			if reporter == nil {
				t.Fatal("NewReporter() returned nil")
			}
			if reporter.format != tt.format {
				t.Errorf("NewReporter() format = %s, want %s", reporter.format, tt.format)
			}
			if reporter.output != tt.output {
				t.Errorf("NewReporter() output = %s, want %s", reporter.output, tt.output)
			}
		})
	}
}

func TestGenerateTestReport(t *testing.T) {
	reporter := NewReporter("json", "")

	request := &config.Request{
		Method: "GET",
		Path:   "/test",
		Headers: map[string]string{
			"User-Agent": "test-agent",
		},
	}

	response := &executor.Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:     `{"status": "success"}`,
		Duration: 100 * time.Millisecond,
	}

	validation := &validator.ValidationResult{
		Passed:   true,
		Errors:   []string{},
		Warnings: []string{},
	}

	duration := 150 * time.Millisecond

	report := reporter.GenerateTestReport("test-case", request, response, validation, duration)

	if report == nil {
		t.Fatal("GenerateTestReport() returned nil")
	}

	if report.TestName != "test-case" {
		t.Errorf("GenerateTestReport() TestName = %s, want test-case", report.TestName)
	}

	if report.Status != "PASS" {
		t.Errorf("GenerateTestReport() Status = %s, want PASS", report.Status)
	}

	if report.Duration != duration {
		t.Errorf("GenerateTestReport() Duration = %v, want %v", report.Duration, duration)
	}

	if report.Request != request {
		t.Error("GenerateTestReport() Request pointer mismatch")
	}

	if report.Response != response {
		t.Error("GenerateTestReport() Response pointer mismatch")
	}

	if report.ValidationResult != validation {
		t.Error("GenerateTestReport() ValidationResult pointer mismatch")
	}

	if report.Timestamp.IsZero() {
		t.Error("GenerateTestReport() Timestamp should not be zero")
	}
}

func TestGenerateTestReportFail(t *testing.T) {
	reporter := NewReporter("json", "")

	request := &config.Request{
		Method: "POST",
		Path:   "/api/test",
	}

	response := &executor.Response{
		StatusCode: 404,
		Body:       `{"error": "not found"}`,
		Duration:   50 * time.Millisecond,
	}

	validation := &validator.ValidationResult{
		Passed: false,
		Errors: []string{
			"Status code mismatch: expected one of [200], got 404",
			"Body should contain 'success' but it was not found",
		},
		Warnings: []string{},
	}

	report := reporter.GenerateTestReport("fail-test", request, response, validation, 75*time.Millisecond)

	if report.Status != "FAIL" {
		t.Errorf("GenerateTestReport() Status = %s, want FAIL", report.Status)
	}

	if len(report.ValidationResult.Errors) != 2 {
		t.Errorf("GenerateTestReport() ValidationResult.Errors count = %d, want 2", len(report.ValidationResult.Errors))
	}
}

func TestGenerateSuiteReport(t *testing.T) {
	reporter := NewReporter("json", "")

	testReports := []TestReport{
		{
			TestName: "test1",
			Status:   "PASS",
			Duration: 100 * time.Millisecond,
		},
		{
			TestName: "test2",
			Status:   "FAIL",
			Duration: 200 * time.Millisecond,
		},
		{
			TestName: "test3",
			Status:   "PASS",
			Duration: 150 * time.Millisecond,
		},
	}

	totalDuration := 500 * time.Millisecond

	suiteReport := reporter.GenerateSuiteReport("test-suite", testReports, totalDuration)

	if suiteReport == nil {
		t.Fatal("GenerateSuiteReport() returned nil")
	}

	if suiteReport.SuiteName != "test-suite" {
		t.Errorf("GenerateSuiteReport() SuiteName = %s, want test-suite", suiteReport.SuiteName)
	}

	if suiteReport.TotalTests != 3 {
		t.Errorf("GenerateSuiteReport() TotalTests = %d, want 3", suiteReport.TotalTests)
	}

	if suiteReport.PassedTests != 2 {
		t.Errorf("GenerateSuiteReport() PassedTests = %d, want 2", suiteReport.PassedTests)
	}

	if suiteReport.FailedTests != 1 {
		t.Errorf("GenerateSuiteReport() FailedTests = %d, want 1", suiteReport.FailedTests)
	}

	if suiteReport.Duration != totalDuration {
		t.Errorf("GenerateSuiteReport() Duration = %v, want %v", suiteReport.Duration, totalDuration)
	}

	if len(suiteReport.Tests) != 3 {
		t.Errorf("GenerateSuiteReport() Tests count = %d, want 3", len(suiteReport.Tests))
	}

	if suiteReport.Timestamp.IsZero() {
		t.Error("GenerateSuiteReport() Timestamp should not be zero")
	}
}

func TestGenerateSuiteReportEmpty(t *testing.T) {
	reporter := NewReporter("text", "")

	suiteReport := reporter.GenerateSuiteReport("empty-suite", []TestReport{}, 0)

	if suiteReport.TotalTests != 0 {
		t.Errorf("GenerateSuiteReport() TotalTests = %d, want 0", suiteReport.TotalTests)
	}

	if suiteReport.PassedTests != 0 {
		t.Errorf("GenerateSuiteReport() PassedTests = %d, want 0", suiteReport.PassedTests)
	}

	if suiteReport.FailedTests != 0 {
		t.Errorf("GenerateSuiteReport() FailedTests = %d, want 0", suiteReport.FailedTests)
	}
}

func TestSaveSuiteReport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "reporter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := tmpDir + "/test-report.json"
	reporter := NewReporter("json", outputFile)

	suiteReport := &SuiteReport{
		SuiteName:   "test-suite",
		TotalTests:  2,
		PassedTests: 1,
		FailedTests: 1,
		Duration:    300 * time.Millisecond,
		Tests: []TestReport{
			{
				TestName: "test1",
				Status:   "PASS",
			},
			{
				TestName: "test2",
				Status:   "FAIL",
			},
		},
		Timestamp: time.Now(),
	}

	err = reporter.SaveSuiteReport(suiteReport)
	if err != nil {
		t.Fatalf("SaveSuiteReport() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("SaveSuiteReport() did not create output file")
	}

	// Verify file content
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var savedReport SuiteReport
	err = json.Unmarshal(data, &savedReport)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved report: %v", err)
	}

	if savedReport.SuiteName != suiteReport.SuiteName {
		t.Errorf("Saved report SuiteName = %s, want %s", savedReport.SuiteName, suiteReport.SuiteName)
	}

	if savedReport.TotalTests != suiteReport.TotalTests {
		t.Errorf("Saved report TotalTests = %d, want %d", savedReport.TotalTests, suiteReport.TotalTests)
	}
}

func TestSaveSuiteReportNoOutput(t *testing.T) {
	reporter := NewReporter("json", "") // No output file

	suiteReport := &SuiteReport{
		SuiteName: "test-suite",
	}

	err := reporter.SaveSuiteReport(suiteReport)
	if err != nil {
		t.Errorf("SaveSuiteReport() should not fail when no output file specified: %v", err)
	}
}

func TestSaveSuiteReportInvalidPath(t *testing.T) {
	reporter := NewReporter("json", "/invalid/path/that/does/not/exist/report.json")

	suiteReport := &SuiteReport{
		SuiteName: "test-suite",
	}

	err := reporter.SaveSuiteReport(suiteReport)
	if err == nil {
		t.Error("SaveSuiteReport() should fail with invalid path")
	}
}

func TestPrintTestReportJSON(t *testing.T) {
	// Capture stdout for testing
	// Note: In a real test environment, you might want to use a more sophisticated approach
	// to capture and verify output. For now, we'll just test that the methods don't panic.

	reporter := NewReporter("json", "")

	testReport := &TestReport{
		TestName: "json-test",
		Status:   "PASS",
		Duration: 100 * time.Millisecond,
		Request: &config.Request{
			Method: "GET",
			Path:   "/test",
		},
		Response: &executor.Response{
			StatusCode: 200,
			Body:       `{"status": "ok"}`,
		},
		ValidationResult: &validator.ValidationResult{
			Passed: true,
		},
		Timestamp: time.Now(),
	}

	// Test that PrintTestReport doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTestReport() panicked: %v", r)
		}
	}()

	reporter.PrintTestReport(testReport)
}

func TestPrintTestReportText(t *testing.T) {
	reporter := NewReporter("text", "")

	testReport := &TestReport{
		TestName: "text-test",
		Status:   "FAIL",
		Duration: 200 * time.Millisecond,
		Request: &config.Request{
			Method: "POST",
			Path:   "/api/test",
		},
		Response: &executor.Response{
			StatusCode: 404,
			Body:       `{"error": "not found"}`,
		},
		ValidationResult: &validator.ValidationResult{
			Passed: false,
			Errors: []string{
				"Status code mismatch",
				"Body validation failed",
			},
			Warnings: []string{
				"Response time high",
			},
		},
		Timestamp: time.Now(),
	}

	// Test that PrintTestReport doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTestReport() panicked: %v", r)
		}
	}()

	reporter.PrintTestReport(testReport)
}

func TestPrintSuiteReportJSON(t *testing.T) {
	reporter := NewReporter("json", "")

	suiteReport := &SuiteReport{
		SuiteName:   "json-suite",
		TotalTests:  3,
		PassedTests: 2,
		FailedTests: 1,
		Duration:    500 * time.Millisecond,
		Tests: []TestReport{
			{TestName: "test1", Status: "PASS"},
			{TestName: "test2", Status: "PASS"},
			{TestName: "test3", Status: "FAIL"},
		},
		Timestamp: time.Now(),
	}

	// Test that PrintSuiteReport doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintSuiteReport() panicked: %v", r)
		}
	}()

	reporter.PrintSuiteReport(suiteReport)
}

func TestPrintSuiteReportText(t *testing.T) {
	reporter := NewReporter("text", "")

	suiteReport := &SuiteReport{
		SuiteName:   "text-suite",
		TotalTests:  2,
		PassedTests: 1,
		FailedTests: 1,
		Duration:    300 * time.Millisecond,
		Tests: []TestReport{
			{
				TestName: "passing-test",
				Status:   "PASS",
				Duration: 100 * time.Millisecond,
				Request: &config.Request{
					Method: "GET",
					Path:   "/test",
				},
				Response: &executor.Response{
					StatusCode: 200,
				},
				ValidationResult: &validator.ValidationResult{
					Passed: true,
				},
			},
			{
				TestName: "failing-test",
				Status:   "FAIL",
				Duration: 200 * time.Millisecond,
				Request: &config.Request{
					Method: "POST",
					Path:   "/fail",
				},
				Response: &executor.Response{
					StatusCode: 500,
				},
				ValidationResult: &validator.ValidationResult{
					Passed: false,
					Errors: []string{"Server error"},
				},
			},
		},
		Timestamp: time.Now(),
	}

	// Test that PrintSuiteReport doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintSuiteReport() panicked: %v", r)
		}
	}()

	reporter.PrintSuiteReport(suiteReport)
}

func TestReportFormatHandling(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "json format",
			format: "json",
		},
		{
			name:   "text format",
			format: "text",
		},
		{
			name:   "unknown format defaults to text",
			format: "unknown",
		},
		{
			name:   "empty format defaults to text",
			format: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewReporter(tt.format, "")

			// Create a simple test report
			testReport := &TestReport{
				TestName: "format-test",
				Status:   "PASS",
				Request: &config.Request{
					Method: "GET",
					Path:   "/test",
				},
				Response: &executor.Response{
					StatusCode: 200,
				},
				ValidationResult: &validator.ValidationResult{
					Passed: true,
				},
			}

			// Test that both print methods work without panicking
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Print methods panicked with format %s: %v", tt.format, r)
				}
			}()

			reporter.PrintTestReport(testReport)

			suiteReport := &SuiteReport{
				SuiteName: "format-suite",
				Tests:     []TestReport{*testReport},
			}
			reporter.PrintSuiteReport(suiteReport)
		})
	}
}

func TestJSONMarshalError(t *testing.T) {
	// Test error handling in printJSONTestReport
	reporter := NewReporter("json", "")

	// Create a test report with a channel (which can't be marshaled to JSON)
	// Note: This is a contrived example since our structs don't contain channels
	// In a real scenario, you might test with circular references or other unmarshalable data

	// For this test, we'll just verify the method handles the report correctly
	testReport := &TestReport{
		TestName: "marshal-test",
		Status:   "PASS",
		Request: &config.Request{
			Method: "GET",
			Path:   "/test",
		},
		Response: &executor.Response{
			StatusCode: 200,
			Body:       strings.Repeat("a", 10000), // Large body
		},
		ValidationResult: &validator.ValidationResult{
			Passed: true,
		},
	}

	// Should not panic even with large data
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printJSONTestReport() panicked: %v", r)
		}
	}()

	reporter.PrintTestReport(testReport)
}