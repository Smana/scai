package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CloneRepository clones a Git repository to the specified destination and returns the commit SHA
func CloneRepository(repoURL, destDir string) (string, error) {
	// Validate URL
	if !strings.HasPrefix(repoURL, "https://") && !strings.HasPrefix(repoURL, "http://") {
		return "", fmt.Errorf("invalid repository URL: must start with https:// or http://")
	}

	// Check if destination already exists
	if _, err := os.Stat(destDir); err == nil {
		// Directory exists, remove it to allow fresh clone
		if err := os.RemoveAll(destDir); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Clone options
	cloneOpts := &git.CloneOptions{
		URL:      repoURL,
		Progress: nil, // Can add progress output here if needed
		Depth:    1,   // Shallow clone - we only need the latest commit
	}

	// Clone the repository
	repo, err := git.PlainClone(destDir, false, cloneOpts)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get commit SHA
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitSHA := ref.Hash().String()

	return commitSHA, nil
}

// CloneRepositoryWithBranch clones a specific branch of a Git repository
func CloneRepositoryWithBranch(repoURL, branch, destDir string) error {
	// Check if destination already exists
	if _, err := os.Stat(destDir); err == nil {
		// Directory exists, remove it to allow fresh clone
		if err := os.RemoveAll(destDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Clone options with branch specification
	cloneOpts := &git.CloneOptions{
		URL:           repoURL,
		Progress:      nil,
		Depth:         1,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	}

	// Clone the repository
	_, err := git.PlainClone(destDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository branch '%s': %w", branch, err)
	}

	return nil
}

// GetRepositoryInfo extracts repository information
func GetRepositoryInfo(repoPath string) (map[string]string, error) {
	info := make(map[string]string)

	// Open repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	info["commit"] = ref.Hash().String()[:8] // Short commit hash
	info["branch"] = ref.Name().Short()

	// Get remote URL
	remote, err := repo.Remote("origin")
	if err == nil && len(remote.Config().URLs) > 0 {
		info["remote"] = remote.Config().URLs[0]
	}

	return info, nil
}

// IsGitRepository checks if a directory is a Git repository
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}
