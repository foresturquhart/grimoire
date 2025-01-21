package grimoire

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// FindGitRoot traverses upward from a given directory to locate the top-level .git folder.
// If no Git repository is found, an error is returned.
func FindGitRoot(startDir string) (string, error) {
	current := startDir

	for {
		// Check if .git exists in the current directory
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return current, nil // Found the Git root
		}

		// Move to the parent directory
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no Git repository found starting from %s", startDir) // Reached root without finding .git
		}

		current = parent
	}
}

// ChangeCounter defines a function type for getting change counts.
type ChangeCounter func(repoDir string) (map[string]int, error)

// SortByChangeFrequency sorts a list of files based on their commit frequency in descending order.
func SortByChangeFrequency(repoDir string, files []string, counter ChangeCounter) ([]string, error) {
	commitCounts, err := counter(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit counts: %w", err)
	}

	// Sort files based on commit frequency
	sort.Slice(files, func(i, j int) bool {
		return commitCounts[files[i]] < commitCounts[files[j]] // Ascending order by commit count
	})

	return files, nil
}

// DefaultChangeCounter uses the git command to count changes
func DefaultChangeCounter(repoDir string) (map[string]int, error) {
	return getGitCommitCounts(repoDir)
}

// getGitCommitCounts collects the number of commits that modified each file in the repository.
func getGitCommitCounts(repoDir string) (map[string]int, error) {
	cmd := exec.Command("git", "-C", repoDir, "log", "--name-only", "-l 99999", "--pretty=format:", "--no-merges", "--relative")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	// Execute the Git log command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to execute git log: %w", err)
	}

	commitCounts := make(map[string]int)

	// Parse the Git log output
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commitCounts[line]++
		}
	}

	// Handle potential errors while scanning the output
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading git log output: %w", err)
	}

	return commitCounts, nil
}

// IsGitAvailable checks if git is available on the system.
func IsGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
