package analyzer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Smana/scia/internal/types"
)

// AnalyzeFromZip analyzes a zip file containing application code
func (a *Analyzer) AnalyzeFromZip(zipPath string) (*types.Analysis, error) {
	// Extract zip file
	repoPath, err := a.extractZip(zipPath)
	if err != nil {
		return nil, fmt.Errorf("zip extraction failed: %w", err)
	}

	// Use regular analysis on extracted directory
	analysis := &types.Analysis{
		RepoURL:  zipPath, // Store zip path as "URL"
		RepoPath: repoPath,
	}

	// Detect framework
	framework, err := a.detectFramework(repoPath)
	if err != nil {
		return nil, err
	}
	analysis.Framework = framework

	// Detect language
	language := a.detectLanguage(repoPath)
	analysis.Language = language

	// Extract dependencies
	deps, err := a.extractDependencies(repoPath, language)
	if err != nil {
		return nil, err
	}
	analysis.Dependencies = deps

	// Detect start command
	startCmd := a.detectStartCommand(repoPath, framework)
	analysis.StartCommand = startCmd

	// Detect port
	port := a.detectPort(repoPath, framework)
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

// extractZip extracts a zip file to the work directory
func (a *Analyzer) extractZip(zipPath string) (string, error) {
	// Create extraction directory
	extractDir := filepath.Join(a.workDir, "repos")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", err
	}

	// Extract zip filename without extension
	zipFileName := filepath.Base(zipPath)
	zipFileName = strings.TrimSuffix(zipFileName, filepath.Ext(zipFileName))

	targetPath := filepath.Join(extractDir, zipFileName)

	// Remove if exists
	if err := os.RemoveAll(targetPath); err != nil {
		return "", err
	}

	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return "", err
	}

	// Open zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Extract all files
	for _, file := range reader.File {
		if err := extractZipFile(file, targetPath); err != nil {
			return "", fmt.Errorf("failed to extract %s: %w", file.Name, err)
		}
	}

	return targetPath, nil
}

// extractZipFile extracts a single file from zip archive
func extractZipFile(file *zip.File, destDir string) error {
	// Build destination path
	destPath := filepath.Join(destDir, file.Name)

	// Check for zip slip vulnerability
	if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", file.Name)
	}

	// Create directory if it's a directory
	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.Mode())
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return nil
}

// IsZipFile checks if a path is a zip file
func IsZipFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".zip")
}

// IsGitHubURL checks if a string is a GitHub URL
func IsGitHubURL(s string) bool {
	s = strings.ToLower(s)
	return strings.HasPrefix(s, "http://github.com") ||
		strings.HasPrefix(s, "https://github.com") ||
		strings.HasPrefix(s, "git@github.com")
}
