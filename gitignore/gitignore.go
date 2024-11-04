package gitignore

import (
	"path/filepath"
	"strings"

	"github.com/adnsv/gitparator/wildpath"
)

type Stack struct {
	patterns [][]string
	basePath string
}

func New(basePath string) *Stack {
	basePath = filepath.ToSlash(basePath)
	return &Stack{
		patterns: make([][]string, 0),
		basePath: basePath,
	}
}

func (s *Stack) PushPatterns(patterns []string) {
	normalizedPatterns := make([]string, len(patterns))
	for i, pattern := range patterns {
		normalizedPatterns[i] = filepath.ToSlash(pattern)
	}
	s.patterns = append(s.patterns, normalizedPatterns)
}

func (s *Stack) PopPatterns() {
	if len(s.patterns) > 0 {
		s.patterns = s.patterns[:len(s.patterns)-1]
	}
}

func (s *Stack) ShouldIgnore(path string) bool {
	// Normalize input path to forward slashes
	path = filepath.ToSlash(path)

	// Make path relative to base directory
	relPath, err := filepath.Rel(s.basePath, path)
	if err != nil {
		return false
	}
	// Ensure relative path uses forward slashes
	relPath = filepath.ToSlash(relPath)

	// Check if path is outside base directory
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	// Process patterns from most specific (last) to least specific (first)
	for i := len(s.patterns) - 1; i >= 0; i-- {
		levelPatterns := s.patterns[i]
		levelResult := false
		foundMatch := false

		// Process patterns within each level from first to last
		for j := 0; j < len(levelPatterns); j++ {
			pattern := levelPatterns[j]
			// Skip empty patterns
			if pattern == "" {
				continue
			}

			isNegated := strings.HasPrefix(pattern, "!")
			if isNegated {
				pattern = pattern[1:] // Remove the ! prefix
			}

			// Handle absolute path patterns
			if strings.HasPrefix(pattern, "/") {
				pattern = pattern[1:] // Remove leading slash
			}

			// Handle directory-specific patterns
			if strings.HasSuffix(pattern, "/") {
				dirPattern := strings.TrimSuffix(pattern, "/")
				// Try matching both with and without **/ prefix for directory patterns
				matched := wildpath.Match("**/"+dirPattern+"/**/*", relPath)
				if !matched {
					matched = wildpath.Match(dirPattern+"/**/*", relPath)
				}
				if matched {
					foundMatch = true
					levelResult = !isNegated
				}
				continue
			}

			// For patterns without slashes, try both with and without **/ prefix
			matched := false
			if !strings.Contains(pattern, "/") {
				// Try with **/ prefix first
				matched = wildpath.Match("**/"+pattern, relPath)
				if !matched {
					// If that fails, try without prefix
					matched = wildpath.Match(pattern, relPath)
				}
			} else {
				// For patterns with slashes, use as-is
				matched = wildpath.Match(pattern, relPath)
			}

			if matched {
				foundMatch = true
				levelResult = !isNegated
			}
		}

		// If we found any match in this level, return its result
		if foundMatch {
			return levelResult
		}
	}

	return false
}
