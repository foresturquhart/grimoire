package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// GitExecutor defines the interface for running git-related commands.
type GitExecutor interface {
	// ListFileChanges returns a ReadCloser that streams file paths that were changed in commits.
	// The caller is responsible for closing the returned stream.
	ListFileChanges(repoDir string) (io.ReadCloser, error)

	// IsAvailable indicates whether git is installed and can be found in PATH.
	IsAvailable() bool
}

// DefaultGitExecutor is the default, concrete implementation of GitExecutor.
type DefaultGitExecutor struct{}

// NewDefaultGitExecutor returns a new DefaultGitExecutor instance.
func NewDefaultGitExecutor() *DefaultGitExecutor {
	return &DefaultGitExecutor{}
}

// cmdReadCloser wraps the command’s stdout so that closing it will also Wait() on the command.
// This avoids leaving any zombie processes.
type cmdReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

// Close closes the underlying pipe and then waits for the command to finish.
func (crc *cmdReadCloser) Close() error {
	closeErr := crc.ReadCloser.Close()
	waitErr := crc.cmd.Wait()

	// If both errors are nil, return nil
	if closeErr == nil && waitErr == nil {
		return nil
	}

	// If only one error is non-nil, return that error
	if closeErr == nil {
		return waitErr
	}
	if waitErr == nil {
		return closeErr
	}

	// If both errors are non-nil, combine them
	return fmt.Errorf("multiple errors on close: %w (close); %v (wait)", closeErr, waitErr)
}

// ListFileChanges runs the `git log --name-only ...` command and returns a stream of file paths.
// Callers must close the returned ReadCloser to free resources and reap the spawned process.
func (e *DefaultGitExecutor) ListFileChanges(repoDir string) (io.ReadCloser, error) {
	cmd := exec.Command(
		"git",
		"-C", repoDir,
		"log",
		"--name-only",
		"-n", "99999",
		"--pretty=format:",
		"--no-merges",
		"--relative",
	)
	return e.executeWithReader(cmd, os.Stderr)
}

// IsAvailable returns true if the `git` executable is found in the system's PATH.
func (e *DefaultGitExecutor) IsAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// executeWithReader starts the given command and returns a ReadCloser for stdout.
// Once the caller closes it, the child process is reaped via cmd.Wait().
func (e *DefaultGitExecutor) executeWithReader(cmd *exec.Cmd, stderr io.Writer) (io.ReadCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Wrap stdout in cmdReadCloser so that calling Close() will also wait on the command.
	return &cmdReadCloser{
		ReadCloser: stdout,
		cmd:        cmd,
	}, nil
}

// Git is a higher-level interface to a GitExecutor, providing convenience methods
// for repository discovery and commit analysis.
type Git struct {
	executor GitExecutor
}

// NewGit returns a new Git instance that delegates to the provided GitExecutor.
func NewGit(executor GitExecutor) *Git {
	return &Git{executor: executor}
}

// IsAvailable indicates whether the underlying GitExecutor is capable of running git.
func (g *Git) IsAvailable() bool {
	return g.executor.IsAvailable()
}

// FindRepositoryRoot walks up the directory tree from startDir until it finds
// a `.git` directory. It returns the path to that directory, or an error if none is found.
func (g *Git) FindRepositoryRoot(startDir string) (string, error) {
	current := startDir
	for {
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			// Found the Git root
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root of filesystem without finding .git
			return "", fmt.Errorf("no repository found starting from %s", startDir)
		}
		current = parent
	}
}

// GetCommitCounts returns a map of file paths to the number of commits in which each file appears.
func (g *Git) GetCommitCounts(repoDir string) (map[string]int, error) {
	output, err := g.executor.ListFileChanges(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list file changes: %w", err)
	}
	defer output.Close()

	commitCounts := make(map[string]int)

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commitCounts[line]++
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("error reading git log output: %w", scanErr)
	}

	return commitCounts, nil
}

// CommitCounter defines a function type for counting the number of commits per file in a repository.
type CommitCounter func(repoDir string) (map[string]int, error)

// SortFilesByCommitCounts sorts the provided files based on their commit counts in ascending order.
// It uses the provided commitCounter function to retrieve commit counts for each file.
func (g *Git) SortFilesByCommitCounts(repoDir string, filePaths []string, commitCounter CommitCounter) ([]string, error) {
	commitCounts, err := commitCounter(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit counts: %w", err)
	}

	sort.Slice(filePaths, func(i, j int) bool {
		countI := 0
		if c, ok := commitCounts[filePaths[i]]; ok {
			countI = c
		}

		countJ := 0
		if c, ok := commitCounts[filePaths[j]]; ok {
			countJ = c
		}

		if countI != countJ {
			// Sort by ascending commit count
			return countI < countJ
		}

		// If counts are equal, sort alphabetically
		return filePaths[i] < filePaths[j]
	})

	return filePaths, nil
}
