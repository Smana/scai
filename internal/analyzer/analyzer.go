package analyzer

import (
	"os"
	"path/filepath"

	"github.com/Smana/scia/internal/types"
)

// Analyzer handles repository analysis
type Analyzer struct {
	workDir string
	verbose bool
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer(workDir string, verbose bool) *Analyzer {
	return &Analyzer{
		workDir: workDir,
		verbose: verbose,
	}
}

// Analyze performs full repository analysis
func (a *Analyzer) Analyze(repoURL string) (*types.Analysis, error) {
	// Check if it's a zip file
	if IsZipFile(repoURL) {
		return a.AnalyzeFromZip(repoURL)
	}

	// TODO: Implement GitHub repository cloning and analysis
	// For now, return placeholder
	return &types.Analysis{
		RepoURL:      repoURL,
		Framework:    "unknown",
		Language:     "unknown",
		Port:         8080,
		StartCommand: "unknown",
	}, nil
}

// detectFramework detects the application framework
func (a *Analyzer) detectFramework(repoPath string) (string, error) {
	// Check for common framework indicators
	if fileExists(filepath.Join(repoPath, "requirements.txt")) {
		// Python framework detection
		if fileExists(filepath.Join(repoPath, "manage.py")) {
			return "django", nil
		}
		// Check requirements.txt content for framework hints
		return "flask", nil // Default Python framework
	}

	if fileExists(filepath.Join(repoPath, "package.json")) {
		// JavaScript/TypeScript framework detection
		// TODO: Parse package.json to detect Express, Next.js, etc.
		return "express", nil
	}

	if fileExists(filepath.Join(repoPath, "go.mod")) {
		return "go", nil
	}

	if fileExists(filepath.Join(repoPath, "Gemfile")) {
		return "rails", nil
	}

	return "unknown", nil
}

// detectLanguage detects the primary programming language
func (a *Analyzer) detectLanguage(repoPath string) string {
	if fileExists(filepath.Join(repoPath, "requirements.txt")) ||
		fileExists(filepath.Join(repoPath, "setup.py")) ||
		fileExists(filepath.Join(repoPath, "Pipfile")) {
		return "python"
	}

	if fileExists(filepath.Join(repoPath, "package.json")) {
		return "javascript"
	}

	if fileExists(filepath.Join(repoPath, "go.mod")) {
		return "go"
	}

	if fileExists(filepath.Join(repoPath, "Gemfile")) {
		return "ruby"
	}

	if fileExists(filepath.Join(repoPath, "pom.xml")) ||
		fileExists(filepath.Join(repoPath, "build.gradle")) {
		return "java"
	}

	return "unknown"
}

// extractDependencies extracts project dependencies
func (a *Analyzer) extractDependencies(repoPath, language string) ([]string, error) {
	var deps []string

	switch language {
	case "python":
		// TODO: Parse requirements.txt
		deps = []string{"flask"} // Placeholder
	case "javascript":
		// TODO: Parse package.json
		deps = []string{"express"} // Placeholder
	case "go":
		// TODO: Parse go.mod
		deps = []string{} // Placeholder
	}

	return deps, nil
}

// detectStartCommand detects the application start command
func (a *Analyzer) detectStartCommand(repoPath, framework string) string {
	switch framework {
	case "flask":
		return "python app.py"
	case "django":
		return "python manage.py runserver"
	case "express":
		return "npm start"
	case "go":
		return "go run ."
	default:
		return "unknown"
	}
}

// detectPort detects the application port
func (a *Analyzer) detectPort(repoPath, framework string) int {
	// Framework-specific defaults
	switch framework {
	case "flask":
		return 5000
	case "django":
		return 8000
	case "express":
		return 3000
	case "rails":
		return 3000
	case "go":
		return 8080
	default:
		return 8080
	}
}

// extractEnvVars extracts environment variable requirements
func (a *Analyzer) extractEnvVars(repoPath string) map[string]string {
	envVars := make(map[string]string)

	// Check for .env.example
	envExamplePath := filepath.Join(repoPath, ".env.example")
	if fileExists(envExamplePath) {
		// TODO: Parse .env.example file
		envVars["PORT"] = "8080"
		envVars["DATABASE_URL"] = ""
	}

	return envVars
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
