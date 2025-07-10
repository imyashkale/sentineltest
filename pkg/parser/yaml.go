package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"waf-tester/pkg/config"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	validator *validator.Validate
}

func NewParser() *Parser {
	return &Parser{
		validator: validator.New(),
	}
}

func (p *Parser) ParseFile(filename string) (*config.WafTest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return p.ParseYAML(data)
}

func (p *Parser) ParseYAML(data []byte) (*config.WafTest, error) {
	var wafTest config.WafTest
	
	if err := yaml.Unmarshal(data, &wafTest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := p.validator.Struct(&wafTest); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &wafTest, nil
}

func (p *Parser) ParseDirectory(dir string) ([]*config.WafTest, error) {
	var tests []*config.WafTest
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
			test, parseErr := p.ParseFile(path)
			if parseErr != nil {
				return fmt.Errorf("failed to parse %s: %w", path, parseErr)
			}
			tests = append(tests, test)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory %s: %w", dir, err)
	}
	
	return tests, nil
}