package generator

import (
	"fmt"
	"sync"
)

// Registry manages all available format generators
type Registry struct {
	generators map[OutputFormat]FormatGenerator
	mu         sync.RWMutex
}

// NewRegistry creates a new generator registry
func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[OutputFormat]FormatGenerator),
	}
}

// Register adds a format generator to the registry
func (r *Registry) Register(generator FormatGenerator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators[generator.FormatType()] = generator
}

// Get retrieves a format generator by type
func (r *Registry) Get(format OutputFormat) (FormatGenerator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	generator, exists := r.generators[format]
	if !exists {
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}

	return generator, nil
}

// List returns all registered format types
func (r *Registry) List() []OutputFormat {
	r.mu.RLock()
	defer r.mu.RUnlock()

	formats := make([]OutputFormat, 0, len(r.generators))
	for format := range r.generators {
		formats = append(formats, format)
	}

	return formats
}

// GetAll returns all registered generators
func (r *Registry) GetAll() map[OutputFormat]FormatGenerator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[OutputFormat]FormatGenerator)
	for format, generator := range r.generators {
		result[format] = generator
	}

	return result
}
