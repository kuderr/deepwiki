package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
)

// Docusaurus3Generator generates Docusaurus v3.8.x-formatted output files
type Docusaurus3Generator struct{}

// NewDocusaurus3Generator creates a new Docusaurus v3 generator
func NewDocusaurus3Generator() *Docusaurus3Generator {
	return &Docusaurus3Generator{}
}

// FormatType returns the format type this generator handles
func (d3g *Docusaurus3Generator) FormatType() OutputFormat {
	return FormatDocusaurus3
}

// Description returns a human-readable description of the format
func (d3g *Docusaurus3Generator) Description() string {
	return "Docusaurus v3.8.x site with TypeScript configuration and React v18"
}

// Generate creates Docusaurus v3.8.x output files
func (d3g *Docusaurus3Generator) Generate(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	options OutputOptions,
) (*OutputResult, error) {
	startTime := time.Now()

	// Create Docusaurus directory structure
	if err := d3g.organizeDocusaurusFiles(options.Directory); err != nil {
		return nil, fmt.Errorf("failed to create Docusaurus directory structure: %w", err)
	}

	var filesGenerated []string
	var totalSize int64
	var errors []error

	// Generate intro page (Docusaurus home)
	introPath := filepath.Join(options.Directory, "docs", "intro.md")
	if err := d3g.generateIndex(structure, pages, introPath, options); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate intro: %w", err))
	} else {
		if stat, err := os.Stat(introPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, introPath)
	}

	// Generate individual page files with Docusaurus frontmatter
	docsDir := filepath.Join(options.Directory, "docs")
	for pageID, page := range pages {
		fileName := d3g.sanitizeFileName(page.Title) + ".md"
		pagePath := filepath.Join(docsDir, fileName)

		if err := d3g.generatePage(page, pagePath, structure, options); err != nil {
			errors = append(errors, fmt.Errorf("failed to generate page %s: %w", pageID, err))
			continue
		}

		if stat, err := os.Stat(pagePath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, pagePath)
	}

	// Generate sidebars.ts configuration (TypeScript for v3)
	sidebarPath := filepath.Join(options.Directory, "sidebars.ts")
	if err := d3g.generateSidebar(structure, pages, sidebarPath, options); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate sidebar: %w", err))
	} else {
		if stat, err := os.Stat(sidebarPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, sidebarPath)
	}

	// Generate docusaurus.config.ts (TypeScript for v3)
	configPath := filepath.Join(options.Directory, "docusaurus.config.ts")
	if err := d3g.generateConfig(structure, options, configPath); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate Docusaurus config: %w", err))
	} else {
		if stat, err := os.Stat(configPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, configPath)
	}

	// Generate package.json for Docusaurus v3 dependencies
	packagePath := filepath.Join(options.Directory, "package.json")
	if err := d3g.generatePackageJSON(structure, options, packagePath); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate package.json: %w", err))
	} else {
		if stat, err := os.Stat(packagePath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, packagePath)
	}

	// Generate tsconfig.json for TypeScript support
	tsconfigPath := filepath.Join(options.Directory, "tsconfig.json")
	if err := d3g.generateTSConfig(tsconfigPath); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate tsconfig.json: %w", err))
	} else {
		if stat, err := os.Stat(tsconfigPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, tsconfigPath)
	}

	return &OutputResult{
		OutputDir:      options.Directory,
		FilesGenerated: filesGenerated,
		TotalFiles:     len(filesGenerated),
		TotalSize:      totalSize,
		GeneratedAt:    time.Now(),
		ProcessingTime: time.Since(startTime),
		Errors:         errors,
	}, nil
}

// organizeDocusaurusFiles creates the Docusaurus directory structure
func (d3g *Docusaurus3Generator) organizeDocusaurusFiles(outputDir string) error {
	dirs := []string{
		outputDir,
		filepath.Join(outputDir, "docs"),
		filepath.Join(outputDir, "static"),
		filepath.Join(outputDir, "static", "img"),
		filepath.Join(outputDir, "src"),
		filepath.Join(outputDir, "src", "components"),
		filepath.Join(outputDir, "src", "pages"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateIndex generates the intro.md file for Docusaurus v3.8.x
func (d3g *Docusaurus3Generator) generateIndex(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	filePath string,
	options OutputOptions,
) error {
	var content strings.Builder

	// Write Docusaurus v3 frontmatter with enhanced features
	content.WriteString("---\n")
	content.WriteString("sidebar_position: 1\n")
	content.WriteString("slug: /\n")
	content.WriteString(fmt.Sprintf("title: %s\n", structure.Title))
	content.WriteString(fmt.Sprintf("description: %s\n", structure.Description))
	content.WriteString("displayed_sidebar: tutorialSidebar\n")
	content.WriteString("---\n\n")

	// Write header
	content.WriteString(fmt.Sprintf("# %s\n\n", structure.Title))
	content.WriteString(fmt.Sprintf("%s\n\n", structure.Description))

	// Write metadata
	content.WriteString("## 📊 Documentation Information\n\n")
	content.WriteString(fmt.Sprintf("- **Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("- **Project**: %s\n", options.ProjectName))
	content.WriteString(fmt.Sprintf("- **Language**: %s\n", options.Language))
	content.WriteString(fmt.Sprintf("- **Total Pages**: %d\n", len(pages)))
	content.WriteString(fmt.Sprintf("- **Version**: %s\n\n", structure.Version))

	// Write quick navigation
	content.WriteString("## 🚀 Quick Start\n\n")
	content.WriteString(
		"Explore the documentation using the sidebar navigation or start with the high-priority pages below.\n\n",
	)

	// Group pages by importance for quick access
	importanceGroups := map[string][]*generator.WikiPage{
		"high":   {},
		"medium": {},
		"low":    {},
	}

	for _, page := range pages {
		importance := page.Importance
		if importance == "" {
			importance = "medium"
		}
		importanceGroups[importance] = append(importanceGroups[importance], page)
	}

	// Write high importance pages
	if len(importanceGroups["high"]) > 0 {
		content.WriteString("### 🔥 Essential Documentation\n\n")
		for _, page := range importanceGroups["high"] {
			fileName := d3g.sanitizeFileName(page.Title)
			content.WriteString(fmt.Sprintf("- [%s](./%s) - %s\n", page.Title, fileName, page.Description))
		}
		content.WriteString("\n")
	}

	// Write medium importance pages
	if len(importanceGroups["medium"]) > 0 {
		content.WriteString("### 📋 Core Documentation\n\n")
		for _, page := range importanceGroups["medium"] {
			fileName := d3g.sanitizeFileName(page.Title)
			content.WriteString(fmt.Sprintf("- [%s](./%s) - %s\n", page.Title, fileName, page.Description))
		}
		content.WriteString("\n")
	}

	// Write footer
	content.WriteString("---\n\n")
	content.WriteString("*Generated by [DeepWiki CLI](https://github.com/deepwiki-cli/deepwiki-cli)*\n")

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generatePage generates a Docusaurus v3.8.x-formatted page
func (d3g *Docusaurus3Generator) generatePage(
	page *generator.WikiPage,
	filePath string,
	structure *generator.WikiStructure,
	options OutputOptions,
) error {
	var content strings.Builder

	// Generate slug from filename
	fileName := strings.TrimSuffix(filepath.Base(filePath), ".md")

	// Write enhanced Docusaurus v3 frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("id: %s\n", page.ID))
	content.WriteString(fmt.Sprintf("title: %s\n", page.Title))
	content.WriteString(fmt.Sprintf("slug: /%s\n", fileName))
	if page.Description != "" {
		content.WriteString(fmt.Sprintf("description: %s\n", page.Description))
	}
	content.WriteString("displayed_sidebar: tutorialSidebar\n")
	if len(page.FilePaths) > 0 {
		content.WriteString("tags:\n")
		// Add language tags based on file extensions
		languages := make(map[string]bool)
		for _, path := range page.FilePaths {
			ext := filepath.Ext(path)
			switch ext {
			case ".go":
				languages["golang"] = true
			case ".js", ".jsx":
				languages["javascript"] = true
			case ".ts", ".tsx":
				languages["typescript"] = true
			case ".py":
				languages["python"] = true
			case ".java":
				languages["java"] = true
			case ".cpp", ".cc", ".cxx":
				languages["cpp"] = true
			case ".rs":
				languages["rust"] = true
			case ".md":
				languages["documentation"] = true
			}
		}
		for lang := range languages {
			content.WriteString(fmt.Sprintf("  - %s\n", lang))
		}
	}
	content.WriteString("---\n\n")

	// Write description if available
	if page.Description != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", page.Description))
	}

	// Write main content
	content.WriteString(page.Content)
	content.WriteString("\n\n")

	// Write metadata section
	content.WriteString("## 📋 Page Information\n\n")
	content.WriteString(fmt.Sprintf("- **Page ID**: `%s`\n", page.ID))
	content.WriteString(fmt.Sprintf("- **Importance**: %s\n", page.Importance))
	content.WriteString(fmt.Sprintf("- **Word Count**: %d\n", page.WordCount))
	content.WriteString(fmt.Sprintf("- **Source Files**: %d\n", page.SourceFiles))
	content.WriteString(fmt.Sprintf("- **Generated**: %s\n", page.CreatedAt.Format("2006-01-02 15:04:05")))

	if len(page.FilePaths) > 0 {
		content.WriteString("\n### Source Files\n\n")
		for _, path := range page.FilePaths {
			content.WriteString(fmt.Sprintf("- `%s`\n", path))
		}
	}

	if len(page.RelatedPages) > 0 {
		content.WriteString("\n### Related Pages\n\n")
		for _, relatedID := range page.RelatedPages {
			content.WriteString(fmt.Sprintf("- %s\n", relatedID))
		}
	}

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generateSidebar generates the sidebars.ts configuration for v3.8.x
func (d3g *Docusaurus3Generator) generateSidebar(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	filePath string,
	options OutputOptions,
) error {
	var content strings.Builder

	content.WriteString("import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';\n\n")
	content.WriteString("/**\n")
	content.WriteString(" * Creating a sidebar enables you to:\n")
	content.WriteString(" - create an ordered group of docs\n")
	content.WriteString(" - render a sidebar for each doc of that group\n")
	content.WriteString(" - provide next/previous navigation\n")
	content.WriteString(" *\n")
	content.WriteString(" * The sidebars can be generated from the filesystem, or explicitly defined here.\n")
	content.WriteString(" *\n")
	content.WriteString(" * Create as many sidebars as you want.\n")
	content.WriteString(" */\n\n")

	content.WriteString("const sidebars: SidebarsConfig = {\n")
	content.WriteString("  // Auto-generated sidebar\n")
	content.WriteString("  tutorialSidebar: [\n")
	content.WriteString("    'intro',\n")

	// Group pages by importance
	importanceGroups := map[string][]*generator.WikiPage{
		"high":   {},
		"medium": {},
		"low":    {},
	}

	for _, page := range pages {
		importance := page.Importance
		if importance == "" {
			importance = "medium"
		}
		importanceGroups[importance] = append(importanceGroups[importance], page)
	}

	// Add high importance section
	if len(importanceGroups["high"]) > 0 {
		content.WriteString("    {\n")
		content.WriteString("      type: 'category',\n")
		content.WriteString("      label: '🔥 Essential Documentation',\n")
		content.WriteString("      collapsed: false,\n")
		content.WriteString("      items: [\n")
		for _, page := range importanceGroups["high"] {
			content.WriteString(fmt.Sprintf("        '%s',\n", page.ID))
		}
		content.WriteString("      ],\n")
		content.WriteString("    },\n")
	}

	// Add medium importance section
	if len(importanceGroups["medium"]) > 0 {
		content.WriteString("    {\n")
		content.WriteString("      type: 'category',\n")
		content.WriteString("      label: '📋 Core Documentation',\n")
		content.WriteString("      collapsed: false,\n")
		content.WriteString("      items: [\n")
		for _, page := range importanceGroups["medium"] {
			content.WriteString(fmt.Sprintf("        '%s',\n", page.ID))
		}
		content.WriteString("      ],\n")
		content.WriteString("    },\n")
	}

	// Add low importance section
	if len(importanceGroups["low"]) > 0 {
		content.WriteString("    {\n")
		content.WriteString("      type: 'category',\n")
		content.WriteString("      label: '📝 Additional Information',\n")
		content.WriteString("      collapsed: true,\n")
		content.WriteString("      items: [\n")
		for _, page := range importanceGroups["low"] {
			content.WriteString(fmt.Sprintf("        '%s',\n", page.ID))
		}
		content.WriteString("      ],\n")
		content.WriteString("    },\n")
	}

	content.WriteString("  ],\n")
	content.WriteString("};\n\n")
	content.WriteString("export default sidebars;\n")

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generateConfig generates the docusaurus.config.ts file for v3.8.x
func (d3g *Docusaurus3Generator) generateConfig(
	structure *generator.WikiStructure,
	options OutputOptions,
	filePath string,
) error {
	var content strings.Builder

	content.WriteString("import { themes as prismThemes } from 'prism-react-renderer';\n")
	content.WriteString("import type { Config } from '@docusaurus/types';\n")
	content.WriteString("import type * as Preset from '@docusaurus/preset-classic';\n\n")

	content.WriteString("const config: Config = {\n")
	content.WriteString(fmt.Sprintf("  title: '%s',\n", structure.Title))
	content.WriteString(fmt.Sprintf("  tagline: '%s',\n", structure.Description))
	content.WriteString("  favicon: 'img/favicon.ico',\n\n")

	content.WriteString("  // Set the production url of your site here\n")
	content.WriteString("  url: 'https://your-docusaurus-test-site.com',\n")
	content.WriteString("  // Set the /<baseUrl>/ pathname under which your site is served\n")
	content.WriteString("  baseUrl: '/',\n\n")

	content.WriteString("  // GitHub pages deployment config.\n")
	content.WriteString("  organizationName: 'your-org',\n")
	content.WriteString("  projectName: 'your-project',\n\n")

	content.WriteString("  onBrokenLinks: 'throw',\n")
	content.WriteString("  onBrokenMarkdownLinks: 'warn',\n\n")

	content.WriteString("  // Even if you don't use internationalization, you can use this field to set\n")
	content.WriteString("  // useful metadata like html lang. For example, if your site is Chinese, you\n")
	content.WriteString("  // may want to replace \"en\" with \"zh-Hans\".\n")
	content.WriteString("  i18n: {\n")
	content.WriteString(fmt.Sprintf("    defaultLocale: '%s',\n", options.Language))
	content.WriteString(fmt.Sprintf("    locales: ['%s'],\n", options.Language))
	content.WriteString("  },\n\n")

	// V3 future flags for performance
	content.WriteString("  future: {\n")
	content.WriteString("    experimental_faster: true,\n")
	content.WriteString("    v4: true,\n")
	content.WriteString("  },\n\n")

	// Mermaid support
	content.WriteString("  markdown: {\n")
	content.WriteString("    mermaid: true,\n")
	content.WriteString("  },\n")
	content.WriteString("  themes: ['@docusaurus/theme-mermaid'],\n\n")

	content.WriteString("  presets: [\n")
	content.WriteString("    [\n")
	content.WriteString("      'classic',\n")
	content.WriteString("      {\n")
	content.WriteString("        docs: {\n")
	content.WriteString("          sidebarPath: './sidebars.ts',\n")
	content.WriteString("          routeBasePath: '/',\n")
	content.WriteString("        },\n")
	content.WriteString("        blog: false,\n")
	content.WriteString("      } satisfies Preset.Options,\n")
	content.WriteString("    ],\n")
	content.WriteString("  ],\n\n")

	content.WriteString("  themeConfig: {\n")
	content.WriteString("    navbar: {\n")
	content.WriteString(fmt.Sprintf("      title: '%s',\n", structure.Title))
	content.WriteString("      logo: {\n")
	content.WriteString("        alt: 'Logo',\n")
	content.WriteString("        src: 'img/logo.svg',\n")
	content.WriteString("      },\n")
	content.WriteString("    },\n")
	content.WriteString("    footer: {\n")
	content.WriteString("      style: 'dark',\n")
	content.WriteString("      copyright: `Generated by DeepWiki CLI on ${new Date().getFullYear()}`,\n")
	content.WriteString("    },\n")
	content.WriteString("    prism: {\n")
	content.WriteString("      theme: prismThemes.github,\n")
	content.WriteString("      darkTheme: prismThemes.dracula,\n")
	content.WriteString("    },\n")
	content.WriteString("  } satisfies Preset.ThemeConfig,\n")
	content.WriteString("};\n\n")

	content.WriteString("export default config;\n")

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generatePackageJSON generates the package.json file for v3.8.x
func (d3g *Docusaurus3Generator) generatePackageJSON(
	structure *generator.WikiStructure,
	options OutputOptions,
	filePath string,
) error {
	packageData := map[string]interface{}{
		"name":        strings.ToLower(strings.ReplaceAll(structure.Title, " ", "-")),
		"version":     "0.0.0",
		"description": structure.Description,
		"private":     true,
		"scripts": map[string]string{
			"docusaurus":         "docusaurus",
			"start":              "docusaurus start",
			"build":              "docusaurus build",
			"swizzle":            "docusaurus swizzle",
			"deploy":             "docusaurus deploy",
			"clear":              "docusaurus clear",
			"serve":              "docusaurus serve",
			"write-translations": "docusaurus write-translations",
			"write-heading-ids":  "docusaurus write-heading-ids",
			"typecheck":          "tsc",
		},
		"dependencies": map[string]string{
			"@docusaurus/core":           "3.8.1",
			"@docusaurus/preset-classic": "3.8.1",
			"@docusaurus/faster":         "3.8.1",
			"@docusaurus/theme-mermaid":  "3.8.1",
			"@mdx-js/react":              "^3.0.0",
			"clsx":                       "^2.0.0",
			"prism-react-renderer":       "^2.3.0",
			"react":                      "^18.0.0",
			"react-dom":                  "^18.0.0",
		},
		"devDependencies": map[string]string{
			"@docusaurus/module-type-aliases": "3.8.1",
			"@docusaurus/tsconfig":            "3.8.1",
			"@docusaurus/types":               "3.8.1",
			"typescript":                      "~5.6.0",
		},
		"browserslist": map[string][]string{
			"production": {
				">0.5%",
				"not dead",
				"not op_mini all",
			},
			"development": {
				"last 3 chrome version",
				"last 3 firefox version",
				"last 5 safari version",
			},
		},
		"engines": map[string]string{
			"node": ">=18.0",
		},
	}

	return d3g.writeJSONFile(packageData, filePath)
}

// generateTSConfig generates the tsconfig.json file for v3.8.x
func (d3g *Docusaurus3Generator) generateTSConfig(filePath string) error {
	tsconfigData := map[string]interface{}{
		"extends": "@docusaurus/tsconfig",
		"compilerOptions": map[string]interface{}{
			"baseUrl": ".",
		},
	}

	return d3g.writeJSONFile(tsconfigData, filePath)
}

func (d3g *Docusaurus3Generator) writeJSONFile(data interface{}, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0o644)
}

func (d3g *Docusaurus3Generator) sanitizeFileName(name string) string {
	// Basic sanitization
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "/", "-")
	result = strings.ReplaceAll(result, "\\", "-")
	result = strings.ReplaceAll(result, ":", "-")
	result = strings.ReplaceAll(result, "*", "-")
	result = strings.ReplaceAll(result, "?", "-")
	result = strings.ReplaceAll(result, "\"", "-")
	result = strings.ReplaceAll(result, "<", "-")
	result = strings.ReplaceAll(result, ">", "-")
	result = strings.ReplaceAll(result, "|", "-")

	// Remove multiple consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Trim dashes from start and end
	result = strings.Trim(result, "-")

	// Ensure not empty
	if result == "" {
		result = "untitled"
	}

	return result
}
