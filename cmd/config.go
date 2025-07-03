package cmd

import (
	"fmt"
	"os"

	"github.com/kuderr/deepwiki/internal/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Commands for managing deepwiki-cli configuration files.`,
}

// configInitCmd represents the config init command
var configInitCmd = &cobra.Command{
	Use:   "init [filename]",
	Short: "Initialize a new configuration file",
	Long: `Create a new configuration file with default values.

This will create a deepwiki.yaml file with all available configuration options
and their default values. You can then edit this file to customize the behavior.

Examples:
  deepwiki-cli config init
  deepwiki-cli config init myconfig.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigInit,
}

// configValidateCmd represents the config validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate [filename]",
	Short: "Validate a configuration file",
	Long: `Validate that a configuration file has correct syntax and values.

Examples:
  deepwiki-cli config validate
  deepwiki-cli config validate myconfig.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigValidate,
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	filename := "deepwiki.yaml"
	if len(args) > 0 {
		filename = args[0]
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("Configuration file '%s' already exists.\n", filename)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Configuration file creation cancelled.")
			return nil
		}
	}

	// Generate template
	template := config.GenerateTemplate()

	// Write to file
	if err := os.WriteFile(filename, []byte(template), 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("✅ Configuration file created: %s\n", filename)
	fmt.Println("You can now edit this file to customize your settings.")
	fmt.Println("Remember to set your OPENAI_API_KEY environment variable or add it to the config file.")

	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	filename := ""
	if len(args) > 0 {
		filename = args[0]
	}

	// Load and validate configuration
	cfg, err := config.LoadConfig(filename)
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// If we get here, the configuration is valid
	fmt.Println("✅ Configuration is valid!")

	// Show some key settings
	fmt.Printf("LLM Provider: %s\n", cfg.Providers.LLM.Provider)
	fmt.Printf("LLM Model: %s\n", cfg.Providers.LLM.Model)
	fmt.Printf("Embedding Provider: %s\n", cfg.Providers.Embedding.Provider)
	fmt.Printf("Embedding Model: %s\n", cfg.Providers.Embedding.Model)
	fmt.Printf("Output Format: %s\n", cfg.Output.Format)
	fmt.Printf("Language: %s\n", cfg.Output.Language)
	fmt.Printf("Chunk Size: %d\n", cfg.Processing.ChunkSize)

	// Check for API key
	if cfg.Providers.LLM.APIKey == "" {
		fmt.Println("⚠️  Warning: No OpenAI API key configured")
		fmt.Println("   Set OPENAI_API_KEY environment variable or add it to the config file")
	} else {
		fmt.Println("✅ OpenAI API key is configured")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
}
