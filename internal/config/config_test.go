package config

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestNewConfigFromCommand(t *testing.T) {
	tmp := t.TempDir()
	output := filepath.Join(tmp, "out.xml")

	var cfg *Config
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "output"},
			&cli.BoolFlag{Name: "force"},
			&cli.BoolFlag{Name: "no-tree"},
			&cli.BoolFlag{Name: "no-sort"},
			&cli.BoolFlag{Name: "ignore-secrets"},
			&cli.BoolFlag{Name: "redact-secrets"},
			&cli.BoolFlag{Name: "skip-token-count"},
			&cli.StringFlag{Name: "format"},
			&cli.IntFlag{Name: "high-token-threshold"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cfg = NewConfigFromCommand(c)
			return nil
		},
	}

	args := []string{
		"cmd", tmp,
		"--output", output,
		"--force",
		"--no-tree",
		"--no-sort",
		"--ignore-secrets",
		"--redact-secrets",
		"--skip-token-count",
		"--format", "xml",
		"--high-token-threshold", "9000",
	}
	if err := cmd.Run(context.Background(), args); err != nil {
		t.Fatalf("run command: %v", err)
	}

	if cfg == nil {
		t.Fatalf("config not set")
	}

	if cfg.TargetDir != tmp {
		t.Errorf("target dir: got %s want %s", cfg.TargetDir, tmp)
	}
	if cfg.OutputFile != output {
		t.Errorf("output: got %s want %s", cfg.OutputFile, output)
	}
	if !cfg.Force {
		t.Errorf("force not set")
	}
	if cfg.ShowTree {
		t.Errorf("show tree should be false")
	}
	if !cfg.DisableSort {
		t.Errorf("disable sort not set")
	}
	if cfg.Format != "xml" {
		t.Errorf("format got %s", cfg.Format)
	}
	if !cfg.IgnoreSecrets || !cfg.RedactSecrets {
		t.Errorf("secret flags not set")
	}
	if !cfg.SkipTokenCount {
		t.Errorf("skip token count not set")
	}
	if cfg.HighTokenThreshold != 9000 {
		t.Errorf("high token threshold got %d", cfg.HighTokenThreshold)
	}
	if cfg.LargeFileSizeThreshold != DefaultLargeFileSizeThreshold {
		t.Errorf("large file size threshold mismatch")
	}
	if !cfg.AllowedFileExtensions[".go"] {
		t.Errorf("expected .go in allowed extensions")
	}
	if len(cfg.IgnoredPathRegexes) != len(DefaultIgnoredPathPatterns) {
		t.Errorf("regex count mismatch")
	}
	if !cfg.ShouldWriteFile() {
		t.Errorf("ShouldWriteFile expected true")
	}
}

func TestCompileRegexesInvalid(t *testing.T) {
	_, err := compileRegexes([]string{"["})
	if err == nil {
		t.Fatalf("expected error for invalid regex")
	}
}
