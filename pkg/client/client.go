// Package client provides a public API for the WAF testing functionality.
package client

import (
	"context"
	"time"
	"wafguard/internal/core/config"
	"wafguard/internal/executor"
	"wafguard/internal/parser"
	"wafguard/internal/reporter"
	"wafguard/internal/validator"
)

// Client represents a WAF testing client
type Client struct {
	parser    *parser.Parser
	executor  *executor.HTTPExecutor
	validator *validator.ResponseValidator
	reporter  *reporter.Reporter
}

// Config represents the client configuration
type Config struct {
	Timeout     time.Duration
	OutputFile  string
	Format      string // "json" or "text"
	Concurrent  int
}

// TestResult represents the result of running a test
type TestResult struct {
	TestName string
	Passed   bool
	Duration time.Duration
	Errors   []string
	Warnings []string
}

// SuiteResult represents the result of running a test suite
type SuiteResult struct {
	SuiteName    string
	TotalTests   int
	PassedTests  int
	FailedTests  int
	Duration     time.Duration
	TestResults  []TestResult
}

// NewClient creates a new WAF testing client
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}
	if cfg.Concurrent < 1 {
		cfg.Concurrent = 1
	}

	return &Client{
		parser:    parser.NewParser(),
		executor:  executor.NewHTTPExecutor(cfg.Timeout),
		validator: validator.NewResponseValidator(),
		reporter:  reporter.NewReporter(cfg.Format, cfg.OutputFile),
	}
}

// ValidateFile validates a single YAML test file
func (c *Client) ValidateFile(filename string) error {
	_, err := c.parser.ParseFile(filename)
	return err
}

// ValidateDirectory validates all YAML files in a directory
func (c *Client) ValidateDirectory(dir string) error {
	_, err := c.parser.ParseDirectory(dir)
	return err
}

// RunTestFile executes tests from a single YAML file
func (c *Client) RunTestFile(filename string) (*SuiteResult, error) {
	wafTest, err := c.parser.ParseFile(filename)
	if err != nil {
		return nil, err
	}

	return c.runTests(wafTest)
}

// RunTestDirectory executes all tests from YAML files in a directory
func (c *Client) RunTestDirectory(dir string) (*SuiteResult, error) {
	wafTests, err := c.parser.ParseDirectory(dir)
	if err != nil {
		return nil, err
	}

	// Combine all tests into a single suite
	var allTests []config.Test
	for _, wafTest := range wafTests {
		allTests = append(allTests, wafTest.Spec.Tests...)
	}

	if len(wafTests) == 0 {
		return &SuiteResult{
			SuiteName: "Empty Directory",
		}, nil
	}

	// Use the first test's target configuration
	// In a real implementation, you might want to handle multiple targets differently
	combinedTest := &config.WafTest{
		Metadata: config.Metadata{
			Name: "Combined Tests",
		},
		Spec: config.Spec{
			Target: wafTests[0].Spec.Target,
			Tests:  allTests,
		},
	}

	return c.runTests(combinedTest)
}

// RunTestWithContext executes tests with a context for cancellation
func (c *Client) RunTestWithContext(ctx context.Context, filename string) (*SuiteResult, error) {
	wafTest, err := c.parser.ParseFile(filename)
	if err != nil {
		return nil, err
	}

	return c.runTestsWithContext(ctx, wafTest)
}

// runTests executes the actual test logic
func (c *Client) runTests(wafTest *config.WafTest) (*SuiteResult, error) {
	start := time.Now()
	var testResults []TestResult

	for _, test := range wafTest.Spec.Tests {
		testStart := time.Now()

		response, err := c.executor.ExecuteTest(&test, wafTest.Spec.Target.BaseURL)
		if err != nil {
			testResults = append(testResults, TestResult{
				TestName: test.Name,
				Passed:   false,
				Duration: time.Since(testStart),
				Errors:   []string{err.Error()},
			})
			continue
		}

		validation := c.validator.Validate(response, &test.Expected, test.Name)

		testResults = append(testResults, TestResult{
			TestName: test.Name,
			Passed:   validation.Passed,
			Duration: time.Since(testStart),
			Errors:   validation.Errors,
			Warnings: validation.Warnings,
		})
	}

	// Calculate summary
	passed := 0
	for _, result := range testResults {
		if result.Passed {
			passed++
		}
	}

	return &SuiteResult{
		SuiteName:   wafTest.Metadata.Name,
		TotalTests:  len(testResults),
		PassedTests: passed,
		FailedTests: len(testResults) - passed,
		Duration:    time.Since(start),
		TestResults: testResults,
	}, nil
}

// runTestsWithContext executes tests with context support
func (c *Client) runTestsWithContext(ctx context.Context, wafTest *config.WafTest) (*SuiteResult, error) {
	start := time.Now()
	var testResults []TestResult

	for _, test := range wafTest.Spec.Tests {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		testStart := time.Now()

		response, err := c.executor.ExecuteTestWithContext(ctx, &test, wafTest.Spec.Target.BaseURL)
		if err != nil {
			testResults = append(testResults, TestResult{
				TestName: test.Name,
				Passed:   false,
				Duration: time.Since(testStart),
				Errors:   []string{err.Error()},
			})
			continue
		}

		validation := c.validator.Validate(response, &test.Expected, test.Name)

		testResults = append(testResults, TestResult{
			TestName: test.Name,
			Passed:   validation.Passed,
			Duration: time.Since(testStart),
			Errors:   validation.Errors,
			Warnings: validation.Warnings,
		})
	}

	// Calculate summary
	passed := 0
	for _, result := range testResults {
		if result.Passed {
			passed++
		}
	}

	return &SuiteResult{
		SuiteName:   wafTest.Metadata.Name,
		TotalTests:  len(testResults),
		PassedTests: passed,
		FailedTests: len(testResults) - passed,
		Duration:    time.Since(start),
		TestResults: testResults,
	}, nil
}