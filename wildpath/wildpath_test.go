package wildpath

import (
	"reflect"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		// Basic literal matching
		{"exact match", "file.txt", "file.txt", true},
		{"exact mismatch", "file.txt", "file.exe", false},
		{"subpath no match", "file.txt", "dir/file.txt", false},

		// Single star * matching
		{"star suffix", "*.txt", "file.txt", true},
		{"star prefix", "file.*", "file.txt", true},
		{"star middle", "file*.txt", "file123.txt", true},
		{"star multiple", "fi*le*.txt", "file123.txt", true},
		{"star no match extension", "*.txt", "file.exe", false},
		{"star no match path", "*.txt", "dir/file.txt", false},

		// Question mark ? matching
		{"question single", "file.???", "file.txt", true},
		{"question multiple", "???.txt", "abc.txt", true},
		{"question mix stars", "?*?.*", "abc.txt", true},
		{"question no match", "???.txt", "abcd.txt", false},

		// Character ranges []
		{"range single letter", "[a-z].txt", "a.txt", true},
		{"range single digit", "[0-9].txt", "5.txt", true},
		{"range multiple chars", "[a-z]*.txt", "abc123.txt", true},
		{"range uppercase", "[A-Z]*.txt", "Test.txt", true},
		{"range mixed case", "[a-zA-Z]*.txt", "Test.txt", true},
		{"range alphanumeric", "[a-zA-Z0-9]*.txt", "Test123.txt", true},
		{"range no match", "[a-z].txt", "5.txt", false},
		{"range negated", "[!0-9]*.txt", "abc.txt", true},
		{"range negated no match", "[!a-z]*.txt", "abc.txt", false},
		{"range complex", "[a-z][0-9][0-9].txt", "a12.txt", true},
		{"range with dash", "[a-z-]*.txt", "a-bc.txt", true},
		{"range single char", "[abc]*.txt", "abc.txt", true},
		{"range single no match", "[abc]*.txt", "def.txt", false},

		// Directory separator /
		{"dir exact", "dir/file.txt", "dir/file.txt", true},
		{"dir no match", "dir/file.txt", "dir/subdir/file.txt", false},
		{"dir prefix", "/dir/file.txt", "dir/file.txt", false}, // root-relative
		{"dir wildcard", "dir/*.txt", "dir/file.txt", true},
		{"dir wildcard no match", "dir/*.txt", "dir/subdir/file.txt", false},

		// Globstar ** matching
		{"globstar both ends", "**/file.txt", "file.txt", true},
		{"globstar both ends deep", "**/file.txt", "deep/path/file.txt", true},
		{"globstar prefix", "**/test/*.txt", "deep/path/test/file.txt", true},
		{"globstar suffix", "dir/**/*.txt", "dir/deep/path/file.txt", true},
		{"globstar middle", "dir/**/test/*.txt", "dir/deep/path/test/file.txt", true},
		{"globstar empty", "dir/**", "dir", true},
		{"globstar single", "dir/**", "dir/file.txt", true},
		{"globstar multiple", "dir/**", "dir/sub/file.txt", true},
		{"globstar no match", "dir/**/*.txt", "other/file.txt", false},

		// Complex patterns
		{"complex 1", "**/[a-z]*/[0-9]*.txt", "deep/abc/123.txt", true},
		{"complex 2", "**/*[!.]*/*.txt", "dir/test_file/doc.txt", true},
		{"complex 3", "**/[a-z][a-z][0-9][0-9]/**/*.txt", "path/ab12/deep/file.txt", true},
		{"complex no match", "**/[a-z][a-z][0-9][0-9]/**/*.txt", "path/12ab/deep/file.txt", false},

		// Edge cases
		{"empty pattern", "", "", true},
		{"empty pattern no match", "", "file.txt", false},
		{"pattern with spaces", "* *.txt", "a b.txt", true},
		{"unclosed range", "[a-z.txt", "[a-z.txt", true},     // treated as literal
		{"escaped range", "\\[a-z].txt", "[a-z].txt", false}, // we don't support escaping
		{"multiple stars", "**.txt", "file.txt", true},
		{"mixed slashes", "dir/*/file.txt", "dir\\sub\\file.txt", false}, // strict slash matching

		// Globstar matching empty/same directory
		{"globstar empty dir", "dir/**", "dir", true},
		{"globstar single dir", "dir/**", "dir/", true},
		{"globstar empty with file", "dir/**/*.txt", "dir/file.txt", true},

		// Multiple globstars
		{"multiple globstars empty", "dir/**/**", "dir", true},
		{"multiple globstars file", "dir/**/**/*.txt", "dir/file.txt", true},
		{"multiple globstars nested", "dir/**/**/*.txt", "dir/a/b/file.txt", true},

		// Globstar combinations
		{"globstar prefix empty", "**/dir", "dir", true},
		{"globstar suffix empty", "dir/**", "dir", true},
		{"globstar middle empty", "dir/**/end", "dir/end", true},

		// Edge cases with empty parts
		{"empty directory parts", "dir///**", "dir", true},
		{"trailing slash dir", "dir/", "dir", true},
		{"trailing slash globstar", "dir/**", "dir/", true},
		{"multiple slashes", "dir////**///*.txt", "dir/a/b/file.txt", true},

		// Negative cases
		{"different dir", "dir/**", "dir2", false},
		{"parent dir", "dir/**", "..", false},
		{"partial dir match", "dir/**", "directory", false},

		// Root path patterns
		{"root exact match", "/dir/file.txt", "/dir/file.txt", true},
		{"root vs non-root", "/dir/file.txt", "dir/file.txt", false},
		{"non-root vs root", "dir/file.txt", "/dir/file.txt", false},
		{"root with globstar", "/dir/**/*.txt", "/dir/sub/file.txt", true},
		{"root with globstar no match", "/dir/**/*.txt", "dir/sub/file.txt", false},

		// Multiple leading slashes (should be normalized)
		{"multiple root slashes", "///dir/file.txt", "/dir/file.txt", true},
		{"multiple root vs non-root", "///dir/file.txt", "dir/file.txt", false},

		// Root paths with various patterns
		{"root wildcard", "/*.txt", "/file.txt", true},
		{"root wildcard no match", "/*.txt", "file.txt", false},
		{"root character class", "/[a-z]*.txt", "/test.txt", true},
		{"root character class no match", "/[a-z]*.txt", "test.txt", false},

		// Edge cases
		{"root only", "/", "/", true},
		{"root vs empty", "/", "", false},
		{"empty vs root", "", "/", false},
		{"root globstar", "/**", "/dir/file.txt", true},
		{"non-root globstar", "**", "dir/file.txt", true},

		// Existing test cases should still pass
		{"relative path", "dir/file.txt", "dir/file.txt", true},
		{"relative globstar", "dir/**", "dir", true},
		{"relative with slashes", "dir///**", "dir", true},
		{"relative complex", "dir////**///*.txt", "dir/a/b/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v",
					tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestExpandBraces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		// Normal cases
		{"simple", "*.{js,ts}", []string{"*.js", "*.ts"}},
		{"multiple", "*.{js,ts,jsx}", []string{"*.js", "*.ts", "*.jsx"}},
		{"with path", "{src,lib}/*.js", []string{"src/*.js", "lib/*.js"}},

		// Edge cases - braces should not expand
		{"empty braces", "file.{}", []string{"file.{}"}},
		{"single item", "file.{js}", []string{"file.{js}"}},
		{"unclosed brace", "file.{js", []string{"file.{js"}},
		{"no closing", "file.{js,ts", []string{"file.{js,ts"}},
		{"just braces", "{}", []string{"{}"}},
		{"just opening", "{", []string{"{"}},
		{"just closing", "}", []string{"}"}},
		{"no braces", "file.js", []string{"file.js"}},

		// Valid expansions with empty alternatives
		{"empty alternatives", "file.{,ts}", []string{"file.", "file.ts"}},
		{"empty alternative middle", "file.{js,,ts}", []string{"file.js", "file.", "file.ts"}},
		{"empty alternative start", "file.{,js,ts}", []string{"file.", "file.js", "file.ts"}},
		{"empty alternative end", "file.{js,ts,}", []string{"file.js", "file.ts", "file."}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandBraces(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandBraces(%q) = %v, want %v",
					tt.input, got, tt.want)
			}
		})
	}
}

func TestMatchWithBraces(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		// Normal cases
		{"simple match js", "*.{js,ts}", "file.js", true},
		{"simple match ts", "*.{js,ts}", "file.ts", true},
		{"simple no match", "*.{js,ts}", "file.go", false},

		// Edge cases - literal brace matches
		{"empty braces", "file.{}", "file.{}", true},
		{"empty braces no match", "file.{}", "file.", false},
		{"single item braces", "file.{js}", "file.{js}", true},
		{"single item no match", "file.{js}", "file.js", false},
		{"unclosed brace", "file.{js", "file.{js", true},
		{"unclosed brace no match", "file.{js", "file.js", false},

		// Valid expansions
		{"braces with star", "*.{js,ts}", "test.js", true},
		{"braces in dir", "{src,lib}/*.js", "src/test.js", true},
		{"braces with globstar", "**/*.{js,ts}", "dir/test.ts", true},
		{"empty alternative", "file.{,js}", "file.", true},
		{"empty alternative 2", "file.{,js}", "file.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v",
					tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
