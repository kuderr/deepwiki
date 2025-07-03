package prompts

import (
	"strings"
	"testing"
)

func TestTemplateManager(t *testing.T) {
	tm := NewTemplateManager()

	// Test registering a simple template
	err := tm.RegisterTemplate("test", "Hello {{.Name}}!")
	if err != nil {
		t.Fatalf("Failed to register template: %v", err)
	}

	// Test executing the template
	result, err := tm.Execute("test", map[string]string{"Name": "World"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := "Hello World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test non-existent template
	_, err = tm.Execute("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}

func TestDefaultManager(t *testing.T) {
	tm := GetDefaultManager()

	// Check that templates are registered
	templates := tm.ListTemplates()
	expectedTemplates := []string{"wiki_structure", "page_content"}

	for _, expected := range expectedTemplates {
		found := false
		for _, tmpl := range templates {
			if tmpl == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected template %s to be registered", expected)
		}
	}
}

func TestWikiStructurePrompt(t *testing.T) {
	data := WikiStructureData{
		FileTree:    "src/\n  main.go\n  config.go\n",
		ReadmeFile:  "# Test Project\nThis is a test project.",
		ProjectName: "test-project",
	}

	result, err := ExecuteWikiStructurePrompt(data)
	if err != nil {
		t.Fatalf("Failed to execute wiki structure prompt: %v", err)
	}

	// Check that the template was properly executed
	if !strings.Contains(result, "test-project") {
		t.Error("Result should contain project name")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("Result should contain file tree content")
	}
	if !strings.Contains(result, "This is a test project") {
		t.Error("Result should contain README content")
	}
}

func TestPageContentPrompt(t *testing.T) {
	data := PageContentData{
		PageTitle:       "Core Architecture",
		PageDescription: "Overview of the system architecture",
		RelevantFiles:   "main.go:\npackage main\n\nfunc main() {}\n",
		ProjectName:     "test-project",
		Language:        "en",
		FileTree:        "src/\n  main.go\n",
	}

	result, err := ExecutePageContentPrompt(data)
	if err != nil {
		t.Fatalf("Failed to execute page content prompt: %v", err)
	}

	// Check that the template was properly executed
	if !strings.Contains(result, "Core Architecture") {
		t.Error("Result should contain page title")
	}
	if !strings.Contains(result, "test-project") {
		t.Error("Result should contain project name")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("Result should contain relevant files")
	}
}

func TestTemplateRegistrationError(t *testing.T) {
	tm := NewTemplateManager()

	// Test registering template with invalid syntax
	err := tm.RegisterTemplate("invalid", "Hello {{.Name")
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}

func TestTemplateExecution(t *testing.T) {
	tm := NewTemplateManager()

	// Register template with multiple variables
	err := tm.RegisterTemplate("multi", "{{.Greeting}} {{.Name}}! You have {{.Count}} messages.")
	if err != nil {
		t.Fatalf("Failed to register template: %v", err)
	}

	data := map[string]interface{}{
		"Greeting": "Hello",
		"Name":     "Alice",
		"Count":    5,
	}

	result, err := tm.Execute("multi", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := "Hello Alice! You have 5 messages."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
