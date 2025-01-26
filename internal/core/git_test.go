package core

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepositoryRoot(t *testing.T) {
	git := NewGit(NewDefaultGitExecutor())

	// Test case 1: .git directory in the startDir
	tempDir1, err := os.MkdirTemp("", "git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir1)

	gitDir1 := filepath.Join(tempDir1, ".git")
	if err := os.Mkdir(gitDir1, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	root1, err1 := git.FindRepositoryRoot(tempDir1)
	if err1 != nil {
		t.Fatalf("Test Case 1 Failed: Expected no error, got: %v", err1)
	}
	if root1 != tempDir1 {
		t.Fatalf("Test Case 1 Failed: Expected root to be %s, got: %s", tempDir1, root1)
	}

	// Test case 2: .git directory in a parent directory
	tempDir2, err := os.MkdirTemp("", "git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir2)

	gitDir2 := filepath.Join(tempDir2, ".git")
	if err := os.Mkdir(gitDir2, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}
	subdir2 := filepath.Join(tempDir2, "subdir")
	if err := os.Mkdir(subdir2, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	root2, err2 := git.FindRepositoryRoot(subdir2)
	if err2 != nil {
		t.Errorf("Test Case 2 Failed: Expected no error, got: %v", err2)
	}
	if root2 != tempDir2 {
		t.Errorf("Test Case 2 Failed: Expected root to be %s, got: %s", tempDir2, root2)
	}

	// Test case 3: No .git directory found
	tempDir3, err := os.MkdirTemp("", "git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir3)

	_, err3 := git.FindRepositoryRoot(tempDir3)
	if err3 == nil {
		t.Errorf("Test Case 3 Failed: Expected error, got nil")
	}
	if err3 != nil && !strings.Contains(err3.Error(), "no repository found") {
		t.Errorf("Test Case 3 Failed: Expected 'no repository found' error, got: %v", err3)
	}

	// Test case 4: Start from root-like directory, no .git (more robust no .git test)
	// Create a deeper temp dir structure to simulate starting further from root
	tempDir4Base, err := os.MkdirTemp("", "git-test-base-")
	if err != nil {
		t.Fatalf("Failed to create base temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir4Base)
	tempDir4 := filepath.Join(tempDir4Base, "level1", "level2")
	if err := os.MkdirAll(tempDir4, 0755); err != nil {
		t.Fatalf("Failed to create deep temp dir: %v", err)
	}

	_, err4 := git.FindRepositoryRoot(tempDir4)
	if err4 == nil {
		t.Errorf("Test Case 4 Failed: Expected error, got nil")
	}
	if err4 != nil && !strings.Contains(err4.Error(), "no repository found") {
		t.Errorf("Test Case 4 Failed: Expected 'no repository found' error, got: %v", err4)
	}
}

// MockGitExecutor is a mock implementation of GitExecutor for testing purposes.
type MockGitExecutor struct {
	MockListFileChanges func(repoDir string) (io.ReadCloser, error)
	MockIsAvailable     func() bool
}

func (m *MockGitExecutor) ListFileChanges(repoDir string) (io.ReadCloser, error) {
	if m.MockListFileChanges != nil {
		return m.MockListFileChanges(repoDir)
	}
	return io.NopCloser(strings.NewReader("")), nil // Default to no changes
}

func (m *MockGitExecutor) IsAvailable() bool {
	if m.MockIsAvailable != nil {
		return m.MockIsAvailable()
	}
	return true // Default to git available
}

func TestGetCommitCounts(t *testing.T) {
	tests := []struct {
		name              string
		listChangesOutput string
		listChangesError  error
		expectedCounts    map[string]int
		expectError       bool
	}{
		{
			name:              "No changes",
			listChangesOutput: "",
			expectedCounts:    map[string]int{},
		},
		{
			name:              "Single file, single commit",
			listChangesOutput: "file1.go\n",
			expectedCounts:    map[string]int{"file1.go": 1},
		},
		{
			name:              "Single file, multiple commits",
			listChangesOutput: "file1.go\nfile1.go\nfile1.go\n",
			expectedCounts:    map[string]int{"file1.go": 3},
		},
		{
			name:              "Multiple files, single commit each",
			listChangesOutput: "file1.go\nfile2.go\nfile3.go\n",
			expectedCounts:    map[string]int{"file1.go": 1, "file2.go": 1, "file3.go": 1},
		},
		{
			name: "Multiple files, multiple commits",
			listChangesOutput: `file1.go
file2.go
file1.go
file3.go
file2.go
`,
			expectedCounts: map[string]int{"file1.go": 2, "file2.go": 2, "file3.go": 1},
		},
		{
			name:              "Empty lines in output",
			listChangesOutput: "\nfile1.go\n\nfile2.go\n\n",
			expectedCounts:    map[string]int{"file1.go": 1, "file2.go": 1},
		},
		{
			name:             "Error from ListFileChanges",
			listChangesError: ErrTest, // Define a test error
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockGitExecutor{
				MockListFileChanges: func(repoDir string) (io.ReadCloser, error) {
					if tt.listChangesError != nil {
						return nil, tt.listChangesError
					}
					return io.NopCloser(strings.NewReader(tt.listChangesOutput)), nil
				},
			}
			git := NewGit(mockExecutor)

			counts, err := git.GetCommitCounts("dummyRepoDir")

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				return // Stop here for error cases
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(counts) != len(tt.expectedCounts) {
				t.Errorf("Counts map length mismatch: got %v, want %v", len(counts), len(tt.expectedCounts))
			}

			for file, expectedCount := range tt.expectedCounts {
				actualCount, ok := counts[file]
				if !ok {
					t.Errorf("Missing file in counts: %s", file)
					continue
				}
				if actualCount != expectedCount {
					t.Errorf("Count mismatch for file %s: got %d, want %d", file, actualCount, expectedCount)
				}
			}
		})
	}
}

func TestSortFilesByCommitCounts(t *testing.T) {
	tests := []struct {
		name          string
		filePaths     []string
		commitCounts  map[string]int
		expectedOrder []string
		expectError   bool // While sorting itself unlikely to error, including for completeness
	}{
		{
			name:          "Empty file list",
			filePaths:     []string{},
			commitCounts:  map[string]int{},
			expectedOrder: []string{},
		},
		{
			name:          "Single file",
			filePaths:     []string{"file1.go"},
			commitCounts:  map[string]int{"file1.go": 1},
			expectedOrder: []string{"file1.go"},
		},
		{
			name:          "Already sorted by commits",
			filePaths:     []string{"file1.go", "file2.go", "file3.go"},
			commitCounts:  map[string]int{"file1.go": 1, "file2.go": 2, "file3.go": 3},
			expectedOrder: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:          "Reverse sorted by commits",
			filePaths:     []string{"file3.go", "file2.go", "file1.go"},
			commitCounts:  map[string]int{"file1.go": 1, "file2.go": 2, "file3.go": 3},
			expectedOrder: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:          "Mixed commit counts",
			filePaths:     []string{"file3.go", "file1.go", "file2.go"},
			commitCounts:  map[string]int{"file1.go": 2, "file2.go": 1, "file3.go": 3},
			expectedOrder: []string{"file2.go", "file1.go", "file3.go"},
		},
		{
			name:          "Same commit counts, sorted alphabetically",
			filePaths:     []string{"file3.go", "file1.go", "file2.go"},
			commitCounts:  map[string]int{"file1.go": 1, "file2.go": 1, "file3.go": 1},
			expectedOrder: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:          "Mixed commit counts and same counts, alphabetical tie-breaker",
			filePaths:     []string{"fileC.go", "fileA.go", "fileB.go", "fileD.go"},
			commitCounts:  map[string]int{"fileA.go": 2, "fileB.go": 1, "fileC.go": 2, "fileD.go": 1},
			expectedOrder: []string{"fileB.go", "fileD.go", "fileA.go", "fileC.go"}, // B and D (count 1, then alpha), A and C (count 2, then alpha)
		},
		{
			name:          "File with no commit count (treated as 0 commits)",
			filePaths:     []string{"file2.go", "file1.go", "file3.go"},
			commitCounts:  map[string]int{"file1.go": 1, "file2.go": 2}, // file3.go has no count
			expectedOrder: []string{"file3.go", "file1.go", "file2.go"}, // file3.go (0), file1.go (1), file2.go (2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommitCounter := func(repoDir string) (map[string]int, error) {
				return tt.commitCounts, nil
			}

			git := NewGit(&MockGitExecutor{}) // Executor not used for this test
			sortedFiles, err := git.SortFilesByCommitCounts("dummyRepoDir", tt.filePaths, mockCommitCounter)

			if tt.expectError && err == nil {
				t.Fatalf("Expected error but got nil")
			} else if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if !slicesAreEqual(sortedFiles, tt.expectedOrder) {
					t.Errorf("Sorted file order mismatch: got %v, want %v", sortedFiles, tt.expectedOrder)
				}
			}
		})
	}
}

// Define a test error for error case testing
var ErrTest = TestError("test error")

type TestError string

func (e TestError) Error() string { return string(e) }

func slicesAreEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v := range s1 {
		if v != s2[i] {
			return false
		}
	}
	return true
}
