# WAF Tester

A Go-based tool for testing Web Application Firewalls (WAFs) using Kubernetes-like YAML configuration files.

## Features

- 🚀 **Kubernetes-like YAML Configuration**: Define HTTP requests and expected responses in familiar YAML format
- 📊 **Structured Logging**: Comprehensive logging with configurable levels and formats
- 🔄 **Concurrent Testing**: Run multiple tests simultaneously for faster execution
- 📋 **Flexible Response Validation**: Validate status codes, headers, and body content
- 📈 **Multiple Output Formats**: JSON and text output formats with optional file export
- 🎯 **Attack Vector Testing**: Pre-built test cases for common web vulnerabilities

## Installation

```bash
go build -o waf-tester cmd/waf-tester/main.go
```

## Usage

### Run Tests

```bash
# Run a single test file
./waf-tester run examples/test-configs/sql-injection-test.yaml

# Run all tests in a directory
./waf-tester run examples/test-configs/

# Run with custom options
./waf-tester run examples/test-configs/ \
  --concurrent 5 \
  --format json \
  --output results.json \
  --log-level debug
```

### Validate Tests

```bash
# Validate test file syntax
./waf-tester validate examples/test-configs/sql-injection-test.yaml

# Validate all files in directory
./waf-tester validate examples/test-configs/
```

### Command Line Options

- `--concurrent, -c`: Number of concurrent test executions (default: 1)
- `--format, -F`: Output format (json, text) (default: text)
- `--output, -o`: Output file for test results
- `--log-level, -l`: Log level (debug, info, warn, error) (default: info)
- `--log-format, -f`: Log format (json, text) (default: text)

## YAML Configuration

### Basic Structure

```yaml
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: my-test
  description: Test description
spec:
  target:
    baseUrl: https://example.com
    timeout: 30s
  tests:
    - name: test-case-1
      request:
        method: POST
        path: /api/endpoint
        headers:
          Content-Type: application/json
        body: '{"key": "value"}'
      expected:
        status: [200, 201]
        headers:
          Content-Type: application/json
        body:
          contains: ["success"]
          not_contains: ["error"]
```

### Request Configuration

```yaml
request:
  method: GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS
  path: /api/path
  headers:
    Header-Name: header-value
  body: "request body content"
```

### Response Validation

```yaml
expected:
  status: [200, 201, 403]  # List of acceptable status codes
  headers:
    Expected-Header: expected-value
  body:
    contains: ["text1", "text2"]      # Body must contain these strings
    not_contains: ["error", "fail"]   # Body must not contain these strings
    exact: "exact match string"       # Body must exactly match
    regex: "^pattern.*$"              # Body must match regex pattern
```

## Example Test Cases

The `examples/test-configs/` directory contains pre-built test cases for common web vulnerabilities:

- **SQL Injection**: `sql-injection-test.yaml`
- **Cross-Site Scripting (XSS)**: `xss-test.yaml`
- **Directory Traversal**: `directory-traversal-test.yaml`

## Project Structure

```
waf-tester/
├── cmd/waf-tester/           # CLI application
├── pkg/
│   ├── config/               # Configuration types
│   ├── parser/               # YAML parser
│   ├── executor/             # HTTP request executor
│   ├── validator/            # Response validator
│   └── reporter/             # Test result reporting
├── internal/logger/          # Structured logging
├── examples/test-configs/    # Example test cases
└── README.md
```

## Development

### Dependencies

- Go 1.21+
- github.com/spf13/cobra - CLI framework
- github.com/sirupsen/logrus - Structured logging
- gopkg.in/yaml.v3 - YAML parsing
- github.com/go-playground/validator/v10 - Struct validation

### Building

```bash
go mod tidy
go build -o waf-tester cmd/waf-tester/main.go
```

### Testing

```bash
# Validate example configurations
./waf-tester validate examples/test-configs/

# Run example tests (requires internet connection)
./waf-tester run examples/test-configs/sql-injection-test.yaml
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

MIT License - see LICENSE file for details.