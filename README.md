# Gitparator 

Gitparator is a command-line tool for comparing the content and structure of two Git repositories. It helps you identify differences between repositories, including files that are identical, different, or exist only in one of the repositories. Gitparator is particularly useful for validating repository synchronization, checking fork differences, and ensuring content consistency across related repositories.

## Features 
 
- **File-by-File Comparison** : Compares files between two repositories, identifying identical files, differing files, and unique files in each repository.
 
- **Respects `.gitignore` Rules** : Optionally respects `.gitignore` files to exclude irrelevant files from the comparison.
 
- **Exclude Specific Paths** : Allows you to specify files or directories to exclude from the comparison.
 
- **Detailed Diffs** : Generates detailed diffs for differing files, which are included in the HTML report.
 
- **HTML Report Generation** : Produces a visually appealing HTML report with the comparison results.
 
- **Configurable via CLI and Config File** : Supports configuration through both command-line flags and an optional configuration file.
 
- **Compare with Local Repositories** : Allows comparing with a target repository located on the local filesystem.

## Installation 


```shell
go install github.com/adnsv/gitparator@latest
```

## Usage 


```shell
gitparator [flags]
```

### Basic Example 

Compare the current repository with a target repository:


```shell
gitparator --target-url https://github.com/username/target-repo.git --branch main --detailed-diff
```

### Compare with a Local Target Repository 


```shell
gitparator --target-path /path/to/local/target-repo --detailed-diff
```

### Using a Configuration File 
Create a configuration file named `.gitparator.yaml` in the current directory:

```yaml
version: ">=1.0.0"
target_url: 'https://github.com/username/target-repo.git'
branch: 'main'
exclude_paths:
  - 'logs/**'
  - '*.tmp'
respect_gitignore: true
detailed_diff: true
```

Run Gitparator:


```shell
gitparator
```

## Configuration File 
Gitparator supports an optional configuration file in YAML format. By default, it looks for a file named `.gitparator.yaml` in the current working directory. You can specify a different configuration file using the `--config` flag.**Important:**  The configuration file must include a `version` field specifying the compatible version(s) of Gitparator using semantic versioning constraints.
### Configuration Options 
 
- `version` (string, **required** ): Specifies the compatible version(s) of Gitparator. Supports semantic version expressions.
 
- `target_url` (string, optional): URL of the target repository to compare with.
 
- `target_path` (string, optional): Path to the target repository on the local filesystem. If specified, `target_url`, `branch`, and `tag` are ignored.
 
- `branch` (string, optional): Branch to compare. If not specified, defaults to `main` (ignored if `target_path` is specified).
 
- `tag` (string, optional): Tag to compare (ignored if `target_path` is specified).
 
- `temp_dir` (string, optional): Temporary directory for cloning the target repository. Defaults to `gitparator_temp` (ignored if `target_path` is specified).
 
- `output_file` (string, optional): Output report file name. Defaults to `report.html`.
 
- `exclude_paths` (list of strings, optional): Paths or patterns to exclude from the comparison. Supports glob patterns.
 
- `respect_gitignore` (bool, optional): Whether to respect `.gitignore` rules. Defaults to `true`.
 
- `detailed_diff` (bool, optional): Whether to generate detailed diffs for differing files. Defaults to `false`.

### Example Configuration File 


```yaml
# .gitparator.yaml
version: ">=1.0.0"
target_path: '/path/to/local/target-repo'
# target_url: 'https://github.com/username/target-repo.git' # Ignored when target_path is specified
branch: 'develop'  # Ignored when target_path is specified
temp_dir: 'temp_clone'  # Ignored when target_path is specified
output_file: 'comparison_report.html'
exclude_paths:
  - 'logs/**'
  - '*.tmp'
  - 'node_modules/**'
respect_gitignore: true
detailed_diff: true
```

### Notes on Configuration Options 
 
- **`version`** : Uses semantic versioning constraints to specify compatible versions of Gitparator. For example, `">=1.0.0"`.
 
- **`target_path`** : Path to a local Git repository. If specified, `target_url`, `branch`, `tag`, and `temp_dir` are ignored.
 
- **`target_url`** : The URL of the Git repository you want to compare against.
 
- **`branch` and `tag`** : Specify either a branch or a tag to compare. If both are provided, the tag takes precedence.
 
- **`exclude_paths`** : Supports glob patterns. For example, `logs/**` excludes all files and folders within the `logs` directory.
 
- **`respect_gitignore`** : When set to `true`, Gitparator will read the `.gitignore` file in the source repository and exclude those paths from the comparison.
 
- **`detailed_diff`** : When enabled, Gitparator will generate diffs for files that differ and include them in the HTML report.

## Examples 

### Compare with a Specific Branch 


```shell
gitparator --target-url https://github.com/username/target-repo.git --branch develop
```

### Compare with a Specific Tag 


```shell
gitparator --target-url https://github.com/username/target-repo.git --tag v1.2.3
```

### Exclude Specific Paths 


```shell
gitparator --exclude-paths 'docs/**' --exclude-paths '*.md'
```

### Generate Detailed Diffs 


```shell
gitparator --detailed-diff
```

### Specify Output File 


```shell
gitparator --output-file my_report.html
```

### Use a Custom Configuration File 


```shell
gitparator --config /path/to/myconfig.yaml
```

### View Application Version 


```shell
gitparator --version
```

## Flags and Options 
 
- `-u, --target-url` (string): URL of the target repository.
 
- `-p, --target-path` (string): Path to the target repository on the local filesystem.
 
- `-b, --branch` (string): Branch to compare (default is `main`, ignored if `--target-path` is specified).
 
- `-t, --tag` (string): Tag to compare (ignored if `--target-path` is specified).
 
- `--temp-dir` (string): Temporary directory for cloning (default is `gitparator_temp`, ignored if `--target-path` is specified).
 
- `-o, --output-file` (string): Output report file (default is `report.html`).
 
- `-e, --exclude-paths` (string array): Paths to exclude; supports multiple entries.
 
- `--respect-gitignore` (bool): Respect `.gitignore` rules (default is `true`).
 
- `-d, --detailed-diff` (bool): Generate detailed diffs for differing files (default is `false`).
 
- `-c, --config` (string): Path to configuration file (default is `.gitparator.yaml` in current directory).
 
- `--version`: Display application version.
 
- `-h, --help`: Display help information.

## License 
This project is licensed under the MIT License. See the [LICENSE]()  file for details.


## Additional Notes 

### Version Compatibility 
 
- **Version Field in Configuration** : The `version` field in the configuration file is required and ensures that the configuration is compatible with the version of Gitparator you are running.
 
- **Semantic Versioning** : Gitparator uses semantic versioning (SemVer) for version numbers. The application uses the `github.com/blang/semver/v4` package for parsing and comparing versions.
 
- **Example Version Expression** : 
  - `">=1.0.0"`
 
- **Error Handling** : If the application version does not satisfy the version constraint specified in the configuration file, Gitparator will display an error and exit.
