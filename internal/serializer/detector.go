package serializer

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// MinifiedFileThresholds contains configuration values for detecting minified files.
// It now includes general, JS-specific, and CSS-specific thresholds.
type MinifiedFileThresholds struct {
	// General thresholds for both file types.
	MaxLineLength        int     // Maximum line length to flag minification.
	MaxLinesPerCharRatio float64 // Maximum ratio of non-blank lines to total characters.
	MinNonBlankLines     int     // Minimum number of non-blank lines to perform detection.
	MinTotalChars        int     // Minimum file size to consider for detection.
	SingleLineMinLength  int     // For single-line files: length threshold for minification.

	// JS-specific thresholds.
	JS_SingleCharVarStrongThreshold   int // If count > this, immediately flag as minified.
	JS_SingleCharVarModerateThreshold int // If count > this, add to pattern score.
	JS_ShortParamsThreshold           int // Number of short parameter clusters to add score.
	JS_ChainedMethodsThreshold        int // Number of chained method patterns to add score.
	JS_PatternScoreThreshold          int // Overall pattern score threshold to flag minification.

	// CSS-specific thresholds.
	CSS_ClearIndicatorSemicolonCount    int // Clear indicator: high semicolon count.
	CSS_ClearIndicatorSpaceDivisor      int // Clear indicator: space count divisor.
	CSS_PatternIndicatorSemicolonCount  int // Pattern indicator: semicolon count.
	CSS_PatternIndicatorSpaceDivisor    int // Pattern indicator: space count divisor.
	CSS_NoSpacesAroundBracketsThreshold int // Threshold for missing spaces around brackets.
	CSS_NoSpacesAfterColonsThreshold    int // Threshold for missing spaces after colons.
	CSS_PatternScoreThreshold           int // Overall pattern score threshold.
}

// DefaultMinifiedFileThresholds defines default values for minified file detection.
var DefaultMinifiedFileThresholds = MinifiedFileThresholds{
	// General thresholds.
	MaxLineLength:        500,
	MaxLinesPerCharRatio: 0.02,
	MinNonBlankLines:     5,
	MinTotalChars:        200,
	SingleLineMinLength:  1000,

	// JS-specific thresholds.
	JS_SingleCharVarStrongThreshold:   15,
	JS_SingleCharVarModerateThreshold: 8,
	JS_ShortParamsThreshold:           3,
	JS_ChainedMethodsThreshold:        2,
	JS_PatternScoreThreshold:          3,

	// CSS-specific thresholds.
	CSS_ClearIndicatorSemicolonCount:    30,
	CSS_ClearIndicatorSpaceDivisor:      30,
	CSS_PatternIndicatorSemicolonCount:  20,
	CSS_PatternIndicatorSpaceDivisor:    20,
	CSS_NoSpacesAroundBracketsThreshold: 10,
	CSS_NoSpacesAfterColonsThreshold:    10,
	CSS_PatternScoreThreshold:           3,
}

// Precompiled regex patterns for JS heuristics.
var (
	reSingleCharVar  = regexp.MustCompile(`\b[a-z]\b=`)
	reShortParams    = regexp.MustCompile(`\([a-z],[a-z],[a-z](,[a-z])*\)`)
	reConsecutiveEnd = regexp.MustCompile(`\){4,}|\]{4,}`)
	reChainedMethods = regexp.MustCompile(`\.[a-zA-Z]+\([^)]*\)\.[a-zA-Z]+\([^)]*\)\.[a-zA-Z]+\(`)
	reLongOperators  = regexp.MustCompile(`[+\-/*&|^]{5,}`)
)

// Precompiled regex patterns for CSS heuristics.
var (
	reNoSpacesAroundBrackets = regexp.MustCompile(`[^\s{][{]|[}][^\s}]`)
	reNoSpacesAfterColons    = regexp.MustCompile(`:[^}\s]`)
)

// IsMinifiedFile returns true if the file content appears to be minified.
// It applies general heuristics and delegates to JS- or CSS-specific checks.
func IsMinifiedFile(content, filePath string, thresholds MinifiedFileThresholds) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".js" && ext != ".css" {
		return false
	}

	// Skip very small files.
	if len(content) < thresholds.MinTotalChars {
		return false
	}

	// Gather general file metrics.
	lines := strings.Split(content, "\n")
	nonBlankLines, maxLineLength := analyzeLines(lines)

	// Special handling for single-line files.
	if nonBlankLines == 1 && len(content) > thresholds.SingleLineMinLength {
		if strings.Count(content, "/*") < 3 && strings.Count(content, "//") < 5 {
			log.Debug().Str("path", filePath).Msg("Detected single-line minified file")
			return true
		}
	}

	// Apply general heuristics if file is sufficiently large.
	if nonBlankLines >= thresholds.MinNonBlankLines {
		if maxLineLength > thresholds.MaxLineLength {
			return true
		}
		if float64(nonBlankLines)/float64(len(content)) < thresholds.MaxLinesPerCharRatio {
			return true
		}
	}

	// Delegate file-type–specific checks.
	switch ext {
	case ".js":
		return isMinifiedJS(content, nonBlankLines, maxLineLength, thresholds)
	case ".css":
		return isMinifiedCSS(content, thresholds)
	}

	return false
}

// analyzeLines computes the number of non-blank lines and the maximum line length.
func analyzeLines(lines []string) (nonBlankLines, maxLineLength int) {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			nonBlankLines++
		}
		if len(line) > maxLineLength {
			maxLineLength = len(line)
		}
	}
	return
}

// isMinifiedJS applies JavaScript-specific heuristics.
func isMinifiedJS(content string, nonBlankLines, maxLineLength int, thresholds MinifiedFileThresholds) bool {
	// Check for a sourceMappingURL which is a strong signal.
	if strings.Contains(content, "sourceMappingURL") {
		return true
	}

	// Only check patterns if content appears suspicious.
	shouldCheckPatterns := nonBlankLines < thresholds.MinNonBlankLines || maxLineLength > thresholds.MaxLineLength/2
	if !shouldCheckPatterns {
		return false
	}

	patternScore := 0

	// Heuristic: Many single-character variables.
	singleCharCount := len(reSingleCharVar.FindAllString(content, -1))
	if singleCharCount > thresholds.JS_SingleCharVarStrongThreshold {
		return true
	} else if singleCharCount > thresholds.JS_SingleCharVarModerateThreshold {
		patternScore += 2
	}

	// Heuristic: Clusters of short parameter names.
	if len(reShortParams.FindAllString(content, -1)) > thresholds.JS_ShortParamsThreshold {
		patternScore += 2
	}

	// Heuristic: Consecutive closing brackets/parentheses.
	if reConsecutiveEnd.MatchString(content) {
		patternScore++
	}

	// Heuristic: Chained methods without spacing.
	if len(reChainedMethods.FindAllString(content, -1)) > thresholds.JS_ChainedMethodsThreshold {
		patternScore += 2
	}

	// Heuristic: Long strings of operators.
	if reLongOperators.MatchString(content) {
		patternScore++
	}

	return patternScore >= thresholds.JS_PatternScoreThreshold
}

// isMinifiedCSS applies CSS-specific heuristics.
func isMinifiedCSS(content string, thresholds MinifiedFileThresholds) bool {
	semicolonCount := strings.Count(content, ";")
	spaceCount := strings.Count(content, " ")

	// Very clear indicator: high semicolon count and very few spaces.
	if semicolonCount > thresholds.CSS_ClearIndicatorSemicolonCount &&
		spaceCount < len(content)/thresholds.CSS_ClearIndicatorSpaceDivisor {
		return true
	}

	cssPatternScore := 0

	// Moderate indicator: many semicolons with relatively few spaces.
	if semicolonCount > thresholds.CSS_PatternIndicatorSemicolonCount &&
		spaceCount < len(content)/thresholds.CSS_PatternIndicatorSpaceDivisor {
		cssPatternScore += 2
	}

	// Check for lack of spacing around brackets.
	if len(reNoSpacesAroundBrackets.FindAllString(content, -1)) > thresholds.CSS_NoSpacesAroundBracketsThreshold {
		cssPatternScore += 2
	}

	// Check for missing spaces after colons.
	if len(reNoSpacesAfterColons.FindAllString(content, -1)) > thresholds.CSS_NoSpacesAfterColonsThreshold {
		cssPatternScore++
	}

	// Check for !important usage without spacing.
	if strings.Count(content, "!important") < strings.Count(content, "!important;") {
		cssPatternScore++
	}

	return cssPatternScore >= thresholds.CSS_PatternScoreThreshold
}
