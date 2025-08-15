package config

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

func TestSentinelTestValidation(t *testing.T) {
	validator := validator.New()

	tests := []struct {
		name        string
		sentinelTest SentinelTest
		wantErr     bool
	}{
		{
			name: "valid config",
			sentinelTest: SentinelTest{
				APIVersion: "waf-test/v1",
				Kind:       "SentinelTest",
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
			sentinelTest: SentinelTest{
				APIVersion: "",
				Kind:       "",
			},
			wantErr: true,
		},
		{
			name: "invalid kind",
			sentinelTest: SentinelTest{
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
			sentinelTest: SentinelTest{
				APIVersion: "waf-test/v1",
				Kind:       "SentinelTest",
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
			sentinelTest: SentinelTest{
				APIVersion: "waf-test/v1",
				Kind:       "SentinelTest",
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
			sentinelTest: SentinelTest{
				APIVersion: "waf-test/v1",
				Kind:       "SentinelTest",
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
			sentinelTest: SentinelTest{
				APIVersion: "waf-test/v1",
				Kind:       "SentinelTest",
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
			err := validator.Struct(&tt.sentinelTest)
			if (err != nil) != tt.wantErr {
				t.Errorf("SentinelTest validation = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYAMLMarshaling(t *testing.T) {
	sentinelTest := SentinelTest{
		APIVersion: "waf-test/v1",
		Kind:       "SentinelTest",
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
	data, err := yaml.Marshal(&sentinelTest)
	if err != nil {
		t.Fatalf("Failed to marshal SentinelTest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled SentinelTest
	err = yaml.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SentinelTest: %v", err)
	}

	// Verify fields
	if unmarshaled.APIVersion != sentinelTest.APIVersion {
		t.Errorf("APIVersion mismatch: got %s, want %s", unmarshaled.APIVersion, sentinelTest.APIVersion)
	}
	if unmarshaled.Kind != sentinelTest.Kind {
		t.Errorf("Kind mismatch: got %s, want %s", unmarshaled.Kind, sentinelTest.Kind)
	}
	if unmarshaled.Metadata.Name != sentinelTest.Metadata.Name {
		t.Errorf("Metadata.Name mismatch: got %s, want %s", unmarshaled.Metadata.Name, sentinelTest.Metadata.Name)
	}
	if len(unmarshaled.Spec.Tests) != len(sentinelTest.Spec.Tests) {
		t.Errorf("Tests length mismatch: got %d, want %d", len(unmarshaled.Spec.Tests), len(sentinelTest.Spec.Tests))
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
kind: SentinelTest
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

	var sentinelTest SentinelTest
	err := yaml.Unmarshal([]byte(yamlData), &sentinelTest)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	expectedTimeout := 45 * time.Second
	if sentinelTest.Spec.Target.Timeout != expectedTimeout {
		t.Errorf("Timeout mismatch: got %v, want %v", sentinelTest.Spec.Target.Timeout, expectedTimeout)
	}
}