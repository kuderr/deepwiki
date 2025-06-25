package prompts

// PageContentData contains data for page content generation
type PageContentData struct {
	PageTitle       string
	PageDescription string
	RelevantFiles   string
	ProjectName     string
	Language        string
	FileTree        string
}

// PageContentPrompt is the template for generating wiki page content
const PageContentPrompt = `You are an expert technical writer generating comprehensive documentation for the "{{.PageTitle}}" section of {{.ProjectName}}.

Page Description: {{.PageDescription}}

Generate a wiki page about "{{.PageTitle}}" based ONLY on these source files:

{{.RelevantFiles}}

Requirements:
1. **Start with a <details> block** listing ALL source files used (minimum 5 files when available):
   <details>
   <summary>📁 Source Files</summary>

   - [filename1.ext](path/to/file1.ext)
   - [filename2.ext](path/to/file2.ext)
   - [filename3.ext](path/to/file3.ext)
   - ...

   </details>

2. **Use extensive Mermaid diagrams** throughout the page:
   - Use flowchart TD for process flows and architecture
   - Use sequenceDiagram for interactions and API calls
   - Use classDiagram for object relationships and data structures
   - Use graph LR for dependencies and connections
   - Place diagrams strategically to illustrate concepts

3. **Include comprehensive code snippets** with proper syntax highlighting:
   - Show key functions, classes, and configurations
   - Include usage examples where relevant
   - Use appropriate language tags (go, python, javascript, etc.)

4. **Cite sources extensively** using this format: [filename.ext:line_numbers](#)
   - Include specific line numbers or ranges
   - Cite for every major statement or code reference
   - Examples: [main.go:15-25](#), [config.py:42](#), [api.ts:108-120](#)

5. **Use markdown tables** for structured data:
   - Configuration options
   - API endpoints
   - Function parameters
   - Error codes

6. **Ensure technical accuracy** - base all content on actual source code
   - Don't assume functionality not shown in the files
   - Explain what the code actually does, not what it might do
   - Include error handling and edge cases when visible

7. **Structure the content** with clear sections:
   - Brief overview
   - Key concepts and architecture
   - Detailed implementation
   - Usage examples
   - Configuration and setup
   - Troubleshooting (if applicable)

8. **Focus on the {{.Language}} audience** - adjust technical depth accordingly

Write comprehensive, developer-focused documentation that thoroughly explains this aspect of the project.`

// RegisterPageContentPrompt registers the page content prompt template
func RegisterPageContentPrompt(tm *TemplateManager) error {
	return tm.RegisterTemplate("page_content", PageContentPrompt)
}
