package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Smana/scia/internal/store"
)

var (
	cfgFile string
	workDir string
	verbose bool

	// Version information set by main package
	version string
	commit  string
	date    string
	builtBy string

	// Global store instance
	globalStore store.Store
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "scia",
	Short: "Smart Cloud Infrastructure Automation - AI-powered deployment system",
	Long: `SCIA (Smart Cloud Infrastructure Automation) analyzes code repositories,
determines optimal deployment strategies using AI, and automatically provisions
infrastructure using Terraform.

Example:
  scia deploy "Deploy this Flask app on AWS" https://github.com/Arvo-AI/hello_world
  scia deploy "Deploy microservices" /path/to/app.zip`,
}

// SetVersionInfo sets version information from main package
func SetVersionInfo(v, c, d, b string) {
	version = v
	commit = c
	date = d
	builtBy = b
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s by %s)", version, commit, date, builtBy)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initDatabase)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.scia.yaml)")
	rootCmd.PersistentFlags().StringVar(&workDir, "work-dir", "/tmp/scia", "working directory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to Viper
	_ = viper.BindPFlag("workdir", rootCmd.PersistentFlags().Lookup("work-dir"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initDatabase initializes the SQLite database for deployment tracking
func initDatabase() {
	// Get home directory for database storage
	home, err := os.UserHomeDir()
	if err != nil {
		// Fail silently - database is optional
		return
	}

	// Database path: ~/.scia/deployments.db
	sciaDir := filepath.Join(home, ".scia")
	dbPath := filepath.Join(sciaDir, "deployments.db")

	// Create .scia directory if it doesn't exist
	if err := os.MkdirAll(sciaDir, 0o755); err != nil {
		// Fail silently - database is optional
		if verbose {
			fmt.Printf("Warning: failed to create .scia directory: %v\n", err)
		}
		return
	}

	// Create SQLite store
	sqliteStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		// Fail silently - database is optional
		if verbose {
			fmt.Printf("Warning: failed to create database: %v\n", err)
		}
		return
	}

	// Initialize database schema
	ctx := context.Background()
	if err := sqliteStore.Initialize(ctx); err != nil {
		// Fail silently - database is optional
		if verbose {
			fmt.Printf("Warning: failed to initialize database: %v\n", err)
		}
		_ = sqliteStore.Close()
		return
	}

	// Set global store
	globalStore = sqliteStore

	if verbose {
		fmt.Printf("✓ Database initialized: %s\n", dbPath)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".scia")
	}

	// Read environment variables with SCIA_ prefix
	// e.g., SCIA_OLLAMA_URL=http://remote:11434
	viper.SetEnvPrefix("SCIA")
	viper.AutomaticEnv()
	// Replace dots with underscores for env vars (ollama.url -> SCIA_OLLAMA_URL)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file if exists
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	// Set defaults
	// LLM configuration
	viper.SetDefault("llm.provider", "ollama")
	viper.SetDefault("llm.ollama.url", "http://localhost:11434")
	viper.SetDefault("llm.ollama.model", "qwen2.5-coder:7b")
	viper.SetDefault("llm.ollama.use_docker", true) // Prefer Docker by default
	viper.SetDefault("llm.gemini.model", "gemini-2.0-pro-exp")
	viper.SetDefault("llm.openai.model", "gpt-4o")

	// Cloud configuration
	viper.SetDefault("cloud.provider", "aws")
	viper.SetDefault("cloud.default_region", "eu-west-3")
	viper.SetDefault("aws.region", "eu-west-3") // Legacy support

	// Terraform configuration
	viper.SetDefault("terraform.bin", "tofu")
	viper.SetDefault("terraform.backend.type", "s3")
	viper.SetDefault("terraform.backend.s3_key", "terraform.tfstate")
}
