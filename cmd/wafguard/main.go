package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
	"wafguard/internal/logger"
	"wafguard/internal/core/config"
	"wafguard/internal/executor"
	"wafguard/internal/parser"
	"wafguard/internal/reporter"
	"wafguard/internal/validator"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logLevel   string
	logFormat  string
	outputFile string
	format     string
	concurrent int
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "wafguard",
		Short: "WafGuard - WAF Testing Tool",
		Long:  "WafGuard is a professional tool for testing Web Application Firewalls using Kubernetes-like YAML configuration files",
	}

	var runCmd = &cobra.Command{
		Use:   "run [file or directory]",
		Short: "Run WAF tests",
		Long:  "Run WAF tests from a YAML file or directory containing YAML files",
		Args:  cobra.ExactArgs(1),
		RunE:  runTests,
	}

	var validateCmd = &cobra.Command{
		Use:   "validate [file or directory]",
		Short: "Validate YAML test files",
		Long:  "Validate YAML test files without executing them",
		Args:  cobra.ExactArgs(1),
		RunE:  validateTests,
	}

	runCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	runCmd.Flags().StringVarP(&logFormat, "log-format", "f", "text", "Log format (json, text)")
	runCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for test results")
	runCmd.Flags().StringVarP(&format, "format", "F", "text", "Output format (json, text)")
	runCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 1, "Number of concurrent test executions")

	validateCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	validateCmd.Flags().StringVarP(&logFormat, "log-format", "f", "text", "Log format (json, text)")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTests(cmd *cobra.Command, args []string) error {
	setupLogger()
	
	path := args[0]
	logger.WithFields(logrus.Fields{
		"path":       path,
		"concurrent": concurrent,
		"format":     format,
	}).Info("Starting WAF tests")

	p := parser.NewParser()
	var tests []*config.SentinelTest
	var err error

	if isDirectory(path) {
		tests, err = p.ParseDirectory(path)
	} else {
		test, parseErr := p.ParseFile(path)
		if parseErr != nil {
			return parseErr
		}
		tests = []*config.SentinelTest{test}
	}

	if err != nil {
		return fmt.Errorf("failed to parse tests: %w", err)
	}

	if len(tests) == 0 {
		logger.Warn("No tests found")
		return nil
	}

	return executeTests(tests)
}

func validateTests(cmd *cobra.Command, args []string) error {
	setupLogger()
	
	path := args[0]
	logger.WithFields(logrus.Fields{
		"path": path,
	}).Info("Validating WAF test files")

	p := parser.NewParser()
	var tests []*config.SentinelTest
	var err error

	if isDirectory(path) {
		tests, err = p.ParseDirectory(path)
	} else {
		test, parseErr := p.ParseFile(path)
		if parseErr != nil {
			return parseErr
		}
		tests = []*config.SentinelTest{test}
	}

	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"test_count": len(tests),
	}).Info("All test files are valid")

	return nil
}

func executeTests(tests []*config.SentinelTest) error {
	rep := reporter.NewReporter(format, outputFile)
	start := time.Now()
	
	var allReports []reporter.TestReport
	var mu sync.Mutex

	for _, test := range tests {
		logger.WithFields(logrus.Fields{
			"test_suite": test.Metadata.Name,
			"test_count": len(test.Spec.Tests),
		}).Info("Executing test suite")

		httpExecutor := executor.NewHTTPExecutor(test.Spec.Target.Timeout)
		responseValidator := validator.NewResponseValidator()

		if concurrent <= 1 {
			reports := executeTestsSequentially(test, httpExecutor, responseValidator, rep)
			mu.Lock()
			allReports = append(allReports, reports...)
			mu.Unlock()
		} else {
			reports := executeTestsConcurrently(test, httpExecutor, responseValidator, rep, concurrent)
			mu.Lock()
			allReports = append(allReports, reports...)
			mu.Unlock()
		}
	}

	duration := time.Since(start)
	suiteReport := rep.GenerateSuiteReport("All Tests", allReports, duration)
	
	rep.PrintSuiteReport(suiteReport)
	
	if err := rep.SaveSuiteReport(suiteReport); err != nil {
		logger.Error("Failed to save report:", err)
	}

	if suiteReport.FailedTests > 0 {
		os.Exit(1)
	}

	return nil
}

func executeTestsSequentially(sentinelTest *config.SentinelTest, httpExecutor *executor.HTTPExecutor, responseValidator *validator.ResponseValidator, rep *reporter.Reporter) []reporter.TestReport {
	var reports []reporter.TestReport

	for _, test := range sentinelTest.Spec.Tests {
		start := time.Now()
		
		response, err := httpExecutor.ExecuteTest(&test, sentinelTest.Spec.Target.BaseURL)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"test_name": test.Name,
				"error":     err,
			}).Error("Failed to execute test")
			continue
		}

		validation := responseValidator.Validate(response, &test.Expected, test.Name)
		duration := time.Since(start)

		report := rep.GenerateTestReport(test.Name, &test.Request, response, validation, duration)
		reports = append(reports, *report)
		
		rep.PrintTestReport(report)
	}

	return reports
}

func executeTestsConcurrently(sentinelTest *config.SentinelTest, httpExecutor *executor.HTTPExecutor, responseValidator *validator.ResponseValidator, rep *reporter.Reporter, maxConcurrent int) []reporter.TestReport {
	var reports []reporter.TestReport
	var mu sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, maxConcurrent)
	ctx := context.Background()

	for _, test := range sentinelTest.Spec.Tests {
		wg.Add(1)
		go func(t config.Test) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			start := time.Now()
			
			response, err := httpExecutor.ExecuteTestWithContext(ctx, &t, sentinelTest.Spec.Target.BaseURL)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"test_name": t.Name,
					"error":     err,
				}).Error("Failed to execute test")
				return
			}

			validation := responseValidator.Validate(response, &t.Expected, t.Name)
			duration := time.Since(start)

			report := rep.GenerateTestReport(t.Name, &t.Request, response, validation, duration)
			
			mu.Lock()
			reports = append(reports, *report)
			mu.Unlock()
			
			rep.PrintTestReport(report)
		}(test)
	}

	wg.Wait()
	return reports
}

func setupLogger() {
	logger.SetLevel(logLevel)
	logger.SetFormatter(logFormat)
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}