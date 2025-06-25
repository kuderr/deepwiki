package prompts

// WikiStructureData contains data for wiki structure generation
type WikiStructureData struct {
	FileTree    string
	ReadmeFile  string
	ProjectName string
	Language    string
}

// TODO(kuderr): make pages count configurable

// WikiStructurePrompt is the template for generating wiki structure
const WikiStructurePrompt = `You are an expert technical writer creating a comprehensive wiki structure for a software project.

Analyze this local directory and create a 5 page wiki structure for it.

1. The complete file tree:
<file_tree>
{{.FileTree}}
</file_tree>

2. The README file (if available):
<readme>
{{.ReadmeFile}}
</readme>

Project Name: {{.ProjectName}}
Target Language: {{.Language}}

Create a structured wiki with covering these essential areas:
- **Overview and Introduction**: Project purpose, key features, getting started
- **System Architecture**: High-level design, components, data flow
- **Core Features**: Main functionality, use cases, examples
- **Data Management**: Storage, processing, models, schemas
- **Components and Modules**: Detailed breakdown of major components
- **API and Services**: External interfaces, endpoints, integration points
- **Deployment and Configuration**: Setup, environment, deployment strategies
- **Development and Extension**: Contributing, customization, plugin development

Requirements:
1. Each page should have a clear, descriptive title
2. Pages should be logically organized with parent-child relationships where appropriate
3. Include importance levels: "high" (core functionality), "medium" (important features), "low" (auxiliary)
4. Ensure comprehensive coverage without redundancy
5. Focus on technical documentation that developers would need

Return your response in this XML format:
<wiki_structure>
  <title>{{.ProjectName}} Documentation</title>
  <description>Brief description of the project and this wiki</description>
  <pages>
    <page>
      <id>unique-identifier</id>
      <title>Page Title</title>
      <importance>high|medium|low</importance>
      <parent_id>parent-page-id (optional)</parent_id>
      <description>Brief description of what this page covers</description>
    </page>
    <!-- More pages... -->
  </pages>
</wiki_structure>`

// RegisterWikiStructurePrompt registers the wiki structure prompt template
func RegisterWikiStructurePrompt(tm *TemplateManager) error {
	return tm.RegisterTemplate("wiki_structure", WikiStructurePrompt)
}
