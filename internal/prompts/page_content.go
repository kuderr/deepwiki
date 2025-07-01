package prompts

import "github.com/kuderr/deepwiki/pkg/types"

// PageContentData contains data for page content generation
type PageContentData struct {
	PageTitle       string
	PageDescription string
	RelevantFiles   string
	ProjectName     string
	Language        types.Language
	FileTree        string
}

// PageContentPrompt is the template for generating wiki page content
const PageContentPrompt = `
You are an expert technical writer and software architect.
Your task is to generate a comprehensive and accurate technical wiki page about a specific feature, system, or module for the "{{.PageTitle}}" section of {{.ProjectName}}.

Page Description: {{.PageDescription}}

Generate a wiki page about based ONLY on these source files:

{{.RelevantFiles}}
</relevant_files>

The complete file tree:
<file_tree>
{{.FileTree}}
</file_tree>

Requirements:
1. **Use extensive Mermaid diagrams** throughout the page:
   - Use flowchart TD for process flows and architecture
   - Use sequenceDiagram for interactions and API calls
   - Use classDiagram for object relationships and data structures
   - Use graph LR for dependencies and connections
   - Place diagrams strategically to illustrate concepts
   - Use only valid Mermaid syntax

2. **Include comprehensive code snippets** with proper syntax highlighting:
   - Show key functions, classes, and configurations
   - Include usage examples where relevant
   - Use appropriate language tags (go, python, javascript, etc.)

3. **Use markdown tables** for structured data:
   - Configuration options
   - API endpoints
   - Function parameters
   - Error codes

4. **Ensure technical accuracy** - base all content on actual source code
   - Don't assume functionality not shown in the files
   - Don't make up information not based on the relevant files.
   - Don't make up recommendations not based on the relevant files
   - Explain what the code actually does, not what it might do
   - Include error handling and edge cases when visible

5. **Structure the content** with clear sections:
   - Brief overview
   - Key concepts and architecture
   - Detailed implementation
   - Usage examples
   - Configuration and setup
   - Troubleshooting (if applicable)

6. **Clarity and Conciseness:** Use clear, professional, and concise technical language suitable for other developers working on or learning about the project. Avoid unnecessary jargon, but use correct technical terms where appropriate.

7. **Generate the content in {{.Language}} language**

Write comprehensive, developer-focused documentation that thoroughly explains this aspect of the project.
`

// RegisterPageContentPrompt registers the page content prompt template
func RegisterPageContentPrompt(tm *TemplateManager) error {
	return tm.RegisterTemplate("page_content", PageContentPrompt)
}
