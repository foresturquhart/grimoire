package core

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// helper to create file with contents
func createFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestDefaultWalker_Walk(t *testing.T) {
	tmp := t.TempDir()

	// create files and directories
	createFile(t, filepath.Join(tmp, "file1.go"))
	createFile(t, filepath.Join(tmp, "skipme.go"))
	createFile(t, filepath.Join(tmp, "not_allowed.txt"))
	createFile(t, filepath.Join(tmp, "output.md"))

	// directory excluded via .grimoireignore
	exclDir := filepath.Join(tmp, "excluded_dir")
	createFile(t, filepath.Join(exclDir, "c.go"))

	// subdir with gitignore rule
	subdir := filepath.Join(tmp, "subdir1")
	createFile(t, filepath.Join(subdir, "a.go"))
	createFile(t, filepath.Join(subdir, "ignored.go"))
	createFile(t, filepath.Join(subdir, "nested", "b.go"))

	// directory ignored via regex
	createFile(t, filepath.Join(tmp, "skip_dir", "d.go"))

	// write ignore files
	os.WriteFile(filepath.Join(tmp, ".grimoireignore"), []byte("excluded_dir/\n"), 0o644)
	os.WriteFile(filepath.Join(subdir, ".gitignore"), []byte("ignored.go\n"), 0o644)

	allowed := map[string]bool{".go": true}
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`skip_dir`),
		regexp.MustCompile(`^skipme\.go$`),
	}

	dw := NewDefaultWalker(tmp, allowed, regexes, filepath.Join(tmp, "output.md"))
	files, err := dw.Walk()
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	expected := map[string]bool{
		"file1.go": true,
		filepath.ToSlash(filepath.Join("subdir1", "a.go")):           true,
		filepath.ToSlash(filepath.Join("subdir1", "nested", "b.go")): true,
	}
	if len(files) != len(expected) {
		t.Fatalf("expected %d files, got %d: %v", len(expected), len(files), files)
	}
	for _, f := range files {
		if !expected[f] {
			t.Errorf("unexpected file %s", f)
		}
		delete(expected, f)
	}
	for k := range expected {
		t.Errorf("missing expected file %s", k)
	}
}
