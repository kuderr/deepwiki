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
  <description>One-sentence project + wiki purpose.</description>
  <pages>
    <!-- repeat -->
    <page>
      <id>kebab-case-id</id>
      <title>Page Title (<= 6 words)</title>
      <importance>high|medium|low</importance>
      <parent_id>optional-parent-id</parent_id>

      <!-- WHY this page exists; max 25 words; user-visible -->
      <description>…</description>
    </page>
  </pages>
</wiki_structure>

# RULES
1. **Root = core**: only high-importance pages at root (Overview, Architecture, Core APIs, Data Models).
   Move configs, CI, linting, deployment under a single low-importance subtree (e.g. Operations/Tooling).
2. **Zero duplication**: a topic appears once; cross-link elsewhere.
3. **Compact depth**: 2 levels preferred; 3 only if essential.
4. **Language**: titles, descriptions, notes in **{{.Language}}**.
5. **Self-check** before output:
   – every top-level page is high;  
   – no overlap;  
   – each page has needs_diagram=true.
`

// RegisterWikiStructurePrompt registers the wiki structure prompt template
func RegisterWikiStructurePrompt(tm *TemplateManager) error {
	return tm.RegisterTemplate("wiki_structure", WikiStructurePrompt)
}
