# wildpath

Package wildpath provides advanced path pattern matching with support for gitignore-style patterns and brace expansion.

## Features

- Single-star (`*`) and question-mark (`?`) matching
- Double-star (`**`) matching for zero or more directories
- Range matching (`[a-z]`, `[abc]`, `[!0-9]`, etc.)
- Brace expansion (`{js,ts}`)
- Syntax support compatible with gitignore

## Pattern Syntax

- `*` - matches any sequence of characters within a path component
- `?` - matches any single character
- `**` - matches zero or more directories
- `[abc]` - matches any character in brackets
- `[a-z]` - matches any character in the range
- `[!abc]` or `[^abc]` - matches any character not in brackets
- `{js,ts}` - matches any of the comma-separated patterns
- Leading `/` - makes the pattern root-relative

## Usage

```go
import "github.com/yourusername/gitparator/wildpath"

// Simple pattern matching
matched := wildpath.Match("*.txt", "file.txt")                  // true
matched = wildpath.Match("src/**/*.go", "src/pkg/main.go")      // true

// Character ranges
matched = wildpath.Match("[a-z]*.txt", "test.txt")             // true
matched = wildpath.Match("[!0-9]*.txt", "test.txt")            // true

// Brace expansion
matched = wildpath.Match("*.{js,ts}", "module.js")             // true
matched = wildpath.Match("lib/*.{js,ts}", "lib/utils.ts")      // true

// Root-relative patterns
matched = wildpath.Match("/root/*.txt", "/root/file.txt")      // true
matched = wildpath.Match("/root/*.txt", "other/file.txt")      // false
```

## Pattern Matching Rules

1. Path components are separated by forward slashes (`/`)
2. `*` matches any sequence of characters within a single path component
3. `?` matches exactly one character
4. `**` matches zero or more path components
5. Character ranges support:
   - Single characters: `[abc]`
   - Ranges: `[a-z]`
   - Negation: `[!abc]` or `[^abc]`
6. Brace expansion creates multiple patterns:
   - `{js,ts}` expands to two patterns
7. Leading slash makes pattern root-relative
8. Paths are normalized (consecutive slashes removed)

## Notes

- All paths use forward slashes, regardless of OS
- Empty patterns match only empty paths
- Unclosed brackets/braces are treated as literals
- Pattern matching is case-sensitive
- Root-relative patterns must match exactly
