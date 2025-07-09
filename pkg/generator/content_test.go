package generator

import (
	"testing"
)

func TestContentProcessor_cleanThinkBlocks(t *testing.T) {
	cp := NewContentProcessor()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple think block",
			input:    "Some text <think>this is thinking</think> more text",
			expected: "Some text  more text",
		},
		{
			name:     "multiline think block",
			input:    "Before\n<think>\nThis is multiline\nthinking content\n</think>\nAfter",
			expected: "Before\n\nAfter",
		},
		{
			name:     "multiple think blocks",
			input:    "Text <think>first</think> middle <think>second</think> end",
			expected: "Text  middle  end",
		},
		{
			name:     "nested content in think block",
			input:    "Text <think>Some <em>nested</em> content</think> end",
			expected: "Text  end",
		},
		{
			name:     "no think blocks",
			input:    "Just regular text with no think blocks",
			expected: "Just regular text with no think blocks",
		},
		{
			name:     "empty think block",
			input:    "Text <think></think> more text",
			expected: "Text  more text",
		},
		{
			name:     "think block with whitespace",
			input:    "Text <think>   \n  \n  </think> more text",
			expected: "Text  more text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cp.cleanThinkBlocks(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestContentProcessor_cleanEmptyCodeBlocks(t *testing.T) {
	cp := NewContentProcessor()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty code block",
			input:    "Text\n```\n```\nMore text",
			expected: "Text\n\nMore text",
		},
		{
			name:     "empty code block with whitespace",
			input:    "Text\n```   \n   ```\nMore text",
			expected: "Text\n\nMore text",
		},
		{
			name:     "multiple empty code blocks",
			input:    "Text\n```\n```\nMiddle\n```\n```\nEnd",
			expected: "Text\n\nMiddle\n\nEnd",
		},
		{
			name:     "empty code block with language",
			input:    "Text\n```go\n```\nMore text",
			expected: "Text\n\nMore text",
		},
		{
			name:     "non-empty code block",
			input:    "Text\n```\nsome code\n```\nMore text",
			expected: "Text\n```\nsome code\n```\nMore text",
		},
		{
			name:     "no code blocks",
			input:    "Just regular text",
			expected: "Just regular text",
		},
		{
			name:     "mixed empty and non-empty code blocks",
			input:    "Text\n```\n```\nMiddle\n```\ncode here\n```\nEnd",
			expected: "Text\n\nMiddle\n```\ncode here\n```\nEnd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cp.cleanEmptyCodeBlocks(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestContentProcessor_cleanMarkdownCodeBlocks(t *testing.T) {
	cp := NewContentProcessor()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple markdown block",
			input:    "Text\n````markdown\n# Header\nContent\n````\nMore text",
			expected: "Text\n# Header\nContent\nMore text",
		},
		{
			name:     "markdown block with whitespace",
			input:    "Text\n````markdown   \n# Header\nContent\n````\nMore text",
			expected: "Text\n# Header\nContent\nMore text",
		},
		{
			name:     "multiple markdown blocks",
			input:    "Text\n````markdown\nFirst\n````\nMiddle\n````markdown\nSecond\n````\nEnd",
			expected: "Text\nFirst\nMiddle\nSecond\nEnd",
		},
		{
			name:     "markdown block with complex content",
			input:    "Text\n````markdown\n# Header\n\n```go\nfunc test() {}\n```\n\nMore content\n````\nEnd",
			expected: "Text\n# Header\n\n```go\nfunc test() {}\n```\n\nMore content\nEnd",
		},
		{
			name:     "no markdown blocks",
			input:    "Just regular text with ```code``` blocks",
			expected: "Just regular text with ```code``` blocks",
		},
		{
			name:     "empty markdown block",
			input:    "Text\n````markdown\n\n````\nMore text",
			expected: "Text\n\nMore text",
		},
		{
			name:     "markdown block with only whitespace",
			input:    "Text\n````markdown\n   \n  \n````\nMore text",
			expected: "Text\n  \nMore text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cp.cleanMarkdownCodeBlocks(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestContentProcessor_CleanMarkdown_Integration(t *testing.T) {
	cp := NewContentProcessor()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "all cleanup functions together",
			input: `Some text
<think>
This is thinking content
that should be removed
</think>

` + "````markdown" + `
# Header
Content here
` + "````" + `

` + "```" + `

` + "```" + `

More content`,
			expected: "Some text\n\n# Header\nContent here\n\nMore content",
		},
		{
			name: "mixed content with all cleanup types",
			input: `Introduction

<think>Planning this section</think>

` + "````markdown" + `
## Section 1
Some content
` + "````" + `

` + "```" + `
` + "```" + `

Final text`,
			expected: "Introduction\n\n## Section 1\nSome content\n\nFinal text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cp.CleanMarkdown(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}