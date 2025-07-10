package validator

import (
	"fmt"
	"regexp"
	"strings"
	"waf-tester/internal/logger"
	"waf-tester/pkg/config"
	"waf-tester/pkg/executor"

	"github.com/sirupsen/logrus"
)

type ValidationResult struct {
	Passed   bool
	Errors   []string
	Warnings []string
}

type ResponseValidator struct{}

func NewResponseValidator() *ResponseValidator {
	return &ResponseValidator{}
}

func (v *ResponseValidator) Validate(response *executor.Response, expected *config.Expected, testName string) *ValidationResult {
	result := &ValidationResult{
		Passed:   true,
		Errors:   []string{},
		Warnings: []string{},
	}

	logger.WithFields(logrus.Fields{
		"test_name":   testName,
		"status_code": response.StatusCode,
	}).Debug("Starting response validation")

	v.validateStatusCode(response, expected, result)
	v.validateHeaders(response, expected, result)
	v.validateBody(response, expected, result)

	if len(result.Errors) > 0 {
		result.Passed = false
	}

	logger.WithFields(logrus.Fields{
		"test_name":    testName,
		"passed":       result.Passed,
		"error_count":  len(result.Errors),
		"warning_count": len(result.Warnings),
	}).Info("Response validation completed")

	return result
}

func (v *ResponseValidator) validateStatusCode(response *executor.Response, expected *config.Expected, result *ValidationResult) {
	if len(expected.Status) == 0 {
		result.Warnings = append(result.Warnings, "No expected status codes defined")
		return
	}

	for _, expectedStatus := range expected.Status {
		if response.StatusCode == expectedStatus {
			logger.WithFields(logrus.Fields{
				"expected_status": expectedStatus,
				"actual_status":   response.StatusCode,
			}).Debug("Status code validation passed")
			return
		}
	}

	result.Errors = append(result.Errors, fmt.Sprintf(
		"Status code mismatch: expected one of %v, got %d",
		expected.Status,
		response.StatusCode,
	))
}

func (v *ResponseValidator) validateHeaders(response *executor.Response, expected *config.Expected, result *ValidationResult) {
	if len(expected.Headers) == 0 {
		return
	}

	for expectedKey, expectedValue := range expected.Headers {
		actualValue, exists := response.Headers[expectedKey]
		if !exists {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Missing expected header: %s",
				expectedKey,
			))
			continue
		}

		if !strings.Contains(actualValue, expectedValue) {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Header value mismatch for %s: expected to contain '%s', got '%s'",
				expectedKey,
				expectedValue,
				actualValue,
			))
		}
	}
}

func (v *ResponseValidator) validateBody(response *executor.Response, expected *config.Expected, result *ValidationResult) {
	if expected.Body == nil {
		return
	}

	bodyExpected := expected.Body

	if bodyExpected.Exact != "" {
		if response.Body != bodyExpected.Exact {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Body exact match failed: expected '%s', got '%s'",
				bodyExpected.Exact,
				response.Body,
			))
		}
		return
	}

	if bodyExpected.Regex != "" {
		matched, err := regexp.MatchString(bodyExpected.Regex, response.Body)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Invalid regex pattern '%s': %v",
				bodyExpected.Regex,
				err,
			))
			return
		}
		if !matched {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Body regex match failed: pattern '%s' did not match response body",
				bodyExpected.Regex,
			))
		}
		return
	}

	for _, expectedContains := range bodyExpected.Contains {
		if !strings.Contains(response.Body, expectedContains) {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Body should contain '%s' but it was not found",
				expectedContains,
			))
		}
	}

	for _, expectedNotContains := range bodyExpected.NotContains {
		if strings.Contains(response.Body, expectedNotContains) {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"Body should not contain '%s' but it was found",
				expectedNotContains,
			))
		}
	}
}

func (v *ResponseValidator) ValidateMultiple(responses []*executor.Response, expected []*config.Expected, testNames []string) []*ValidationResult {
	if len(responses) != len(expected) || len(responses) != len(testNames) {
		panic("mismatched lengths in ValidateMultiple")
	}

	results := make([]*ValidationResult, len(responses))
	for i := range responses {
		results[i] = v.Validate(responses[i], expected[i], testNames[i])
	}

	return results
}