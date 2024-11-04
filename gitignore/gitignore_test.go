package gitignore

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestStack_ShouldIgnore(t *testing.T) {
	// Helper function to normalize paths for cross-platform testing
	normPath := func(path string) string {
		return filepath.ToSlash(path)
	}

	// Helper function to create a stack with patterns
	setupStack := func(basePath string, patternLevels ...[]string) *Stack {
		stack := New(basePath)
		for _, patterns := range patternLevels {
			stack.PushPatterns(patterns)
		}
		return stack
	}

	tests := []struct {
		name           string
		basePath       string
		patternLevels  [][]string
		testPath       string
		expectedIgnore bool
	}{
		{
			name:     "simple direct match",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt"},
			},
			testPath:       "/project/test.txt",
			expectedIgnore: true,
		},
		{
			name:     "simple non-match",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt"},
			},
			testPath:       "/project/test.go",
			expectedIgnore: false,
		},
		{
			name:     "nested directory pattern",
			basePath: "/project",
			patternLevels: [][]string{
				{"docs/**/*.pdf"},
			},
			testPath:       "/project/docs/subfolder/file.pdf",
			expectedIgnore: true,
		},
		{
			name:     "multiple pattern levels - match in parent",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt"},          // root level
				{"!important.txt"}, // subdirectory level
				{"temp/*.txt"},     // sub-subdirectory level
			},
			testPath:       "/project/normal.txt",
			expectedIgnore: true,
		},
		{
			name:     "multiple pattern levels - match in child",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt"},          // root level
				{"!important.txt"}, // subdirectory level
				{"temp/*.txt"},     // sub-subdirectory level
			},
			testPath:       "/project/subdir/temp/test.txt",
			expectedIgnore: true,
		},
		{
			name:     "negated pattern",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt", "!important.txt"},
			},
			testPath:       "/project/important.txt",
			expectedIgnore: false,
		},
		{
			name:     "directory-specific pattern",
			basePath: "/project",
			patternLevels: [][]string{
				{"node_modules/"},
			},
			testPath:       "/project/node_modules/package.json",
			expectedIgnore: true,
		},
		{
			name:     "complex nested patterns",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.log", "build/"},                // root patterns
				{"!important.log", "temp/"},        // first level
				{"**/*.tmp", "!temp/keepthis.tmp"}, // second level
			},
			testPath:       "/project/logs/important.log",
			expectedIgnore: false,
		},
		{
			name:     "pattern with special characters",
			basePath: "/project",
			patternLevels: [][]string{
				{"[a-z]*.txt"},
			},
			testPath:       "/project/abc123.txt",
			expectedIgnore: true,
		},
		{
			name:     "relative path pattern",
			basePath: "/project",
			patternLevels: [][]string{
				{"foo/bar/*.txt"},
			},
			testPath:       "/project/foo/bar/test.txt",
			expectedIgnore: true,
		},
		{
			name:     "outside base directory",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt"},
			},
			testPath:       "/other/test.txt",
			expectedIgnore: false,
		},
		{
			name:           "empty pattern stack",
			basePath:       "/project",
			patternLevels:  [][]string{},
			testPath:       "/project/anything.txt",
			expectedIgnore: false,
		},
		{
			name:     "pattern with spaces",
			basePath: "/project",
			patternLevels: [][]string{
				{"* *.txt", "test space.log"},
			},
			testPath:       "/project/hello world.txt",
			expectedIgnore: true,
		},
		{
			name:     "case sensitivity test",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.TXT"},
			},
			testPath:       "/project/test.txt",
			expectedIgnore: runtime.GOOS != "windows", // Windows is case-insensitive
		},
		{
			name:     "nested patterns override parent",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.log"},             // root level
				{"!debug.log"},        // override in subdirectory
				{"debug/special.log"}, // specific file in subdir
			},
			testPath:       "/project/debug/special.log",
			expectedIgnore: true,
		},
		{
			name:     "multiple negations",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.log"},
				{"!important/*.log"},
				{"important/temp/*.log"},
				{"!important/temp/debug.log"},
			},
			testPath:       "/project/important/temp/debug.log",
			expectedIgnore: false,
		},
		{
			name:     "directory pattern with subdirs",
			basePath: "/project",
			patternLevels: [][]string{
				{"node_modules/"},
			},
			testPath:       "/project/packages/node_modules/some/deep/file.js",
			expectedIgnore: true,
		},
		{
			name:     "directory pattern exact match",
			basePath: "/project",
			patternLevels: [][]string{
				{"temp/"},
			},
			testPath:       "/project/temp",
			expectedIgnore: false, // Should not ignore the directory itself
		},
		{
			name:     "multiple star patterns",
			basePath: "/project",
			patternLevels: [][]string{
				{"**/*.{js,ts}"},
			},
			testPath:       "/project/src/deep/nested/file.ts",
			expectedIgnore: true,
		},
		{
			name:     "character class with negation",
			basePath: "/project",
			patternLevels: [][]string{
				{"**/[!.]*"},
			},
			testPath:       "/project/.hidden",
			expectedIgnore: false,
		},
		{
			name:     "backslash in pattern",
			basePath: "/project",
			patternLevels: [][]string{
				{"foo\\bar\\*.txt"},
			},
			testPath:       "/project/foo/bar/test.txt",
			expectedIgnore: true,
		},
		{
			name:     "mixed slashes in path",
			basePath: "/project",
			patternLevels: [][]string{
				{"docs/**/*.md"},
			},
			testPath:       "/project/docs\\subfolder\\README.md",
			expectedIgnore: true,
		},
		{
			name:     "absolute path pattern",
			basePath: "/project",
			patternLevels: [][]string{
				{"/absolute/*.txt"},
			},
			testPath:       "/project/absolute/file.txt",
			expectedIgnore: true,
		},
		{
			name:     "dot-dot in path",
			basePath: "/project",
			patternLevels: [][]string{
				{"**/*.txt"},
			},
			testPath:       "/project/../outside.txt",
			expectedIgnore: false,
		},
		{
			name:     "double star at start",
			basePath: "/project",
			patternLevels: [][]string{
				{"**/node_modules/**"},
			},
			testPath:       "/project/any/path/node_modules/file.js",
			expectedIgnore: true,
		},
		{
			name:     "double star at end",
			basePath: "/project",
			patternLevels: [][]string{
				{"build/**"},
			},
			testPath:       "/project/build/any/path/file.txt",
			expectedIgnore: true,
		},
		{
			name:     "multiple consecutive slashes",
			basePath: "/project",
			patternLevels: [][]string{
				{"docs///temp///*.txt"},
			},
			testPath:       "/project/docs/temp/file.txt",
			expectedIgnore: true,
		},
		{
			name:     "trailing slash in pattern",
			basePath: "/project",
			patternLevels: [][]string{
				{"temp/"},
			},
			testPath:       "/project/temp/file.txt",
			expectedIgnore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := setupStack(normPath(tt.basePath), tt.patternLevels...)
			got := stack.ShouldIgnore(normPath(tt.testPath))
			if got != tt.expectedIgnore {
				t.Errorf("Stack.ShouldIgnore() = %v, want %v", got, tt.expectedIgnore)
				t.Logf("Patterns: %v", tt.patternLevels)
				t.Logf("Test path: %s", tt.testPath)
			}
		})
	}
}

// Test pattern stack manipulation
func TestStack_PatternManipulation(t *testing.T) {
	stack := New("/project")

	// Test pushing patterns
	patterns1 := []string{"*.txt", "*.log"}
	patterns2 := []string{"!important.txt"}
	stack.PushPatterns(patterns1)
	stack.PushPatterns(patterns2)

	if len(stack.patterns) != 2 {
		t.Errorf("Expected 2 pattern levels, got %d", len(stack.patterns))
	}

	// Test popping patterns
	stack.PopPatterns()
	if len(stack.patterns) != 1 {
		t.Errorf("Expected 1 pattern level after pop, got %d", len(stack.patterns))
	}

	// Test popping when empty
	stack.PopPatterns()
	stack.PopPatterns() // Should not panic
	if len(stack.patterns) != 0 {
		t.Errorf("Expected empty pattern stack, got %d levels", len(stack.patterns))
	}
}

// Test edge cases
func TestStack_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		patterns       []string
		testPath       string
		expectedIgnore bool
	}{
		{
			name:           "empty base path",
			basePath:       "",
			patterns:       []string{"*.txt"},
			testPath:       "test.txt",
			expectedIgnore: true,
		},
		{
			name:           "empty pattern",
			basePath:       "/project",
			patterns:       []string{""},
			testPath:       "/project/test.txt",
			expectedIgnore: false,
		},
		{
			name:           "invalid pattern",
			basePath:       "/project",
			patterns:       []string{"[invalid"},
			testPath:       "/project/test.txt",
			expectedIgnore: false,
		},
		{
			name:           "dot files",
			basePath:       "/project",
			patterns:       []string{".*"},
			testPath:       "/project/.gitignore",
			expectedIgnore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := New(tt.basePath)
			stack.PushPatterns(tt.patterns)
			got := stack.ShouldIgnore(tt.testPath)
			if got != tt.expectedIgnore {
				t.Errorf("Stack.ShouldIgnore() = %v, want %v", got, tt.expectedIgnore)
			}
		})
	}
}

// Add new test function for pattern stack ordering
func TestStack_PatternOrder(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		patternLevels  [][]string
		testPath       string
		expectedIgnore bool
	}{
		{
			name:     "later patterns override earlier ones",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt"},
				{"!important.txt"},
			},
			testPath:       "/project/important.txt",
			expectedIgnore: false,
		},
		{
			name:     "negation followed by re-ignore in same level",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt", "!important.txt", "*.txt"},
			},
			testPath:       "/project/important.txt",
			expectedIgnore: true,
		},
		{
			name:     "multiple negations with final negation",
			basePath: "/project",
			patternLevels: [][]string{
				{"*.txt", "!important.txt", "*.txt", "!important.txt"},
			},
			testPath:       "/project/important.txt",
			expectedIgnore: false, // The final !important.txt takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := New(tt.basePath)
			for _, patterns := range tt.patternLevels {
				stack.PushPatterns(patterns)
			}
			got := stack.ShouldIgnore(tt.testPath)
			if got != tt.expectedIgnore {
				t.Errorf("Stack.ShouldIgnore() = %v, want %v", got, tt.expectedIgnore)
				t.Logf("Patterns: %v", tt.patternLevels)
				t.Logf("Test path: %s", tt.testPath)
			}
		})
	}
}
