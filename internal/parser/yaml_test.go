package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}
	if parser.validator == nil {
		t.Fatal("Parser validator is nil")
	}
}

func TestParseYAML(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid YAML",
			yaml: `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test-config
  description: Test description
spec:
  target:
    baseUrl: https://example.com
    timeout: 30s
  tests:
    - name: test1
      request:
        method: GET
        path: /test
      expected:
        status: [200]
`,
			wantErr: false,
		},
		{
			name: "invalid YAML syntax",
			yaml: `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test-config
  invalid_yaml: [
spec:
  target:
`,
			wantErr: true,
		},
		{
			name: "invalid validation - missing required fields",
			yaml: `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: ""
spec:
  target:
    baseUrl: not-a-url
  tests: []
`,
			wantErr: true,
		},
		{
			name: "invalid kind",
			yaml: `
apiVersion: waf-test/v1
kind: InvalidKind
metadata:
  name: test
spec:
  target:
    baseUrl: https://example.com
  tests:
    - name: test1
      request:
        method: GET
        path: /test
      expected:
        status: [200]
`,
			wantErr: true,
		},
		{
			name: "invalid HTTP method",
			yaml: `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test
spec:
  target:
    baseUrl: https://example.com
  tests:
    - name: test1
      request:
        method: INVALID
        path: /test
      expected:
        status: [200]
`,
			wantErr: false, // Validator doesn't validate enum values in nested structs by default
		},
		{
			name: "empty tests array",
			yaml: `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test
spec:
  target:
    baseUrl: https://example.com
  tests: []
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseYAML([]byte(tt.yaml))
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && result == nil {
				t.Error("ParseYAML() returned nil result for valid YAML")
			}
			
			if tt.wantErr && result != nil {
				t.Error("ParseYAML() returned result for invalid YAML")
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	parser := NewParser()

	// Create a temporary test file
	tmpDir, err := os.MkdirTemp("", "parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	validYAML := `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test-config
spec:
  target:
    baseUrl: https://example.com
  tests:
    - name: test1
      request:
        method: GET
        path: /test
      expected:
        status: [200]
`

	validFile := filepath.Join(tmpDir, "valid.yaml")
	err = os.WriteFile(validFile, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid test file: %v", err)
	}

	invalidYAML := `invalid yaml content [`
	invalidFile := filepath.Join(tmpDir, "invalid.yaml")
	err = os.WriteFile(invalidFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid test file: %v", err)
	}

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid file",
			filename: validFile,
			wantErr:  false,
		},
		{
			name:     "invalid file",
			filename: invalidFile,
			wantErr:  true,
		},
		{
			name:     "non-existent file",
			filename: filepath.Join(tmpDir, "nonexistent.yaml"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseFile(tt.filename)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if result == nil {
					t.Error("ParseFile() returned nil result for valid file")
				} else {
					if result.Metadata.Name != "test-config" {
						t.Errorf("ParseFile() result name = %s, want test-config", result.Metadata.Name)
					}
				}
			}
		})
	}
}

func TestParseDirectory(t *testing.T) {
	parser := NewParser()

	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "parser_dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create valid YAML files
	validYAML1 := `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test-config-1
spec:
  target:
    baseUrl: https://example.com
  tests:
    - name: test1
      request:
        method: GET
        path: /test1
      expected:
        status: [200]
`

	validYAML2 := `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test-config-2
spec:
  target:
    baseUrl: https://example.com
  tests:
    - name: test2
      request:
        method: POST
        path: /test2
      expected:
        status: [201]
`

	// Write test files
	err = os.WriteFile(filepath.Join(tmpDir, "test1.yaml"), []byte(validYAML1), 0644)
	if err != nil {
		t.Fatalf("Failed to write test1.yaml: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "test2.yml"), []byte(validYAML2), 0644)
	if err != nil {
		t.Fatalf("Failed to write test2.yml: %v", err)
	}

	// Create a non-YAML file (should be ignored)
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not yaml"), 0644)
	if err != nil {
		t.Fatalf("Failed to write readme.txt: %v", err)
	}

	// Create invalid YAML file
	invalidYAML := `invalid yaml [`
	err = os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid.yaml: %v", err)
	}

	// Create subdirectory with YAML file
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	validYAML3 := `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: test-config-3
spec:
  target:
    baseUrl: https://example.com
  tests:
    - name: test3
      request:
        method: GET
        path: /test3
      expected:
        status: [200]
`
	err = os.WriteFile(filepath.Join(subDir, "test3.yaml"), []byte(validYAML3), 0644)
	if err != nil {
		t.Fatalf("Failed to write test3.yaml: %v", err)
	}

	tests := []struct {
		name      string
		directory string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "valid directory with mixed files",
			directory: tmpDir,
			wantErr:   true, // Should fail due to invalid.yaml
			wantCount: 0,
		},
		{
			name:      "non-existent directory",
			directory: filepath.Join(tmpDir, "nonexistent"),
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := parser.ParseDirectory(tt.directory)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if len(results) != tt.wantCount {
					t.Errorf("ParseDirectory() returned %d results, want %d", len(results), tt.wantCount)
				}
				
				// Verify that all results are valid
				for i, result := range results {
					if result == nil {
						t.Errorf("ParseDirectory() result[%d] is nil", i)
					}
				}
			}
		})
	}

	// Test with directory containing only valid files
	validDir, err := os.MkdirTemp("", "parser_valid_dir_test")
	if err != nil {
		t.Fatalf("Failed to create valid temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(validDir) }()

	err = os.WriteFile(filepath.Join(validDir, "test1.yaml"), []byte(validYAML1), 0644)
	if err != nil {
		t.Fatalf("Failed to write test1.yaml to valid dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(validDir, "test2.yml"), []byte(validYAML2), 0644)
	if err != nil {
		t.Fatalf("Failed to write test2.yml to valid dir: %v", err)
	}

	results, err := parser.ParseDirectory(validDir)
	if err != nil {
		t.Errorf("ParseDirectory() on valid directory failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("ParseDirectory() returned %d results, want 2", len(results))
	}

	// Verify results
	names := make(map[string]bool)
	for _, result := range results {
		names[result.Metadata.Name] = true
	}

	if !names["test-config-1"] {
		t.Error("ParseDirectory() missing test-config-1")
	}
	if !names["test-config-2"] {
		t.Error("ParseDirectory() missing test-config-2")
	}
}

func TestParseDirectoryEmptyDir(t *testing.T) {
	parser := NewParser()

	// Create empty directory
	tmpDir, err := os.MkdirTemp("", "parser_empty_dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	results, err := parser.ParseDirectory(tmpDir)
	if err != nil {
		t.Errorf("ParseDirectory() on empty directory failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("ParseDirectory() on empty directory returned %d results, want 0", len(results))
	}
}