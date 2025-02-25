package secrets

import (
	_ "embed"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/sources"
	"path/filepath"
)

//go:embed gitleaks.toml
var DefaultConfig []byte

// Finding represents a simplified gitleaks finding for easier consumption
type Finding struct {
	Description string
	Secret      string
	File        string
	Line        int
}

// Detector provides functionality to scan files for secrets
type Detector struct {
	detector *detect.Detector
}

// NewDetector creates a new secrets detector using the provided configuration
func NewDetector() (*Detector, error) {
	var vc config.ViperConfig
	if err := toml.Unmarshal(DefaultConfig, &vc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gitleaks config: %w", err)
	}

	cfg, err := vc.Translate()
	if err != nil {
		return nil, fmt.Errorf("failed to translate gitleaks config: %w", err)
	}

	return &Detector{
		detector: detect.NewDetector(cfg),
	}, nil
}

// DetectSecretsInFiles scans the provided file paths for secrets
// Returns a slice of findings and a boolean indicating if any secrets were found
func (d *Detector) DetectSecretsInFiles(filePaths []string) ([]Finding, bool, error) {
	if len(filePaths) == 0 {
		return nil, false, nil
	}

	scanTargetChan := make(chan sources.ScanTarget, len(filePaths))
	for _, path := range filePaths {
		// Make sure the path is absolute
		absPath, err := filepath.Abs(path)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to get absolute path for %s", path)
			absPath = path // Fall back to original path
		}
		scanTargetChan <- sources.ScanTarget{Path: absPath}
	}
	close(scanTargetChan)

	gitleaksFindings, err := d.detector.DetectFiles(scanTargetChan)
	if err != nil {
		return nil, false, fmt.Errorf("failed to detect secrets: %w", err)
	}

	// Convert findings to our simplified format
	findings := make([]Finding, 0, len(gitleaksFindings))
	for _, f := range gitleaksFindings {
		findings = append(findings, Finding{
			Description: f.Description,
			Secret:      f.Secret,
			File:        f.File,
			Line:        f.StartLine,
		})
	}

	return findings, len(findings) > 0, nil
}
