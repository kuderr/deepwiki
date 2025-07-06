package prompts

import "github.com/kuderr/deepwiki/pkg/types"

// PageContentData contains data for page content generation
type PageContentData struct {
	Title         string
	Description   string
	RelevantFiles string
	ProjectName   string
	Language      types.Language
	FileTree      string
	OtherPages    []PageSummary
}

type PageSummary struct {
	Title       string
	Description string
}

// PageContentPrompt is the template for generating wiki page content
const PageContentPrompt = `
You are an expert technical writer and software architect.

Task → Write the **{{.Title}}** page for {{.ProjectName}}.
Generate everything in **{{.Language}}** and include diagrams.

# PAGE SCOPE (from outline)
<page_description>
{{.Description}}
</page_description>

# SOURCES (the only ground truth)
<relevant_files>
{{.RelevantFiles}}
</relevant_files>

<file_tree>
{{.FileTree}}
</file_tree>

# PAGE PLAN
### 1. Overview  (≤ 100 words)

### 2. Mermaid Diagram(s)  *(always present)*
   – Choose among: flowchart TD, classDiagram, sequenceDiagram, stateDiagram, graph LR.  
   – Keep each diagram ≤ 12 nodes; split if larger.

### 3. Key Concepts / Responsibilities

### 4. Implementation Details  
   – Code snippets with syntax highlighting and inline comments.  
   – Show edge-case handling.

### 5. Usage Examples  *(runnable or copy-paste ready)*

### 6. Reference Tables (only when useful)  
   – API endpoints, function params, error codes, config opts (if this *is* the config page).

### 7. Troubleshooting / Gotchas  *(optional; only if code reveals error handling)*

# Registry of already-covered pages (do not repeat)
<other_pages>
  <pages>
	{{range $page := .OtherPages }}
    <page>
      <title>{{$page.Title}}</title>
      <description>{{$page.Description}}</description>
    </page>
	{{end}}
  </pages>
</other_pages>

If you need to mention a topic that belongs to an entry in <other_pages>, replace detailed text with md link.

# HARD RULES
1. **Truth-only**: do not invent behaviour; base every statement on the source.
2. **No duplication**: if another page covers a topic, replace with with md link.
3. **Skip build/lint/CI chatter** unless this page’s title implies it.
4. **Mermaid validity** required; diagrams must compile.
5. **Length target**: 600–1200 words.
6. **Language**: all prose, code comments, tables in **{{.Language}}**.
7. All page requirements must comply with <page_description>. If you need to go beyond this, insert a link to another page instead of duplicating.
8. **Output**: return only valid markdown syntax content

# BEFORE RETURNING
**Self-check** before output:
- All sections present & scoped correctly  
- Code blocks run/compile  
- Diagram syntax valid  
`

// RegisterPageContentPrompt registers the page content prompt template
func RegisterPageContentPrompt(tm *TemplateManager) error {
	return tm.RegisterTemplate("page_content", PageContentPrompt)
}
