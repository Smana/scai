package analyzer

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"

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

	// Clone Git repository
	repoDir := filepath.Join(a.workDir, "repo")

	if a.verbose {
		println("Cloning repository:", repoURL)
	}

	commitSHA, err := CloneRepository(repoURL, repoDir)
	if err != nil {
		return nil, err
	}

	// Analyze the cloned repository
	return a.analyzeDirectory(repoDir, repoURL, commitSHA)
}

// analyzeDirectory analyzes a directory containing application code
func (a *Analyzer) analyzeDirectory(repoPath, repoURL, commitSHA string) (*types.Analysis, error) {
	analysis := &types.Analysis{
		RepoURL:   repoURL,
		RepoPath:  repoPath,
		CommitSHA: commitSHA,
		Verbose:   a.verbose,
	}

	// Detect framework and app directory
	framework, appDir, err := a.detectFramework(repoPath)
	if err != nil {
		return nil, err
	}
	analysis.Framework = framework
	analysis.AppDir = appDir

	// Detect language
	language := a.detectLanguage(repoPath)
	analysis.Language = language

	// Detect package manager
	packageManager := a.detectPackageManager(repoPath, language)
	analysis.PackageManager = packageManager

	// Extract dependencies
	deps, err := a.extractDependencies(repoPath, language)
	if err != nil {
		return nil, err
	}
	analysis.Dependencies = deps

	// Detect start command (use app directory and package manager for accurate detection)
	startCmd := a.detectStartCommand(repoPath, framework, appDir, packageManager)
	analysis.StartCommand = startCmd

	// Detect port (scan actual code files)
	port := a.detectPort(repoPath, framework, appDir)
	analysis.Port = port

	// Extract environment variables
	envVars := a.extractEnvVars(repoPath)
	analysis.EnvVars = envVars

	// Check for special files
	analysis.HasDockerfile = fileExists(filepath.Join(repoPath, "Dockerfile"))
	analysis.HasDockerCompose = fileExists(filepath.Join(repoPath, "docker-compose.yml")) ||
		fileExists(filepath.Join(repoPath, "docker-compose.yaml"))

	return analysis, nil
}

// detectFramework detects the application framework and returns the framework name and app directory
func (a *Analyzer) detectFramework(repoPath string) (string, string, error) {
	// Check for Python frameworks (multiple package managers)
	// Priority: Poetry > uv > requirements.txt > Pipfile

	// Poetry projects (pyproject.toml + poetry.lock)
	if pyprojectPath, foundPyproject := findFileRecursive(repoPath, "pyproject.toml"); foundPyproject {
		appDir := filepath.Dir(pyprojectPath)
		relAppDir, _ := filepath.Rel(repoPath, appDir)

		// Check if it's a Poetry project (has poetry.lock)
		poetryLockPath := filepath.Join(appDir, "poetry.lock")
		if fileExists(poetryLockPath) {
			if _, djangoFound := findFileRecursive(repoPath, "manage.py"); djangoFound {
				return "django", relAppDir, nil
			}
			return "flask", relAppDir, nil
		}

		// Check if it's a uv project (has uv.lock)
		uvLockPath := filepath.Join(appDir, "uv.lock")
		if fileExists(uvLockPath) {
			if _, djangoFound := findFileRecursive(repoPath, "manage.py"); djangoFound {
				return "django", relAppDir, nil
			}
			return "flask", relAppDir, nil
		}
	}

	// Traditional requirements.txt
	if reqPath, found := findFileRecursive(repoPath, "requirements.txt"); found {
		appDir := filepath.Dir(reqPath)
		// Make appDir relative to repoPath
		relAppDir, _ := filepath.Rel(repoPath, appDir)

		// Python framework detection
		if _, djangoFound := findFileRecursive(repoPath, "manage.py"); djangoFound {
			return "django", relAppDir, nil
		}
		// Check requirements.txt content for framework hints
		return "flask", relAppDir, nil // Default Python framework
	}

	// Pipfile (Pipenv)
	if pipfilePath, found := findFileRecursive(repoPath, "Pipfile"); found {
		appDir := filepath.Dir(pipfilePath)
		relAppDir, _ := filepath.Rel(repoPath, appDir)

		if _, djangoFound := findFileRecursive(repoPath, "manage.py"); djangoFound {
			return "django", relAppDir, nil
		}
		return "flask", relAppDir, nil
	}

	if pkgPath, found := findFileRecursive(repoPath, "package.json"); found {
		appDir := filepath.Dir(pkgPath)
		relAppDir, _ := filepath.Rel(repoPath, appDir)
		// JavaScript/TypeScript framework detection
		// TODO: Parse package.json to detect Express, Next.js, etc.
		return "express", relAppDir, nil
	}

	if goModPath, found := findFileRecursive(repoPath, "go.mod"); found {
		appDir := filepath.Dir(goModPath)
		relAppDir, _ := filepath.Rel(repoPath, appDir)
		return "go", relAppDir, nil
	}

	if gemfilePath, found := findFileRecursive(repoPath, "Gemfile"); found {
		appDir := filepath.Dir(gemfilePath)
		relAppDir, _ := filepath.Rel(repoPath, appDir)
		return "rails", relAppDir, nil
	}

	return "unknown", ".", nil
}

// detectPackageManager detects the package manager being used
func (a *Analyzer) detectPackageManager(repoPath, language string) string {
	switch language {
	case "python":
		// Check for Poetry (pyproject.toml + poetry.lock)
		if pyprojectPath, found := findFileRecursive(repoPath, "pyproject.toml"); found {
			appDir := filepath.Dir(pyprojectPath)
			if fileExists(filepath.Join(appDir, "poetry.lock")) {
				return "poetry"
			}
			// Check for uv (pyproject.toml + uv.lock)
			if fileExists(filepath.Join(appDir, "uv.lock")) {
				return "uv"
			}
		}
		// Check for Pipenv (Pipfile)
		if _, found := findFileRecursive(repoPath, "Pipfile"); found {
			return "pipenv"
		}
		// Default to pip (requirements.txt)
		if _, found := findFileRecursive(repoPath, "requirements.txt"); found {
			return "pip"
		}
		return "pip" // Default for Python

	case "javascript":
		// Check for yarn.lock
		if _, found := findFileRecursive(repoPath, "yarn.lock"); found {
			return "yarn"
		}
		// Check for pnpm-lock.yaml
		if _, found := findFileRecursive(repoPath, "pnpm-lock.yaml"); found {
			return "pnpm"
		}
		// Default to npm
		if _, found := findFileRecursive(repoPath, "package.json"); found {
			return "npm"
		}
		return "npm"

	case "go":
		return "go"

	case "ruby":
		return "bundler"

	default:
		return "unknown"
	}
}

// detectLanguage detects the primary programming language
func (a *Analyzer) detectLanguage(repoPath string) string {
	// Search recursively for language indicator files
	// Python: Check multiple package managers
	if _, reqFound := findFileRecursive(repoPath, "requirements.txt"); reqFound {
		return "python"
	}
	if _, setupFound := findFileRecursive(repoPath, "setup.py"); setupFound {
		return "python"
	}
	if _, pipFound := findFileRecursive(repoPath, "Pipfile"); pipFound {
		return "python"
	}
	if pyprojectPath, found := findFileRecursive(repoPath, "pyproject.toml"); found {
		// Check if it's a Python project (has poetry.lock or uv.lock)
		appDir := filepath.Dir(pyprojectPath)
		if fileExists(filepath.Join(appDir, "poetry.lock")) || fileExists(filepath.Join(appDir, "uv.lock")) {
			return "python"
		}
	}

	if _, found := findFileRecursive(repoPath, "package.json"); found {
		return "javascript"
	}

	if _, found := findFileRecursive(repoPath, "go.mod"); found {
		return "go"
	}

	if _, found := findFileRecursive(repoPath, "Gemfile"); found {
		return "ruby"
	}

	if _, pomFound := findFileRecursive(repoPath, "pom.xml"); pomFound {
		return "java"
	}
	if _, gradleFound := findFileRecursive(repoPath, "build.gradle"); gradleFound {
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

// detectStartCommand detects the application start command (without cd, as that's handled by the generator)
func (a *Analyzer) detectStartCommand(repoPath, framework, appDir, packageManager string) string {
	switch framework {
	case "flask":
		// Determine the Python entry point
		entryPoint := "app.py"
		if fileExists(filepath.Join(repoPath, appDir, "app.py")) {
			entryPoint = "app.py"
		} else if fileExists(filepath.Join(repoPath, appDir, "main.py")) {
			entryPoint = "main.py"
		}

		// Use package manager-specific command
		switch packageManager {
		case "poetry":
			return "poetry run python " + entryPoint
		case "uv":
			return "uv run " + entryPoint
		case "pipenv":
			return "pipenv run python " + entryPoint
		default: // pip
			return "python3 " + entryPoint
		}

	case "django":
		// Use package manager-specific command for Django
		switch packageManager {
		case "poetry":
			return "poetry run python manage.py runserver 0.0.0.0:8000"
		case "uv":
			return "uv run manage.py runserver 0.0.0.0:8000"
		case "pipenv":
			return "pipenv run python manage.py runserver 0.0.0.0:8000"
		default: // pip
			return "python3 manage.py runserver 0.0.0.0:8000"
		}

	case "express":
		// Use JavaScript package manager-specific command
		switch packageManager {
		case "yarn":
			return "yarn start"
		case "pnpm":
			return "pnpm start"
		default: // npm
			return "npm start"
		}

	case "go":
		return "go run ."

	default:
		return "unknown"
	}
}

// detectPort detects the application port by scanning code files
func (a *Analyzer) detectPort(repoPath, framework, appDir string) int {
	// Try to scan code files for port numbers
	appPath := filepath.Join(repoPath, appDir)

	switch framework {
	case "flask", "django":
		// Scan Python files for port=XXXX
		if port := a.scanPythonFilesForPort(appPath); port > 0 {
			return port
		}
		// Fallback to framework defaults
		if framework == "flask" {
			return 5000
		}
		return 8000

	case "express":
		// TODO: Scan JavaScript files for port
		return 3000

	case "rails":
		return 3000

	case "go":
		// TODO: Scan Go files for port
		return 8080

	default:
		return 8080
	}
}

// scanPythonFilesForPort scans Python files for port configuration
func (a *Analyzer) scanPythonFilesForPort(appPath string) int {
	// Common Python files to check
	filesToCheck := []string{"app.py", "main.py", "wsgi.py", "server.py"}

	for _, filename := range filesToCheck {
		pyFilePath := filepath.Join(appPath, filename)
		if !fileExists(pyFilePath) {
			continue
		}

		content, err := os.ReadFile(pyFilePath)
		if err != nil {
			continue
		}

		// Look for port= pattern (e.g., port=5000, port = 5000)
		re := regexp.MustCompile(`port\s*=\s*(\d+)`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			if port, err := strconv.Atoi(matches[1]); err == nil {
				return port
			}
		}
	}

	return 0 // Not found
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

// findFileRecursive searches for a file recursively in a directory (max depth 3)
func findFileRecursive(dir, filename string) (string, bool) {
	return findFileRecursiveWithDepth(dir, filename, 0, 3)
}

// findFileRecursiveWithDepth searches for a file recursively with depth limit
func findFileRecursiveWithDepth(dir, filename string, currentDepth, maxDepth int) (string, bool) {
	if currentDepth > maxDepth {
		return "", false
	}

	// Check current directory
	targetPath := filepath.Join(dir, filename)
	if fileExists(targetPath) {
		return targetPath, true
	}

	// Search subdirectories
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" && entry.Name() != "node_modules" && entry.Name() != "venv" {
			subdirPath := filepath.Join(dir, entry.Name())
			if found, ok := findFileRecursiveWithDepth(subdirPath, filename, currentDepth+1, maxDepth); ok {
				return found, true
			}
		}
	}

	return "", false
}
