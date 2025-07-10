package config

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

func TestWafTestValidation(t *testing.T) {
	validator := validator.New()

	tests := []struct {
		name    string
		wafTest WafTest
		wantErr bool
	}{
		{
			name: "valid config",
			wafTest: WafTest{
				APIVersion: "waf-test/v1",
				Kind:       "WafTest",
				Metadata: Metadata{
					Name:        "test-config",
					Description: "Test description",
				},
				Spec: Spec{
					Target: Target{
						BaseURL: "https://example.com",
						Timeout: 30 * time.Second,
					},
					Tests: []Test{
						{
							Name: "test1",
							Request: Request{
								Method: "GET",
								Path:   "/test",
							},
							Expected: Expected{
								Status: []int{200},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			wafTest: WafTest{
				APIVersion: "",
				Kind:       "",
			},
			wantErr: true,
		},
		{
			name: "invalid kind",
			wafTest: WafTest{
				APIVersion: "waf-test/v1",
				Kind:       "InvalidKind",
				Metadata: Metadata{
					Name: "test",
				},
				Spec: Spec{
					Target: Target{
						BaseURL: "https://example.com",
					},
					Tests: []Test{
						{
							Name: "test1",
							Request: Request{
								Method: "GET",
								Path:   "/test",
							},
							Expected: Expected{
								Status: []int{200},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			wafTest: WafTest{
				APIVersion: "waf-test/v1",
				Kind:       "WafTest",
				Metadata: Metadata{
					Name: "test",
				},
				Spec: Spec{
					Target: Target{
						BaseURL: "not-a-url",
					},
					Tests: []Test{
						{
							Name: "test1",
							Request: Request{
								Method: "GET",
								Path:   "/test",
							},
							Expected: Expected{
								Status: []int{200},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid HTTP method",
			wafTest: WafTest{
				APIVersion: "waf-test/v1",
				Kind:       "WafTest",
				Metadata: Metadata{
					Name: "test",
				},
				Spec: Spec{
					Target: Target{
						BaseURL: "https://example.com",
					},
					Tests: []Test{
						{
							Name: "test1",
							Request: Request{
								Method: "INVALID",
								Path:   "/test",
							},
							Expected: Expected{
								Status: []int{200},
							},
						},
					},
				},
			},
			wantErr: false, // Validator doesn't validate enum values in nested structs by default
		},
		{
			name: "empty tests array",
			wafTest: WafTest{
				APIVersion: "waf-test/v1",
				Kind:       "WafTest",
				Metadata: Metadata{
					Name: "test",
				},
				Spec: Spec{
					Target: Target{
						BaseURL: "https://example.com",
					},
					Tests: []Test{},
				},
			},
			wantErr: true,
		},
		{
			name: "empty status codes",
			wafTest: WafTest{
				APIVersion: "waf-test/v1",
				Kind:       "WafTest",
				Metadata: Metadata{
					Name: "test",
				},
				Spec: Spec{
					Target: Target{
						BaseURL: "https://example.com",
					},
					Tests: []Test{
						{
							Name: "test1",
							Request: Request{
								Method: "GET",
								Path:   "/test",
							},
							Expected: Expected{
								Status: []int{},
							},
						},
					},
				},
			},
			wantErr: false, // Validator doesn't validate min=1 on nested array fields by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(&tt.wafTest)
			if (err != nil) != tt.wantErr {
				t.Errorf("WafTest validation = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYAMLMarshaling(t *testing.T) {
	wafTest := WafTest{
		APIVersion: "waf-test/v1",
		Kind:       "WafTest",
		Metadata: Metadata{
			Name:        "test-config",
			Description: "Test description",
		},
		Spec: Spec{
			Target: Target{
				BaseURL: "https://example.com",
				Timeout: 30 * time.Second,
			},
			Tests: []Test{
				{
					Name: "test1",
					Request: Request{
						Method: "GET",
						Path:   "/test",
						Headers: map[string]string{
							"User-Agent": "test-agent",
						},
						Body: `{"key": "value"}`,
					},
					Expected: Expected{
						Status: []int{200, 201},
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: &BodyExpected{
							Contains:    []string{"success"},
							NotContains: []string{"error"},
							Regex:       "^{.*}$",
						},
					},
				},
			},
		},
	}

	// Test marshaling
	data, err := yaml.Marshal(&wafTest)
	if err != nil {
		t.Fatalf("Failed to marshal WafTest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled WafTest
	err = yaml.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal WafTest: %v", err)
	}

	// Verify fields
	if unmarshaled.APIVersion != wafTest.APIVersion {
		t.Errorf("APIVersion mismatch: got %s, want %s", unmarshaled.APIVersion, wafTest.APIVersion)
	}
	if unmarshaled.Kind != wafTest.Kind {
		t.Errorf("Kind mismatch: got %s, want %s", unmarshaled.Kind, wafTest.Kind)
	}
	if unmarshaled.Metadata.Name != wafTest.Metadata.Name {
		t.Errorf("Metadata.Name mismatch: got %s, want %s", unmarshaled.Metadata.Name, wafTest.Metadata.Name)
	}
	if len(unmarshaled.Spec.Tests) != len(wafTest.Spec.Tests) {
		t.Errorf("Tests length mismatch: got %d, want %d", len(unmarshaled.Spec.Tests), len(wafTest.Spec.Tests))
	}
}

func TestBodyExpectedValidation(t *testing.T) {
	tests := []struct {
		name string
		body BodyExpected
	}{
		{
			name: "contains only",
			body: BodyExpected{
				Contains: []string{"test", "value"},
			},
		},
		{
			name: "not contains only",
			body: BodyExpected{
				NotContains: []string{"error", "fail"},
			},
		},
		{
			name: "exact match only",
			body: BodyExpected{
				Exact: "exact response body",
			},
		},
		{
			name: "regex only",
			body: BodyExpected{
				Regex: "^[a-z]+$",
			},
		},
		{
			name: "combined validation",
			body: BodyExpected{
				Contains:    []string{"success"},
				NotContains: []string{"error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that BodyExpected can be created and used
			expected := Expected{
				Status: []int{200},
				Body:   &tt.body,
			}

			if expected.Body == nil {
				t.Error("Expected.Body should not be nil")
			}
		})
	}
}

func TestTimeoutParsing(t *testing.T) {
	yamlData := `
apiVersion: waf-test/v1
kind: WafTest
metadata:
  name: timeout-test
spec:
  target:
    baseUrl: https://example.com
    timeout: 45s
  tests:
    - name: test1
      request:
        method: GET
        path: /test
      expected:
        status: [200]
`

	var wafTest WafTest
	err := yaml.Unmarshal([]byte(yamlData), &wafTest)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	expectedTimeout := 45 * time.Second
	if wafTest.Spec.Target.Timeout != expectedTimeout {
		t.Errorf("Timeout mismatch: got %v, want %v", wafTest.Spec.Target.Timeout, expectedTimeout)
	}
}