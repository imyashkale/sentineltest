# WafGuard

A professional Go-based tool for testing Web Application Firewalls (WAFs) using Kubernetes-like YAML configuration files.

## Features

- 🚀 **Kubernetes-like YAML Configuration**: Define HTTP requests and expected responses in familiar YAML format
- 📊 **Structured Logging**: Comprehensive logging with configurable levels and formats
- 🔄 **Concurrent Testing**: Run multiple tests simultaneously for faster execution
- 📋 **Flexible Response Validation**: Validate status codes, headers, and body content
- 📈 **Multiple Output Formats**: JSON and text output formats with optional file export
- 🎯 **Attack Vector Testing**: Pre-built test cases for common web vulnerabilities

## Installation

### Option 1: Install Globally (Recommended)

Install WafGuard globally so you can use `wafguard` from anywhere:

```bash
# Clone the repository
git clone <repository-url>
cd wafguard

# Install globally (to /usr/local/bin)
make install

# Now you can use wafguard from anywhere!
wafguard --help
```

### Option 2: Local Go Installation

Install to your Go bin directory (no sudo required):

```bash
# Install locally to $GOPATH/bin or ~/go/bin
make install-local

# Make sure your Go bin is in PATH
export PATH=$PATH:$GOPATH/bin
# or
export PATH=$PATH:$HOME/go/bin

# Now you can use wafguard globally
wafguard --help
```

### Option 3: Build Only

Build the binary without installing:

```bash
# Build binary to ./bin/wafguard
make build

# Use the local binary
./bin/wafguard --help
```

### Uninstall

To remove WafGuard from your system:

```bash
make uninstall
```

## Usage

### Run Tests

```bash
# Run a single test file
wafguard run examples/test-configs/sql-injection-test.yaml

# Run all tests in a directory
wafguard run examples/test-configs/

# Run with custom options
wafguard run examples/test-configs/ \
  --concurrent 5 \
  --format json \
  --output results.json \
  --log-level debug
```

### Validate Tests

```bash
# Validate test file syntax
wafguard validate examples/test-configs/sql-injection-test.yaml

# Validate all files in directory
wafguard validate examples/test-configs/
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
wafguard/
├── cmd/wafguard/             # CLI application
├── internal/                 # Internal packages (core logic)
│   ├── core/config/          # Configuration types
│   ├── parser/               # YAML parser
│   ├── executor/             # HTTP request executor
│   ├── validator/            # Response validator
│   ├── reporter/             # Test result reporting
│   └── logger/               # Structured logging
├── pkg/                      # Public API
│   ├── client/               # Public client interface
│   └── types/                # Public type definitions
├── examples/test-configs/    # Example test cases
├── Makefile                  # Build automation
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
# Using Make (recommended)
make build

# Or manually
go mod tidy
go build -o wafguard cmd/wafguard/main.go

# For development
make dev-deps  # Install development dependencies
make test      # Run tests
make lint      # Run linter
make format    # Format code
```

### Testing

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Validate example configurations
wafguard validate examples/test-configs/

# Run example tests (requires internet connection)
wafguard run examples/test-configs/sql-injection-test.yaml
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

MIT License - see LICENSE file for details.