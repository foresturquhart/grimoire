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

// SortByCommitFrequency sorts a list of files based on their commit frequency in descending order.
func SortByCommitFrequency(repoDir string, files []string) ([]string, error) {
	commitCounts, err := getCommitCounts(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit counts: %w", err)
	}

	// Sort files based on commit frequency
	sort.Slice(files, func(i, j int) bool {
		return commitCounts[files[i]] < commitCounts[files[j]] // Ascending order by commit count
	})

	return files, nil
}

// getCommitCounts collects the number of commits that modified each file in the repository.
func getCommitCounts(repoDir string) (map[string]int, error) {
	cmd := exec.Command("git", "-C", repoDir, "log", "--name-only", "--pretty=format:", "--no-merges", "--relative")
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
