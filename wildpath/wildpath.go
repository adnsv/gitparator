package wildpath

import (
	"strings"
)

// Match checks if the given filename matches the pattern.
// Supports gitignore-style syntax:
//   - * matches any sequence of characters within a path component
//   - ? matches any single character
//   - ** matches zero or more directories
//   - [abc] matches any character in brackets
//   - [a-z] matches any character in the range
//   - [!abc] or [^abc] matches any character not in brackets
//   - {js,ts} matches any of the comma-separated patterns
//   - Leading / makes the pattern root-relative
//
// Match checks if the given filename matches the pattern.
// Supports gitignore-style syntax:
//   - * matches any sequence of characters within a path component
//   - ? matches any single character
//   - ** matches zero or more directories
//   - [abc] matches any character in brackets
//   - [a-z] matches any character in the range
//   - [!abc] or [^abc] matches any character not in brackets
//   - {js,ts} matches any of the comma-separated patterns
//   - Leading / makes the pattern root-relative
func Match(pattern, filename string) bool {
	// Handle brace expansion
	if strings.Contains(pattern, "{") {
		patterns := expandBraces(pattern)
		for _, p := range patterns {
			if matchSinglePattern(p, filename) {
				return true
			}
		}
		return false
	}
	return matchSinglePattern(pattern, filename)
}

// expandBraces expands patterns like "*.{js,ts}" into []string{"*.js", "*.ts"}
func expandBraces(pattern string) []string {
	start := strings.Index(pattern, "{")
	if start == -1 {
		return []string{pattern}
	}

	end := strings.Index(pattern[start:], "}")
	if end == -1 {
		return []string{pattern} // unclosed brace, treat as literal
	}
	end += start

	// Get content between braces
	content := pattern[start+1 : end]

	// Empty braces or no comma - treat as literal
	if content == "" || !strings.Contains(content, ",") {
		return []string{pattern}
	}

	prefix := pattern[:start]
	suffix := pattern[end+1:]
	alternatives := strings.Split(content, ",")

	var results []string
	// Recursively handle nested braces in suffix
	suffixExpanded := expandBraces(suffix)

	for _, alt := range alternatives {
		for _, suffixPattern := range suffixExpanded {
			results = append(results, prefix+alt+suffixPattern)
		}
	}

	return results
}

func matchSinglePattern(pattern, filename string) bool {
	// Normalize paths by removing consecutive slashes
	// Keep track if pattern starts with slash (root-relative)
	patternParts, patternHasRoot := normalize(pattern)
	filenameParts, filenameHasRoot := normalize(filename)

	// If pattern is root-relative, the file path must also be root-relative
	if patternHasRoot != filenameHasRoot {
		return false
	}

	return matchParts(patternParts, filenameParts, 0, 0)
}

func normalize(s string) ([]string, bool) {
	// Track if pattern starts with slash
	hasRoot := strings.HasPrefix(s, "/")

	// Split by slash and filter out empty parts
	parts := strings.Split(s, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result, hasRoot
}

func matchParts(pattern, filename []string, patternIdx, filenameIdx int) bool {
	for patternIdx < len(pattern) {
		// If we've consumed all filename parts
		if filenameIdx == len(filename) {
			// Skip over trailing ** patterns
			for patternIdx < len(pattern) && pattern[patternIdx] == "**" {
				patternIdx++
			}
			// Return true if we've consumed all patterns
			return patternIdx == len(pattern)
		}

		// Handle globstar (**) pattern
		if pattern[patternIdx] == "**" {
			// Try matching the rest of the pattern with current and all remaining positions
			nextPattern := patternIdx + 1
			if nextPattern == len(pattern) {
				return true // ** at end matches everything
			}

			// Try matching rest of pattern at current position and every subsequent position
			for i := filenameIdx; i <= len(filename); i++ {
				if matchParts(pattern, filename, nextPattern, i) {
					return true
				}
			}
			return false
		}

		// If we have filename parts to match
		if filenameIdx < len(filename) {
			if !matchSinglePart(pattern[patternIdx], filename[filenameIdx]) {
				return false
			}
			patternIdx++
			filenameIdx++
			continue
		}

		return false
	}

	// Both pattern and filename should be fully consumed
	return filenameIdx == len(filename)
}

func matchSinglePart(pattern, str string) bool {
	if pattern == "*" || pattern == str {
		return true
	}

	p := []rune(pattern)
	s := []rune(str)

	i, j := 0, 0
	starIdx := -1
	starMatch := 0

	for j < len(s) {
		if i < len(p) && (p[i] == '*') {
			starIdx = i
			starMatch = j
			i++
		} else if i < len(p) && (p[i] == '?' || p[i] == s[j]) {
			i++
			j++
		} else if i < len(p) && p[i] == '[' {
			closeIdx := findClosingBracket(p[i:])
			if closeIdx == -1 {
				return false
			}
			if matchCharacterRange(p[i+1:i+closeIdx], s[j]) {
				i += closeIdx + 1
				j++
			} else {
				if starIdx == -1 {
					return false
				}
				i = starIdx + 1
				starMatch++
				j = starMatch
			}
		} else {
			if starIdx == -1 {
				return false
			}
			i = starIdx + 1
			starMatch++
			j = starMatch
		}
	}

	for i < len(p) && p[i] == '*' {
		i++
	}

	return i == len(p)
}

func findClosingBracket(pattern []rune) int {
	for i := 1; i < len(pattern); i++ {
		if pattern[i] == ']' {
			return i
		}
	}
	return -1
}

func matchCharacterRange(rangePattern []rune, char rune) bool {
	if len(rangePattern) == 0 {
		return false
	}

	isNegated := false
	startIdx := 0
	if rangePattern[0] == '!' || rangePattern[0] == '^' {
		isNegated = true
		startIdx = 1
	}

	matched := false
	for i := startIdx; i < len(rangePattern); i++ {
		if i+2 < len(rangePattern) && rangePattern[i+1] == '-' {
			start := rangePattern[i]
			end := rangePattern[i+2]
			if char >= start && char <= end {
				matched = true
				break
			}
			i += 2
		} else {
			if rangePattern[i] == char {
				matched = true
				break
			}
		}
	}

	return matched != isNegated
}
