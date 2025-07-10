# WafGuard

A professional tool for testing Web Application Firewalls (WAFs) by sending HTTP requests and validating responses against expected outcomes.

## What It Does

WafGuard executes HTTP tests against WAF-protected endpoints using YAML configuration files. It validates responses (status codes, headers, body content) and generates detailed reports showing which tests passed or failed.

**Use Cases:**
- Test WAF rule effectiveness
- Validate security configurations
- Automated security testing in CI/CD
- Penetration testing workflows

## Installation

```bash
# Install globally (recommended)
make install

# Now use from anywhere
wafguard --help
```

## Quick Start

1. **Create a test configuration** (`test.yaml`):

```yaml
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: sql-injection-test
spec:
  target:
    baseUrl: https://your-app.com
    timeout: 30s
  tests:
    - name: basic-sql-injection
      request:
        method: POST
        path: /login
        headers:
          Content-Type: application/json
        body: '{"username": "admin'' OR 1=1--", "password": "test"}'
      expected:
        status: [403, 400]  # WAF should block this
        body:
          contains: ["blocked", "rejected"]
```

2. **Run the test**:

```bash
wafguard run test.yaml
```

3. **View results**:

```
Test: basic-sql-injection
Status: PASS
Duration: 245ms
Request: POST /login
Response Status: 403
---

Suite: sql-injection-test
Total Tests: 1
Passed: 1
Failed: 0
Success Rate: 100.00%
```

## Configuration Format

### Basic Structure

```yaml
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test-name
  description: "Optional description"
spec:
  target:
    baseUrl: https://target.com    # Required
    timeout: 30s                   # Optional, default 30s
  tests:
    - name: test-case-name
      request:
        method: GET|POST|PUT|DELETE
        path: /api/endpoint
        headers:                   # Optional
          Header-Name: value
        body: "request body"       # Optional
      expected:
        status: [200, 201]         # List of acceptable status codes
        headers:                   # Optional header validation
          Content-Type: application/json
        body:                      # Optional body validation
          contains: ["success"]
          not_contains: ["error"]
          exact: "exact match"
          regex: "^pattern.*$"
```

### Request Options

- **method**: HTTP method (GET, POST, PUT, DELETE, etc.)
- **path**: URL path (relative to baseUrl)
- **headers**: Key-value pairs for HTTP headers
- **body**: Request body content (for POST/PUT requests)

### Response Validation

- **status**: Array of acceptable HTTP status codes
- **headers**: Expected response headers (partial matching)
- **body.contains**: Strings that must be present in response body
- **body.not_contains**: Strings that must NOT be present
- **body.exact**: Exact body content match
- **body.regex**: Regular expression pattern match

## Commands

```bash
# Run tests
wafguard run test.yaml                    # Single file
wafguard run tests/                       # Directory
wafguard run tests/ --concurrent 5        # Parallel execution

# Validate configuration
wafguard validate test.yaml               # Check syntax

# Output options
wafguard run test.yaml --format json      # JSON output
wafguard run test.yaml --output results.json  # Save to file
```

## Output Formats

### Text Output (Default)
```
Test: sql-injection-test
Status: FAIL
Duration: 156ms
Request: POST /login
Response Status: 200
Validation Errors:
  - Expected status codes [403, 400], got 200
  - Body should contain 'blocked' but it was not found
```

### JSON Output
```json
{
  "test_name": "sql-injection-test",
  "status": "FAIL", 
  "duration": 156000000,
  "request": {
    "Method": "POST",
    "Path": "/login",
    "Body": "{\"username\": \"admin' OR 1=1--\"}"
  },
  "response": {
    "StatusCode": 200,
    "Body": "{\"message\": \"Login successful\"}"
  },
  "validation_result": {
    "Passed": false,
    "Errors": [
      "Expected status codes [403, 400], got 200",
      "Body should contain 'blocked' but it was not found"
    ]
  }
}
```

## Examples

The `examples/test-configs/` directory contains ready-to-use test cases:

- **SQL Injection**: `sql-injection-test.yaml`
- **Cross-Site Scripting**: `xss-test.yaml` 
- **Directory Traversal**: `directory-traversal-test.yaml`

```bash
# Test all examples
wafguard run examples/test-configs/

# Test specific attack vector
wafguard run examples/test-configs/sql-injection-test.yaml
```

## Development

```bash
# Build
make build

# Run tests
make test

# Install locally
make install-local
```

## License

MIT License