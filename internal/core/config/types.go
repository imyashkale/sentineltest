package config

import (
	"time"
)

type SentinelTest struct {
	APIVersion string   `yaml:"apiVersion" validate:"required"`
	Kind       string   `yaml:"kind" validate:"required,eq=SentinelTest"`
	Metadata   Metadata `yaml:"metadata" validate:"required"`
	Spec       Spec     `yaml:"spec" validate:"required"`
}

type Metadata struct {
	Name        string `yaml:"name" validate:"required"`
	Description string `yaml:"description,omitempty"`
}

type Spec struct {
	Target Target `yaml:"target" validate:"required"`
	Tests  []Test `yaml:"tests" validate:"required,min=1"`
}

type Target struct {
	BaseURL string        `yaml:"baseUrl" validate:"required,url"`
	Timeout time.Duration `yaml:"timeout,omitempty"`
}

type Test struct {
	Name     string   `yaml:"name" validate:"required"`
	Request  Request  `yaml:"request" validate:"required"`
	Expected Expected `yaml:"expected" validate:"required"`
}

type Request struct {
	Method  string            `yaml:"method" validate:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	Path    string            `yaml:"path" validate:"required"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}

type Expected struct {
	Status  []int             `yaml:"status" validate:"required,min=1"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    *BodyExpected     `yaml:"body,omitempty"`
}

type BodyExpected struct {
	Contains    []string `yaml:"contains,omitempty"`
	NotContains []string `yaml:"not_contains,omitempty"`
	Exact       string   `yaml:"exact,omitempty"`
	Regex       string   `yaml:"regex,omitempty"`
}