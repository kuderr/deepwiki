package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kuderr/deepwiki/pkg/generator"
)

// Docusaurus2Generator generates Docusaurus v2.4.x-formatted output files
type Docusaurus2Generator struct{}

// NewDocusaurus2Generator creates a new Docusaurus v2 generator
func NewDocusaurus2Generator() *Docusaurus2Generator {
	return &Docusaurus2Generator{}
}

// FormatType returns the format type this generator handles
func (d2g *Docusaurus2Generator) FormatType() OutputFormat {
	return FormatDocusaurus2
}

// Description returns a human-readable description of the format
func (d2g *Docusaurus2Generator) Description() string {
	return "Docusaurus v2.4.x site with JavaScript configuration and React v17"
}

// Generate creates Docusaurus v2.4.x output files
func (d2g *Docusaurus2Generator) Generate(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	options OutputOptions,
) (*OutputResult, error) {
	startTime := time.Now()

	// Create Docusaurus directory structure
	if err := d2g.organizeDocusaurusFiles(options.Directory); err != nil {
		return nil, fmt.Errorf("failed to create Docusaurus directory structure: %w", err)
	}

	var filesGenerated []string
	var totalSize int64
	var errors []error

	// Generate intro page (Docusaurus home)
	introPath := filepath.Join(options.Directory, "docs", "intro.md")
	if err := d2g.generateIndex(structure, pages, introPath, options); err != nil {
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
		fileName := d2g.sanitizeFileName(page.Title) + ".md"
		pagePath := filepath.Join(docsDir, fileName)

		if err := d2g.generatePage(page, pagePath, structure, options); err != nil {
			errors = append(errors, fmt.Errorf("failed to generate page %s: %w", pageID, err))
			continue
		}

		if stat, err := os.Stat(pagePath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, pagePath)
	}

	// Generate sidebars.js configuration
	sidebarPath := filepath.Join(options.Directory, "sidebars.js")
	if err := d2g.generateSidebar(structure, pages, sidebarPath, options); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate sidebar: %w", err))
	} else {
		if stat, err := os.Stat(sidebarPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, sidebarPath)
	}

	// Generate basic docusaurus.config.js
	configPath := filepath.Join(options.Directory, "docusaurus.config.js")
	if err := d2g.generateConfig(structure, options, configPath); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate Docusaurus config: %w", err))
	} else {
		if stat, err := os.Stat(configPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, configPath)
	}

	// Generate package.json for Docusaurus dependencies
	packagePath := filepath.Join(options.Directory, "package.json")
	if err := d2g.generatePackageJSON(structure, options, packagePath); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate package.json: %w", err))
	} else {
		if stat, err := os.Stat(packagePath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, packagePath)
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
func (d2g *Docusaurus2Generator) organizeDocusaurusFiles(outputDir string) error {
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

// generateIndex generates the intro.md file for Docusaurus v2.4.x
func (d2g *Docusaurus2Generator) generateIndex(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	filePath string,
	options OutputOptions,
) error {
	var content strings.Builder

	// Write Docusaurus frontmatter
	content.WriteString("---\n")
	content.WriteString("sidebar_position: 1\n")
	content.WriteString("slug: /\n")
	content.WriteString(fmt.Sprintf("title: %s\n", structure.Title))
	content.WriteString(fmt.Sprintf("description: %s\n", structure.Description))
	content.WriteString("---\n\n")

	// Write header
	content.WriteString(fmt.Sprintf("# %s\n\n", structure.Title))
	content.WriteString(fmt.Sprintf("%s\n\n", structure.Description))

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
			fileName := d2g.sanitizeFileName(page.Title)
			content.WriteString(fmt.Sprintf("- [%s](./%s) - %s\n", page.Title, fileName, page.Description))
		}
		content.WriteString("\n")
	}

	// Write medium importance pages
	if len(importanceGroups["medium"]) > 0 {
		content.WriteString("### 📋 Core Documentation\n\n")
		for _, page := range importanceGroups["medium"] {
			fileName := d2g.sanitizeFileName(page.Title)
			content.WriteString(fmt.Sprintf("- [%s](./%s) - %s\n", page.Title, fileName, page.Description))
		}
		content.WriteString("\n")
	}

	// Write footer
	content.WriteString("---\n\n")
	content.WriteString("*Generated by [DeepWiki CLI](https://github.com/kuderr/deepwiki)*\n")

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generatePage generates a Docusaurus v2.4.x-formatted page
func (d2g *Docusaurus2Generator) generatePage(
	page *generator.WikiPage,
	filePath string,
	structure *generator.WikiStructure,
	options OutputOptions,
) error {
	var content strings.Builder

	// Generate slug from filename
	fileName := strings.TrimSuffix(filepath.Base(filePath), ".md")

	// Write Docusaurus frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("id: %s\n", page.ID))
	content.WriteString(fmt.Sprintf("title: %s\n", page.Title))
	content.WriteString(fmt.Sprintf("slug: /%s\n", fileName))
	if page.Description != "" {
		content.WriteString(fmt.Sprintf("description: %s\n", page.Description))
	}
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

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generateSidebar generates the sidebars.js configuration for v2.4.x
func (d2g *Docusaurus2Generator) generateSidebar(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	filePath string,
	options OutputOptions,
) error {
	var content strings.Builder

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

	content.WriteString("// @ts-check\n\n")
	content.WriteString("/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */\n")
	content.WriteString("const sidebars = {\n")
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
		content.WriteString("      items: [\n")
		for _, page := range importanceGroups["low"] {
			content.WriteString(fmt.Sprintf("        '%s',\n", page.ID))
		}
		content.WriteString("      ],\n")
		content.WriteString("    },\n")
	}

	content.WriteString("  ],\n")
	content.WriteString("};\n\n")
	content.WriteString("module.exports = sidebars;\n")

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generateConfig generates the docusaurus.config.js file for v2.4.x
func (d2g *Docusaurus2Generator) generateConfig(
	structure *generator.WikiStructure,
	options OutputOptions,
	filePath string,
) error {
	var content strings.Builder

	content.WriteString("// @ts-check\n")
	content.WriteString("// Note: type annotations allow type checking and IDEs autocompletion\n\n")

	content.WriteString("const lightCodeTheme = require('prism-react-renderer/themes/github');\n")
	content.WriteString("const darkCodeTheme = require('prism-react-renderer/themes/dracula');\n\n")

	content.WriteString("/** @type {import('@docusaurus/types').Config} */\n")
	content.WriteString("const config = {\n")
	content.WriteString(fmt.Sprintf("  title: '%s',\n", structure.Title))
	content.WriteString(fmt.Sprintf("  tagline: '%s',\n", structure.Description))

	// mermaid
	content.WriteString("  markdown: {\n")
	content.WriteString("    mermaid: true,\n")
	content.WriteString("  },\n")
	content.WriteString("  themes: ['@docusaurus/theme-mermaid'],\n")

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

	content.WriteString("  // Even if you don't use internalization, you can use this field to set useful\n")
	content.WriteString("  // metadata like html lang. For example, if your site is Chinese, you may want\n")
	content.WriteString("  // to replace \"en\" with \"zh-Hans\".\n")
	content.WriteString("  i18n: {\n")
	content.WriteString(fmt.Sprintf("    defaultLocale: '%s',\n", options.Language))
	content.WriteString(fmt.Sprintf("    locales: ['%s'],\n", options.Language))
	content.WriteString("  },\n\n")

	content.WriteString("  presets: [\n")
	content.WriteString("    [\n")
	content.WriteString("      'classic',\n")
	content.WriteString("      /** @type {import('@docusaurus/preset-classic').Options} */\n")
	content.WriteString("      ({\n")
	content.WriteString("        docs: {\n")
	content.WriteString("          sidebarPath: require.resolve('./sidebars.js'),\n")
	content.WriteString("          routeBasePath: '/',\n")
	content.WriteString("        },\n")
	content.WriteString("        blog: false,\n")
	content.WriteString("      }),\n")
	content.WriteString("    ],\n")
	content.WriteString("  ],\n\n")

	content.WriteString("  themeConfig:\n")
	content.WriteString("    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */\n")
	content.WriteString("    ({\n")
	content.WriteString("      navbar: {\n")
	content.WriteString(fmt.Sprintf("        title: '%s',\n", structure.Title))
	content.WriteString("        logo: {\n")
	content.WriteString("          alt: 'Logo',\n")
	content.WriteString("          src: 'img/logo.svg',\n")
	content.WriteString("        },\n")
	content.WriteString("      },\n")
	content.WriteString("      footer: {\n")
	content.WriteString("        style: 'dark',\n")
	content.WriteString("        copyright: `Generated by DeepWiki CLI on ${new Date().getFullYear()}`,\n")
	content.WriteString("      },\n")
	content.WriteString("      prism: {\n")
	content.WriteString("        theme: lightCodeTheme,\n")
	content.WriteString("        darkTheme: darkCodeTheme,\n")
	content.WriteString("      },\n")
	content.WriteString("    }),\n")
	content.WriteString("};\n\n")

	content.WriteString("module.exports = config;\n")

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generatePackageJSON generates the package.json file for v2.4.x
func (d2g *Docusaurus2Generator) generatePackageJSON(
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
		},
		"dependencies": map[string]string{
			"@docusaurus/core":           "2.4.3",
			"@docusaurus/preset-classic": "2.4.3",
			"@docusaurus/theme-mermaid":  "2.4.3",
			"@mdx-js/react":              "^1.6.22",
			"clsx":                       "^1.2.1",
			"prism-react-renderer":       "^1.3.5",
			"react":                      "^17.0.2",
			"react-dom":                  "^17.0.2",
		},
		"devDependencies": map[string]string{
			"@docusaurus/module-type-aliases": "2.4.3",
		},
		"browserslist": map[string][]string{
			"production": {
				">0.5%",
				"not dead",
				"not op_mini all",
			},
			"development": {
				"last 1 chrome version",
				"last 1 firefox version",
				"last 1 safari version",
			},
		},
		"engines": map[string]string{
			"node": ">=16.14",
		},
	}

	return d2g.writeJSONFile(packageData, filePath)
}

func (d2g *Docusaurus2Generator) writeJSONFile(data interface{}, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0o644)
}

func (d2g *Docusaurus2Generator) sanitizeFileName(name string) string {
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
