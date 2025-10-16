package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	workDir string
	verbose bool
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

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.scia.yaml)")
	rootCmd.PersistentFlags().StringVar(&workDir, "work-dir", "/tmp/scia", "working directory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to Viper
	viper.BindPFlag("workdir", rootCmd.PersistentFlags().Lookup("work-dir"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
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
	viper.SetEnvPrefix("SCIA")
	viper.AutomaticEnv()

	// Read config file if exists
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	// Set defaults
	viper.SetDefault("ollama.url", "http://localhost:11434")
	viper.SetDefault("ollama.model", "qwen2.5-coder:7b")
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("terraform.bin", "tofu")
}
