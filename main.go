package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appVer string = ""

type Config struct {
	Version          string   `mapstructure:"version"`
	TargetURL        string   `mapstructure:"target_url"`
	TargetPath       string   `mapstructure:"target_path"`
	TargetZip        string   `mapstructure:"target_zip"` // New field
	Branch           string   `mapstructure:"branch"`
	Tag              string   `mapstructure:"tag"`
	TempDir          string   `mapstructure:"temp_dir"`
	OutputFile       string   `mapstructure:"output_file"`
	ExcludePaths     []string `mapstructure:"exclude_paths"`
	RespectGitignore bool     `mapstructure:"respect_gitignore"`
	DetailedDiff     bool     `mapstructure:"detailed_diff"`
}

type ComparisonResult struct {
	IdenticalFiles  []string
	DifferentFiles  []string
	SourceOnlyFiles []string
	TargetOnlyFiles []string
	ExcludedFiles   []string
	Diffs           map[string]string
}

const defaultConfigFileBase = ".gitparator" // no trailing .yaml or .yml here

func main() {
	var config Config

	rootCmd := &cobra.Command{
		Use:     "gitparator",
		Short:   "Gitparator is a tool for comparing two Git repositories",
		Version: appVersion(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Handle config file loading
			configFile, _ := cmd.Flags().GetString("config")
			configLoadedFromFile := false
			if configFile != "" {
				// Config file explicitly specified
				viper.SetConfigFile(configFile)
				if err := viper.ReadInConfig(); err != nil {
					// Return error if specified config file cannot be loaded
					return fmt.Errorf("error reading config: %w", err)
				}
				configLoadedFromFile = true
			} else {
				// Try default config file
				viper.SetConfigName(defaultConfigFileBase)
				viper.SetConfigType("yaml")
				viper.AddConfigPath(".")

				// Silently ignore if default config file is not found
				if err := viper.ReadInConfig(); err != nil {
					if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
						// Return error only if it's not a file-not-found error
						return fmt.Errorf("error reading default config file: %w", err)
					}
					// Default config file not found - proceed silently
				} else {
					configLoadedFromFile = true
				}
			}

			if configLoadedFromFile {
				fmt.Println("Using config file:", viper.ConfigFileUsed())

				// Unmarshal config
				if err := viper.Unmarshal(&config); err != nil {
					return fmt.Errorf("failed to parse config file: %w", err)
				}

				err := checkConfigVersion(config.Version)
				if err != nil {
					return fmt.Errorf("configuration file version error: %w", err)
				}

			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runMain(&config)
		},
	}

	// Define flags and configuration settings
	rootCmd.Flags().StringP("config", "c", "", fmt.Sprintf("config file (default is %s.yaml in current directory)", defaultConfigFileBase))
	rootCmd.Flags().StringP("target-url", "u", "", "URL of the target repository")
	rootCmd.Flags().StringP("target-path", "p", "", "Path to the target repository")
	rootCmd.Flags().StringP("target-zip", "z", "", "Path to the zipped target repository")
	rootCmd.Flags().StringP("branch", "b", "", "Branch to compare (ignored if --target-path or --target-zip is specified)")
	rootCmd.Flags().StringP("tag", "t", "", "Tag to compare (ignored if --target-path or --target-zip is specified)")
	rootCmd.Flags().StringP("temp-dir", "", ".gitparator_temp", "Temporary directory for cloning (ignored if --target-path or --target-zip is specified)")
	rootCmd.Flags().StringP("output-file", "o", "report.html", "Output report file")
	rootCmd.Flags().StringSliceP("exclude-paths", "e", []string{}, "Paths to exclude")
	rootCmd.Flags().BoolP("respect-gitignore", "", true, "Respect .gitignore rules")
	rootCmd.Flags().BoolP("detailed-diff", "d", false, "Generate detailed diffs for differing files")

	// Bind flags with viper
	viper.BindPFlag("target_url", rootCmd.Flags().Lookup("target-url"))
	viper.BindPFlag("target_path", rootCmd.Flags().Lookup("target-path"))
	viper.BindPFlag("target_zip", rootCmd.Flags().Lookup("target-zip")) // New binding
	viper.BindPFlag("branch", rootCmd.Flags().Lookup("branch"))
	viper.BindPFlag("tag", rootCmd.Flags().Lookup("tag"))
	viper.BindPFlag("temp_dir", rootCmd.Flags().Lookup("temp-dir"))
	viper.BindPFlag("output_file", rootCmd.Flags().Lookup("output-file"))
	viper.BindPFlag("exclude_paths", rootCmd.Flags().Lookup("exclude-paths"))
	viper.BindPFlag("respect_gitignore", rootCmd.Flags().Lookup("respect-gitignore"))
	viper.BindPFlag("detailed_diff", rootCmd.Flags().Lookup("detailed-diff"))

	// Execute the command once
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkConfigVersion(configVersion string) error {
	if configVersion == "" {
		return fmt.Errorf("configuration file does not specify a version")
	}

	appVerStr := appVersion()
	if appVerStr == "#UNAVAILABLE" {
		// we are running an unversioned build, so we can't check the version
		return nil
	}

	appSemVer, err := semver.ParseTolerant(appVerStr)
	if err != nil {
		return fmt.Errorf("invalid application version: %v", err)
	}

	if configVersion == "" {
		return fmt.Errorf("configuration file does not specify a version")
	}

	configFileVersion, err := semver.ParseTolerant(configVersion)
	if err != nil {
		return fmt.Errorf("configuration file version is not a valid semantic version: %v", err)
	}

	if configFileVersion.Compare(appSemVer) > 0 {
		return fmt.Errorf("configuration file requires gitparator version >= %s", configFileVersion)
	}

	return nil
}

func runMain(config *Config) {
	// Validate required configurations
	if config.TargetZip != "" {
		// TargetZip is specified, use the zip file as the target repository
		if config.TargetURL != "" || config.TargetPath != "" {
			fmt.Println("Error: Only one of --target-url, --target-path, or --target-zip should be specified.")
			os.Exit(1)
		}
		if config.Branch != "" || config.Tag != "" {
			fmt.Println("Warning: --branch and --tag options are ignored when --target-zip is specified.")
		}
		if _, err := os.Stat(config.TargetZip); os.IsNotExist(err) {
			fmt.Printf("Error: target zip file '%s' does not exist.\n", config.TargetZip)
			os.Exit(1)
		}

		// Compare repositories
		result := compareWithZip(".", config.TargetZip, config)

		// Generate HTML report
		if err := generateHTMLReport(result, config.OutputFile); err != nil {
			log.Fatalf("Error generating HTML report: %v", err)
		}

		fmt.Printf("Comparison complete. Report generated as %s\n", config.OutputFile)
		return
	} else if config.TargetPath != "" {
		// TargetPath is specified, use the local directory
		if config.TargetURL != "" {
			fmt.Println("Error: Only one of --target-url, --target-path, or --target-zip should be specified.")
			os.Exit(1)
		}
		if config.Branch != "" || config.Tag != "" {
			fmt.Println("Warning: --branch and --tag options are ignored when --target-path is specified.")
		}
		if _, err := os.Stat(config.TargetPath); os.IsNotExist(err) {
			fmt.Printf("Error: target path '%s' does not exist.\n", config.TargetPath)
			os.Exit(1)
		}

		// Compare repositories
		result := compareRepos(".", config.TargetPath, config)

		// Generate HTML report
		if err := generateHTMLReport(result, config.OutputFile); err != nil {
			log.Fatalf("Error generating HTML report: %v", err)
		}

		fmt.Printf("Comparison complete. Report generated as %s\n", config.OutputFile)
		return
	} else if config.TargetURL != "" {
		// TargetURL is specified, clone the repository
		if config.TempDir == "" {
			config.TempDir = "gitparator_temp"
		}
		targetDir := config.TempDir
		if err := cloneRepo(config, targetDir); err != nil {
			log.Fatalf("Error cloning target repository: %v", err)
		}
		defer os.RemoveAll(targetDir)

		// Compare repositories
		result := compareRepos(".", targetDir, config)

		// Generate HTML report
		if err := generateHTMLReport(result, config.OutputFile); err != nil {
			log.Fatalf("Error generating HTML report: %v", err)
		}

		fmt.Printf("Comparison complete. Report generated as %s\n", config.OutputFile)
		return
	} else {
		fmt.Println("Error: one of --target-url, --target-path, or --target-zip must be specified.")
		os.Exit(1)
	}
}

func cloneRepo(config *Config, targetDir string) error {
	cloneOptions := &git.CloneOptions{
		URL:          config.TargetURL,
		Depth:        1, // Shallow clone
		SingleBranch: true,
	}

	if config.Branch != "" {
		cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(config.Branch)
	} else if config.Tag != "" {
		cloneOptions.ReferenceName = plumbing.NewTagReferenceName(config.Tag)
	}

	_, err := git.PlainClone(targetDir, false, cloneOptions)
	return err
}

func compareRepos(sourceDir, targetDir string, config *Config) ComparisonResult {
	result := ComparisonResult{
		Diffs: make(map[string]string),
	}

	sourceFiles := getAllFilesFromDir(sourceDir, config.ExcludePaths, config.RespectGitignore)
	targetFiles := getAllFilesFromDir(targetDir, config.ExcludePaths, config.RespectGitignore)

	compareFileLists(sourceFiles, targetFiles, sourceDir, targetDir, config, &result)

	return result
}

func compareWithZip(sourceDir, zipPath string, config *Config) ComparisonResult {
	result := ComparisonResult{
		Diffs: make(map[string]string),
	}

	sourceFiles := getAllFilesFromDir(sourceDir, config.ExcludePaths, config.RespectGitignore)
	targetFiles := getAllFilesFromZip(zipPath, config.ExcludePaths, config.RespectGitignore)

	compareFileLists(sourceFiles, targetFiles, sourceDir, zipPath, config, &result)

	return result
}

func compareFileLists(sourceFiles, targetFiles []string, sourceDir, targetDir string, config *Config, result *ComparisonResult) {
	sourceMap := make(map[string]string)
	targetMap := make(map[string]string)

	for _, file := range sourceFiles {
		relativePath, err := filepath.Rel(sourceDir, file)
		if err != nil {
			log.Printf("Error getting relative path for %s: %v", file, err)
			continue
		}
		sourceMap[relativePath] = file
	}

	for _, file := range targetFiles {
		relativePath, err := filepath.Rel(targetDir, file)
		if err != nil {
			log.Printf("Error getting relative path for %s: %v", file, err)
			continue
		}
		targetMap[relativePath] = file
	}

	for path, sourceFile := range sourceMap {
		if targetFile, exists := targetMap[path]; exists {
			if filesAreEqual(sourceFile, targetFile) {
				result.IdenticalFiles = append(result.IdenticalFiles, path)
			} else {
				result.DifferentFiles = append(result.DifferentFiles, path)
				if config.DetailedDiff {
					diff := getFileDiff(sourceFile, targetFile)
					result.Diffs[path] = diff
				}
			}
			delete(targetMap, path)
		} else {
			result.SourceOnlyFiles = append(result.SourceOnlyFiles, path)
		}
	}

	for path := range targetMap {
		result.TargetOnlyFiles = append(result.TargetOnlyFiles, path)
	}

	// Sort all slices for consistent output
	sort.Strings(result.IdenticalFiles)
	sort.Strings(result.DifferentFiles)
	sort.Strings(result.SourceOnlyFiles)
	sort.Strings(result.TargetOnlyFiles)
}

func getAllFilesFromDir(dir string, excludePaths []string, respectGitignore bool) []string {
	var files []string
	var gitignorePatterns []string

	if respectGitignore {
		gitignorePatterns = parseGitignore(dir)
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath := strings.TrimPrefix(path, dir)
		if relativePath == "" {
			return nil
		}
		relativePath = strings.TrimPrefix(relativePath, string(os.PathSeparator))

		if shouldExclude(relativePath, excludePaths) || shouldExclude(relativePath, gitignorePatterns) {
			return nil
		}

		if !info.IsDir() && !strings.Contains(path, ".git"+string(os.PathSeparator)) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking through files: %v", err)
	}
	return files
}

func getAllFilesFromZip(zipPath string, excludePaths []string, respectGitignore bool) []string {
	var files []string
	var gitignorePatterns []string

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Fatalf("Error opening zip file: %v", err)
	}
	defer r.Close()

	if respectGitignore {
		gitignorePatterns = parseGitignoreFromZip(r)
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		relativePath := f.Name

		if shouldExclude(relativePath, excludePaths) || shouldExclude(relativePath, gitignorePatterns) {
			continue
		}

		files = append(files, f.Name)
	}
	return files
}

func parseGitignore(dir string) []string {
	var patterns []string
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return patterns
	}

	file, err := os.Open(gitignorePath)
	if err != nil {
		return patterns
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}

	return patterns
}

func parseGitignoreFromZip(r *zip.ReadCloser) []string {
	var patterns []string
	for _, f := range r.File {
		if f.Name == ".gitignore" {
			rc, err := f.Open()
			if err != nil {
				return patterns
			}
			defer rc.Close()

			scanner := bufio.NewScanner(rc)
			for scanner.Scan() {
				line := scanner.Text()
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				patterns = append(patterns, line)
			}
			break
		}
	}
	return patterns
}

func shouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, _ := doublestar.PathMatch(pattern, path)
		if matched {
			return true
		}
	}
	return false
}

func filesAreEqual(file1, file2 string) bool {
	var content1, content2 []byte
	var err1, err2 error

	if strings.HasSuffix(file1, ".zip") {
		// Read from zip file
		content1, err1 = readFileFromZip(file1)
	} else {
		content1, err1 = os.ReadFile(file1)
	}

	if strings.HasSuffix(file2, ".zip") {
		// Read from zip file
		content2, err2 = readFileFromZip(file2)
	} else {
		content2, err2 = os.ReadFile(file2)
	}

	if err1 != nil || err2 != nil {
		return false
	}

	return string(content1) == string(content2)
}

func readFileFromZip(zipFilePath string) ([]byte, error) {
	// Extract the zip path and the file inside the zip
	zipPath, filePath := splitZipPath(zipFilePath)

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == filePath {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("file %s not found in zip archive", filePath)
}

func splitZipPath(zipFilePath string) (zipPath, filePath string) {
	// For simplicity, assume that zipFilePath is in the format "zipfile.zip::filepath"
	parts := strings.SplitN(zipFilePath, "::", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func getFileDiff(file1, file2 string) string {
	var content1, content2 []byte
	var err1, err2 error

	if strings.HasSuffix(file1, ".zip") {
		content1, err1 = readFileFromZip(file1)
	} else {
		content1, err1 = os.ReadFile(file1)
	}

	if strings.HasSuffix(file2, ".zip") {
		content2, err2 = readFileFromZip(file2)
	} else {
		content2, err2 = os.ReadFile(file2)
	}

	if err1 != nil || err2 != nil {
		return "Error reading files for diff"
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(content1), string(content2), false)
	diffHTML := dmp.DiffPrettyHtml(diffs)

	// Wrap in a div for styling
	return fmt.Sprintf("<div class=\"diff-content\">%s</div>", diffHTML)
}

func generateHTMLReport(result ComparisonResult, outputFile string) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Gitparator Report</title>
    <style>
        body { font-family: Arial, sans-serif; background-color: #f8f9fa; margin: 20px; }
        h1 { color: #343a40; }
        h2 { color: #495057; }
        ul { list-style-type: none; padding: 0; }
        li { padding: 5px; }
        .identical { color: #28a745; }
        .different { color: #dc3545; }
        .source-only { color: #007bff; }
        .target-only { color: #fd7e14; }
        .diff-content { background-color: #f1f1f1; padding: 10px; margin-top: 5px; border-radius: 5px; overflow-x: auto; }
        .diff-deleted { background-color: #ffe6e6; }
        .diff-inserted { background-color: #e6ffe6; }
        pre { white-space: pre-wrap; word-wrap: break-word; }
    </style>
</head>
<body>
    <h1>Gitparator Comparison Report</h1>
    <h2>Identical Files</h2>
    <ul>
        {{- range .IdenticalFiles}}
        <li class="identical">{{.}}</li>
        {{- end}}
    </ul>
    <h2>Different Files</h2>
    <ul>
        {{- range .DifferentFiles}}
        <li class="different">{{.}}
            {{- if (index $.Diffs .)}}
                {{index $.Diffs .}}
            {{- end}}
        </li>
        {{- end}}
    </ul>
    <h2>Files Only in Source Repository</h2>
    <ul>
        {{- range .SourceOnlyFiles}}
        <li class="source-only">{{.}}</li>
        {{- end}}
    </ul>
    <h2>Files Only in Target Repository</h2>
    <ul>
        {{- range .TargetOnlyFiles}}
        <li class="target-only">{{.}}</li>
        {{- end}}
    </ul>
</body>
</html>
`
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, result)
}

// Non-essential utilities moved to the end of the file
func appVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
		// Installed with go install
		return info.Main.Version
	} else if appVer != "" {
		// Built with ldflags
		return appVer
	} else {
		return "#UNAVAILABLE"
	}
}
