package prompts

import "github.com/kuderr/deepwiki/pkg/types"

// WikiStructureData contains data for wiki structure generation
type WikiStructureData struct {
	FileTree    string
	ReadmeFile  string
	ProjectName string
	Language    types.Language
}

// WikiStructurePrompt is the template for generating wiki structure
const WikiStructurePrompt = `
You are an expert technical writer and information-architect.

Goal → Design a comprehensive, non-overlapping wiki for **{{.ProjectName}}**.  
All pages assume diagrams are required.

# INPUTS
<file_tree>
{{.FileTree}}
</file_tree>

<readme>
{{.ReadmeFile}}
</readme>

# OUTPUT  (return exactly this XML)
<wiki_structure>
  <title>{{.ProjectName}} Documentation</title>
  
  <!-- Generate description -->
  <description>[Brief description of the project and this wiki]</description>
  <pages>
    <page>
      <id>unique-kebab-case-id</id>
      <title>[Page Title]</title>
      <!-- Generate description -->
      <description>[Brief description of what this page covers]</description>

      <importance>high|medium|low</importance>
      <parent_id>optional-parent-id</parent_id>
    </page>
    <!-- More pages... -->
  </pages>
</wiki_structure>

# RULES
1. Depending on a project create a structured wiki covering all essential aspects, which you consider necessary:
    - **Overview and Introduction**: Project purpose, key features, getting started
    - **System Architecture**: High-level design, components, data flow
    - **Core Features**: Main functionality, use cases, examples
    - **Data Management**: Storage, processing, models, schemas
    - **Components and Modules**: Detailed breakdown of major components
    - **API and Services**: External interfaces, endpoints, integration points
    - **Deployment and Configuration**: Setup, environment, deployment strategies
    - **Development and Extension**: Contributing, customization, plugin development
    - **Operations/Tooling**: configuration, local development, ci, linting, tests
2. **Zero duplication**: a topic appears once; cross-link elsewhere.
3. **Compact depth**: 2 levels preferred; 3 only if essential.
4. **Language**: titles, descriptions, notes in **{{.Language}}**.
5. Focus on technical documentation that developers would need
6. Ensure comprehensive coverage without redundancy
7. When designing the wiki structure, include pages that would benefit from visual diagrams, such as:
  - Architecture overviews
  - Data flow descriptions
  - Component relationships
  - Process workflows
  - State machines
  - Class hierarchies

# BEFORE RETURNING
**Self-check** before output:
   – every top-level page is high;  
   – no overlap;  
`

// RegisterWikiStructurePrompt registers the wiki structure prompt template
func RegisterWikiStructurePrompt(tm *TemplateManager) error {
	return tm.RegisterTemplate("wiki_structure", WikiStructurePrompt)
}
