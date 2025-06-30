package output

import (
	"fmt"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
	outputgen "github.com/deepwiki-cli/deepwiki-cli/pkg/output/generator"
)

// OutputManager implements the Manager interface using a registry of format generators
type OutputManager struct {
	registry *outputgen.Registry
}

// NewOutputManager creates a new OutputManager instance with all generators registered
func NewOutputManager() *OutputManager {
	registry := outputgen.NewRegistry()

	// Register all available format generators
	registry.Register(outputgen.NewMarkdownGenerator())
	registry.Register(outputgen.NewJSONGenerator())
	registry.Register(outputgen.NewDocusaurus2Generator())
	registry.Register(outputgen.NewDocusaurus3Generator())
	registry.Register(outputgen.NewSimpleDocusaurus2Generator())
	registry.Register(outputgen.NewSimpleDocusaurus3Generator())

	return &OutputManager{
		registry: registry,
	}
}

// GenerateOutput generates output in the specified format using the registry
func (om *OutputManager) GenerateOutput(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	options outputgen.OutputOptions,
) (*outputgen.OutputResult, error) {
	// Get the generator for the specified format
	gen, err := om.registry.Get(options.Format)
	if err != nil {
		return nil, fmt.Errorf("unsupported output format: %s", options.Format)
	}

	// Generate output using the appropriate generator
	return gen.Generate(structure, pages, options)
}

// GetRegistry returns the format generator registry
func (om *OutputManager) GetRegistry() *outputgen.Registry {
	return om.registry
}

// ListFormats returns all available output formats
func (om *OutputManager) ListFormats() []outputgen.OutputFormat {
	return om.registry.List()
}

// GetFormatDescription returns the description for a specific format
func (om *OutputManager) GetFormatDescription(format outputgen.OutputFormat) (string, error) {
	gen, err := om.registry.Get(format)
	if err != nil {
		return "", err
	}
	return gen.Description(), nil
}
