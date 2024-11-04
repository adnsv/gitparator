# gitignore

Package gitignore implements a stack-based .gitignore pattern processor that follows Git's ignore rules.

## Features

- Stack-based pattern management for multiple .gitignore files
- Full support for gitignore pattern syntax
- Proper pattern precedence handling
- Path normalization and base path resolution
- Support for negated patterns
- Directory-specific pattern handling

## Usage

```go
import "github.com/yourusername/gitparator/gitignore"

// Create a new stack with base path
stack := gitignore.NewGitignoreStack("/project/root")

// Add patterns from root .gitignore
stack.PushPatterns([]string{
    "*.log",
    "!important.log",
    "build/",
    "node_modules/"
})

// Add patterns from subdirectory .gitignore
stack.PushPatterns([]string{
    "*.tmp",
    "!debug.log"
})

// Check if files should be ignored
stack.ShouldIgnore("/project/root/test.log")      // true
stack.ShouldIgnore("/project/root/important.log") // false
stack.ShouldIgnore("/project/root/build/out.txt") // true

// Remove subdirectory patterns
stack.PopPatterns()
```

## Pattern Precedence Rules

1. Patterns in deeper directories take precedence over patterns in parent directories
2. Within the same .gitignore file:
   - Later patterns take precedence over earlier patterns
   - Negated patterns (`!pattern`) override previous patterns
3. Directory-specific patterns (ending in `/`) only match directories
4. Patterns without slashes match in any directory

## API Reference

### Types

```go
type GitignoreStack struct {
    // contains filtered or unexported fields
}
```

### Functions

#### NewGitignoreStack
```go
func NewGitignoreStack(basePath string) *GitignoreStack
```
Creates a new GitignoreStack with the specified base path.

#### PushPatterns
```go
func (gs *GitignoreStack) PushPatterns(patterns []string)
```
Adds a new group of patterns to the top of the stack.

#### PopPatterns
```go
func (gs *GitignoreStack) PopPatterns()
```
Removes the most recently added group of patterns.

#### ShouldIgnore
```go
func (gs *GitignoreStack) ShouldIgnore(path string) bool
```
Checks if a given path should be ignored.

## Pattern Syntax

- `*` - matches any sequence of characters except slash
- `?` - matches any single character except slash
- `**` - matches zero or more directories
- `[abc]` - matches any character in brackets
- `[a-z]` - matches any character in the range
- `[!abc]` or `[^abc]` - matches any character not in brackets
- `/pattern` - matches from the project root
- `pattern/` - matches directories
- `!pattern` - negates a pattern

## Examples

### Basic Usage
```go
stack := gitignore.NewGitignoreStack("/project")

// Root .gitignore patterns
stack.PushPatterns([]string{
    "*.log",
    "build/",
    "!important.log"
})

fmt.Println(stack.ShouldIgnore("/project/debug.log"))     // true
fmt.Println(stack.ShouldIgnore("/project/important.log")) // false
fmt.Println(stack.ShouldIgnore("/project/build/out.txt")) // true
```

### Multiple Pattern Levels
```go
// Root patterns
stack.PushPatterns([]string{
    "*.log",
    "temp/"
})

// Subdirectory patterns
stack.PushPatterns([]string{
    "!debug.log",
    "*.tmp"
})

fmt.Println(stack.ShouldIgnore("/project/error.log"))  // true
fmt.Println(stack.ShouldIgnore("/project/debug.log"))  // false
fmt.Println(stack.ShouldIgnore("/project/test.tmp"))   // true
```

### Directory-Specific Patterns
```go
stack.PushPatterns([]string{
    "logs/",          // matches directory
    "*.log",          // matches files
    "build/*.out",    // matches in specific directory
    "**/temp/",       // matches in any directory
})
```

## Implementation Notes

- All paths are normalized to use forward slashes
- Relative paths are resolved against the base path
- Empty lines and comments are ignored
- Pattern groups maintain Git's precedence rules
- Directory patterns (ending in `/`) are handled specially
- Patterns are processed from most specific (last) to least specific (first)
