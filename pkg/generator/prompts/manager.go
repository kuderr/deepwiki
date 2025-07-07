package prompts

import (
	"fmt"
	"sync"
)

var (
	defaultManager *TemplateManager
	once           sync.Once
)

// GetDefaultManager returns the default template manager instance
func GetDefaultManager() *TemplateManager {
	once.Do(func() {
		defaultManager = NewTemplateManager()
		initializeDefaultPrompts(defaultManager)
	})
	return defaultManager
}

// initializeDefaultPrompts registers all default prompt templates
func initializeDefaultPrompts(tm *TemplateManager) {
	// Register wiki structure prompt
	if err := RegisterWikiStructurePrompt(tm); err != nil {
		panic("failed to register wiki structure prompt: " + err.Error())
	}

	// Register page content prompt
	if err := RegisterPageContentPrompt(tm); err != nil {
		fmt.Println(err.Error())
		panic("failed to register page content prompt: " + err.Error())
	}
}

// ExecuteWikiStructurePrompt executes the wiki structure generation prompt
func ExecuteWikiStructurePrompt(data WikiStructureData) (string, error) {
	return GetDefaultManager().Execute("wiki_structure", data)
}

// ExecutePageContentPrompt executes the page content generation prompt
func ExecutePageContentPrompt(data PageContentData) (string, error) {
	return GetDefaultManager().Execute("page_content", data)
}
