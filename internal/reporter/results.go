package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"wafguard/internal/logger"
	"wafguard/internal/core/config"
	"wafguard/internal/executor"
	"wafguard/internal/validator"

	"github.com/sirupsen/logrus"
)

type TestReport struct {
	TestName         string                   `json:"test_name"`
	Status           string                   `json:"status"`
	Duration         time.Duration            `json:"duration"`
	Request          *config.Request          `json:"request"`
	Response         *executor.Response       `json:"response"`
	ValidationResult *validator.ValidationResult `json:"validation_result"`
	Timestamp        time.Time                `json:"timestamp"`
}

type SuiteReport struct {
	SuiteName    string        `json:"suite_name"`
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	Duration     time.Duration `json:"duration"`
	Tests        []TestReport  `json:"tests"`
	Timestamp    time.Time     `json:"timestamp"`
}

type Reporter struct {
	format string
	output string
}

func NewReporter(format, output string) *Reporter {
	return &Reporter{
		format: format,
		output: output,
	}
}

func (r *Reporter) GenerateTestReport(testName string, request *config.Request, response *executor.Response, validation *validator.ValidationResult, duration time.Duration) *TestReport {
	status := "PASS"
	if !validation.Passed {
		status = "FAIL"
	}

	return &TestReport{
		TestName:         testName,
		Status:           status,
		Duration:         duration,
		Request:          request,
		Response:         response,
		ValidationResult: validation,
		Timestamp:        time.Now(),
	}
}

func (r *Reporter) GenerateSuiteReport(suiteName string, testReports []TestReport, totalDuration time.Duration) *SuiteReport {
	passed := 0
	failed := 0

	for _, report := range testReports {
		if report.Status == "PASS" {
			passed++
		} else {
			failed++
		}
	}

	return &SuiteReport{
		SuiteName:   suiteName,
		TotalTests:  len(testReports),
		PassedTests: passed,
		FailedTests: failed,
		Duration:    totalDuration,
		Tests:       testReports,
		Timestamp:   time.Now(),
	}
}

func (r *Reporter) PrintTestReport(report *TestReport) {
	switch r.format {
	case "json":
		r.printJSONTestReport(report)
	case "text":
		r.printTextTestReport(report)
	default:
		r.printTextTestReport(report)
	}
}

func (r *Reporter) PrintSuiteReport(report *SuiteReport) {
	switch r.format {
	case "json":
		r.printJSONSuiteReport(report)
	case "text":
		r.printTextSuiteReport(report)
	default:
		r.printTextSuiteReport(report)
	}
}

func (r *Reporter) SaveSuiteReport(report *SuiteReport) error {
	if r.output == "" {
		return nil
	}

	var data []byte
	var err error

	switch r.format {
	case "json":
		data, err = json.MarshalIndent(report, "", "  ")
	default:
		data, err = json.MarshalIndent(report, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(r.output, data, 0644); err != nil {
		return fmt.Errorf("failed to write report to file: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"file":   r.output,
		"format": r.format,
	}).Info("Report saved to file")

	return nil
}

func (r *Reporter) printJSONTestReport(report *TestReport) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal test report to JSON:", err)
		return
	}
	fmt.Println(string(data))
}

func (r *Reporter) printTextTestReport(report *TestReport) {
	fmt.Printf("Test: %s\n", report.TestName)
	fmt.Printf("Status: %s\n", report.Status)
	fmt.Printf("Duration: %s\n", report.Duration)
	fmt.Printf("Request: %s %s\n", report.Request.Method, report.Request.Path)
	fmt.Printf("Response Status: %d\n", report.Response.StatusCode)
	
	if len(report.ValidationResult.Errors) > 0 {
		fmt.Println("Validation Errors:")
		for _, err := range report.ValidationResult.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
	
	if len(report.ValidationResult.Warnings) > 0 {
		fmt.Println("Validation Warnings:")
		for _, warn := range report.ValidationResult.Warnings {
			fmt.Printf("  - %s\n", warn)
		}
	}
	
	fmt.Println("---")
}

func (r *Reporter) printJSONSuiteReport(report *SuiteReport) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal suite report to JSON:", err)
		return
	}
	fmt.Println(string(data))
}

func (r *Reporter) printTextSuiteReport(report *SuiteReport) {
	fmt.Printf("Suite: %s\n", report.SuiteName)
	fmt.Printf("Total Tests: %d\n", report.TotalTests)
	fmt.Printf("Passed: %d\n", report.PassedTests)
	fmt.Printf("Failed: %d\n", report.FailedTests)
	fmt.Printf("Duration: %s\n", report.Duration)
	fmt.Printf("Success Rate: %.2f%%\n", float64(report.PassedTests)/float64(report.TotalTests)*100)
	fmt.Println("====================================")
	
	for _, test := range report.Tests {
		r.printTextTestReport(&test)
	}
}