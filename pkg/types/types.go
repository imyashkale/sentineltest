// Package types provides public type definitions for the WAF testing framework.
package types

import "time"

// WafTestConfig represents a complete WAF test configuration
type WafTestConfig struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}

// Metadata contains test metadata
type Metadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Spec contains the test specification
type Spec struct {
	Target Target `yaml:"target" json:"target"`
	Tests  []Test `yaml:"tests" json:"tests"`
}

// Target defines the target endpoint configuration
type Target struct {
	BaseURL string        `yaml:"baseUrl" json:"baseUrl"`
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Test defines a single test case
type Test struct {
	Name     string   `yaml:"name" json:"name"`
	Request  Request  `yaml:"request" json:"request"`
	Expected Expected `yaml:"expected" json:"expected"`
}

// Request defines the HTTP request configuration
type Request struct {
	Method  string            `yaml:"method" json:"method"`
	Path    string            `yaml:"path" json:"path"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty" json:"body,omitempty"`
}

// Expected defines the expected response validation criteria
type Expected struct {
	Status  []int             `yaml:"status" json:"status"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body    *BodyExpected     `yaml:"body,omitempty" json:"body,omitempty"`
}

// BodyExpected defines body validation criteria
type BodyExpected struct {
	Contains    []string `yaml:"contains,omitempty" json:"contains,omitempty"`
	NotContains []string `yaml:"not_contains,omitempty" json:"not_contains,omitempty"`
	Exact       string   `yaml:"exact,omitempty" json:"exact,omitempty"`
	Regex       string   `yaml:"regex,omitempty" json:"regex,omitempty"`
}

// TestResult represents the result of a single test execution
type TestResult struct {
	TestName string        `json:"test_name"`
	Status   string        `json:"status"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration"`
	Errors   []string      `json:"errors,omitempty"`
	Warnings []string      `json:"warnings,omitempty"`
}

// SuiteResult represents the result of a test suite execution
type SuiteResult struct {
	SuiteName   string        `json:"suite_name"`
	TotalTests  int           `json:"total_tests"`
	PassedTests int           `json:"passed_tests"`
	FailedTests int           `json:"failed_tests"`
	Duration    time.Duration `json:"duration"`
	Tests       []TestResult  `json:"tests"`
	Timestamp   time.Time     `json:"timestamp"`
}

// ClientConfig represents configuration for the WAF testing client
type ClientConfig struct {
	Timeout    time.Duration `json:"timeout,omitempty"`
	OutputFile string        `json:"output_file,omitempty"`
	Format     string        `json:"format,omitempty"` // "json" or "text"
	Concurrent int           `json:"concurrent,omitempty"`
	LogLevel   string        `json:"log_level,omitempty"`
	LogFormat  string        `json:"log_format,omitempty"`
}