package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deepwiki",
	Short: "AI-powered documentation generator for local directories",
	Long: `DeepWiki is a Go-based tool that analyzes your local codebase and generates
comprehensive documentation using OpenAI's GPT models.

It scans your project directory, processes files intelligently, and creates
structured wiki-style documentation with:
- Architectural overviews and diagrams
- Code analysis and explanations  
- API documentation
- Setup and deployment guides

Example usage:
  deepwiki generate
  deepwiki generate --path /path/to/project
  deepwiki generate --output-dir ./docs`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.deepwiki.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, git commit, and build date of deepwiki.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("deepwiki version %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Build date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
