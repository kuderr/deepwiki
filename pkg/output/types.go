package output

import (
	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
	outputgen "github.com/deepwiki-cli/deepwiki-cli/pkg/output/generator"
)

// Manager interface defines the contract for output generation
type Manager interface {
	GenerateOutput(
		structure *generator.WikiStructure,
		pages map[string]*generator.WikiPage,
		options outputgen.OutputOptions,
	) (*outputgen.OutputResult, error)
}
