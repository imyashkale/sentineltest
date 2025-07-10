package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected logrus.Level
	}{
		{
			name:     "debug level",
			level:    "debug",
			expected: logrus.DebugLevel,
		},
		{
			name:     "info level",
			level:    "info",
			expected: logrus.InfoLevel,
		},
		{
			name:     "warn level",
			level:    "warn",
			expected: logrus.WarnLevel,
		},
		{
			name:     "error level",
			level:    "error",
			expected: logrus.ErrorLevel,
		},
		{
			name:     "invalid level defaults to info",
			level:    "invalid",
			expected: logrus.InfoLevel,
		},
		{
			name:     "empty level defaults to info",
			level:    "",
			expected: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.level)
			if log.Level != tt.expected {
				t.Errorf("SetLevel(%s) level = %v, want %v", tt.level, log.Level, tt.expected)
			}
		})
	}
}

func TestSetFormatter(t *testing.T) {
	tests := []struct {
		name            string
		format          string
		expectedType    string
		expectedDefault bool
	}{
		{
			name:         "json formatter",
			format:       "json",
			expectedType: "*logrus.JSONFormatter",
		},
		{
			name:         "text formatter",
			format:       "text",
			expectedType: "*logrus.TextFormatter",
		},
		{
			name:            "invalid format defaults to json",
			format:          "invalid",
			expectedType:    "*logrus.JSONFormatter",
			expectedDefault: true,
		},
		{
			name:            "empty format defaults to json",
			format:          "",
			expectedType:    "*logrus.JSONFormatter",
			expectedDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetFormatter(tt.format)
			
			formatterType := getFormatterType(log.Formatter)
			if formatterType != tt.expectedType {
				t.Errorf("SetFormatter(%s) formatter type = %s, want %s", tt.format, formatterType, tt.expectedType)
			}

			// For text formatter, verify FullTimestamp is set
			if tt.format == "text" {
				if textFormatter, ok := log.Formatter.(*logrus.TextFormatter); ok {
					if !textFormatter.FullTimestamp {
						t.Error("SetFormatter(text) should set FullTimestamp to true")
					}
				}
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}
	if logger != log {
		t.Error("GetLogger() should return the same logger instance")
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Test that the logger is properly initialized
	if log == nil {
		t.Fatal("Logger not initialized")
	}

	// Test default level is Info
	if log.Level != logrus.InfoLevel {
		t.Errorf("Default log level = %v, want %v", log.Level, logrus.InfoLevel)
	}

	// Test default formatter is JSON
	formatterType := getFormatterType(log.Formatter)
	if formatterType != "*logrus.JSONFormatter" {
		t.Errorf("Default formatter type = %s, want *logrus.JSONFormatter", formatterType)
	}
}

func TestLogFunctions(t *testing.T) {
	// Capture log output for testing
	var buf bytes.Buffer
	originalOutput := log.Out
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(originalOutput)
	}()

	// Set to debug level to capture all logs
	SetLevel("debug")
	SetFormatter("json")

	tests := []struct {
		name     string
		logFunc  func(...interface{})
		message  string
		level    string
	}{
		{
			name:     "debug log",
			logFunc:  Debug,
			message:  "debug message",
			level:    "debug",
		},
		{
			name:     "info log",
			logFunc:  Info,
			message:  "info message",
			level:    "info",
		},
		{
			name:     "warn log",
			logFunc:  Warn,
			message:  "warn message",
			level:    "warning",
		},
		{
			name:     "error log",
			logFunc:  Error,
			message:  "error message",
			level:    "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			
			tt.logFunc(tt.message)
			
			output := buf.String()
			if output == "" {
				t.Errorf("%s produced no output", tt.name)
				return
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			err := json.Unmarshal([]byte(output), &logEntry)
			if err != nil {
				t.Errorf("Failed to parse log output as JSON: %v", err)
				return
			}

			// Check level
			if logEntry["level"] != tt.level {
				t.Errorf("%s level = %v, want %s", tt.name, logEntry["level"], tt.level)
			}

			// Check message
			if logEntry["msg"] != tt.message {
				t.Errorf("%s message = %v, want %s", tt.name, logEntry["msg"], tt.message)
			}

			// Check timestamp exists
			if _, exists := logEntry["time"]; !exists {
				t.Errorf("%s missing timestamp", tt.name)
			}
		})
	}
}

func TestWithFields(t *testing.T) {
	// Capture log output for testing
	var buf bytes.Buffer
	originalOutput := log.Out
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(originalOutput)
	}()

	SetLevel("info")
	SetFormatter("json")

	fields := logrus.Fields{
		"user_id":    123,
		"operation":  "test",
		"component":  "logger_test",
	}

	entry := WithFields(fields)
	if entry == nil {
		t.Fatal("WithFields() returned nil")
	}

	entry.Info("test message with fields")

	output := buf.String()
	if output == "" {
		t.Fatal("WithFields() produced no output")
	}

	// Parse JSON log entry
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}

	// Check that fields are present
	for key, expectedValue := range fields {
		actualValue := logEntry[key]
		// Handle type conversion for numbers (JSON unmarshaling converts numbers to float64)
		if key == "user_id" {
			if actualFloat, ok := actualValue.(float64); ok {
				if int(actualFloat) != expectedValue.(int) {
					t.Errorf("WithFields() field %s = %v, want %v", key, actualValue, expectedValue)
				}
			} else {
				t.Errorf("WithFields() field %s = %v (type %T), want %v (type %T)", key, actualValue, actualValue, expectedValue, expectedValue)
			}
		} else {
			if actualValue != expectedValue {
				t.Errorf("WithFields() field %s = %v, want %v", key, actualValue, expectedValue)
			}
		}
	}

	// Check message
	if logEntry["msg"] != "test message with fields" {
		t.Errorf("WithFields() message = %v, want 'test message with fields'", logEntry["msg"])
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// Capture log output for testing
	var buf bytes.Buffer
	originalOutput := log.Out
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(originalOutput)
	}()

	// Set to error level - should filter out debug, info, warn
	SetLevel("error")
	SetFormatter("json")

	// These should not produce output
	buf.Reset()
	Debug("debug message")
	if buf.String() != "" {
		t.Error("Debug message should be filtered at error level")
	}

	buf.Reset()
	Info("info message")
	if buf.String() != "" {
		t.Error("Info message should be filtered at error level")
	}

	buf.Reset()
	Warn("warn message")
	if buf.String() != "" {
		t.Error("Warn message should be filtered at error level")
	}

	// This should produce output
	buf.Reset()
	Error("error message")
	if buf.String() == "" {
		t.Error("Error message should not be filtered at error level")
	}
}

func TestTextFormatterOutput(t *testing.T) {
	// Capture log output for testing
	var buf bytes.Buffer
	originalOutput := log.Out
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(originalOutput)
	}()

	SetLevel("info")
	SetFormatter("text")

	Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Text formatter produced no output")
	}

	// Text format should not be JSON
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	if err == nil {
		t.Error("Text formatter output should not be valid JSON")
	}

	// Should contain the message
	if !strings.Contains(output, "test message") {
		t.Error("Text formatter output should contain the log message")
	}

	// Should contain level info (logrus text format uses "info" not "INFO")
	if !strings.Contains(strings.ToLower(output), "info") {
		t.Errorf("Text formatter output should contain the log level. Output: %s", output)
	}
}

func TestFatalFunction(t *testing.T) {
	// Note: We can't easily test Fatal() because it calls os.Exit()
	// In a real testing scenario, you might use dependency injection
	// or test the underlying logrus functionality directly
	
	// For now, we'll just verify the function exists
	// We can't test if Fatal == nil because function comparison is not allowed
	// Instead, we just ensure it's accessible
	defer func() {
		if r := recover(); r != nil {
			t.Error("Fatal function should be accessible")
		}
	}()
	
	// Just verify we can reference the function without calling it
	_ = Fatal
}

// Helper function to get formatter type as string
func getFormatterType(formatter logrus.Formatter) string {
	switch formatter.(type) {
	case *logrus.JSONFormatter:
		return "*logrus.JSONFormatter"
	case *logrus.TextFormatter:
		return "*logrus.TextFormatter"
	default:
		return "unknown"
	}
}

func TestLoggerConcurrency(t *testing.T) {
	// Test that the logger can handle concurrent access
	// This is more of a smoke test since logrus should handle concurrency
	
	var buf bytes.Buffer
	originalOutput := log.Out
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(originalOutput)
	}()

	SetLevel("info")
	SetFormatter("json")

	// Launch multiple goroutines to log concurrently
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				WithFields(logrus.Fields{
					"goroutine": id,
					"iteration": j,
				}).Info("concurrent log message")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	output := buf.String()
	if output == "" {
		t.Fatal("Concurrent logging produced no output")
	}

	// Count number of log lines (each should be a separate JSON object)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	expectedLines := 100 // 10 goroutines * 10 iterations
	
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		if line == "" {
			continue
		}
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}