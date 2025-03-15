# mkctx - Context Generator for LLMs

`mkctx` is a command-line tool that generates structured context from source code repositories for use with large language models (LLMs) like Claude. It extracts code and visualizes the directory structure in a format optimized for AI context windows.

## Quick Installation

### macOS
    curl -L https://github.com/gcollazo/mkctx/releases/download/25.03.1/mkctx-25.03.1-darwin-arm64.tar.gz | tar xz && sudo mv mkctx /usr/local/bin/

### Linux
    curl -L https://github.com/gcollazo/mkctx/releases/download/25.03.1/mkctx-25.03.1-linux-amd64.tar.gz | tar xz && sudo mv mkctx /usr/local/bin/

### Windows
    # Download in PowerShell
    Invoke-WebRequest -Uri https://github.com/gcollazo/mkctx/releases/download/25.03.1/mkctx-25.03.1-windows-amd64.zip -OutFile mkctx.zip
    Expand-Archive mkctx.zip -DestinationPath .
    # Move to a directory in your PATH

## Command Help

    mkctx - Context Generator for LLMs
    
    USAGE:
      mkctx [OPTIONS] [DIRECTORY]
    
    ARGUMENTS:
      DIRECTORY    Path to the directory to process (required unless --help or --version is specified)
    
    OPTIONS:
      --include PATTERN    Include only files matching the glob pattern (can be used multiple times)
      --exclude PATTERN    Exclude files matching the glob pattern (can be used multiple times)
      --gitignore          Respect patterns from .gitignore file
      --version            Show version information
      --help               Show this help message
    
    EXAMPLES:
      # Process all files in the current directory
      mkctx .
    
      # Include only Go files
      mkctx --include "*.go" /path/to/project

## Sample Output

Below is a sample output from running `mkctx` on a small project:

    # Directory Structure
    
    └── project/
        ├── main.go
        ├── utils/
        │   └── helpers.go
        └── README.md
    
    # Source Code Files
    
    ## main.go
    
    package main
    
    import (
        "fmt"
        "./utils"
    )
    
    func main() {
        fmt.Println("Hello, world!")
        utils.PrintHelper()
    }
    
    ## utils/helpers.go
    
    package utils
    
    import "fmt"
    
    func PrintHelper() {
        fmt.Println("This is a helper function")
    }
    
    ## README.md
    
    # Sample Project
    
    This is a sample project to demonstrate mkctx functionality.
    
    # USER INSTRUCTIONS
    
    ```
    You are a software engineer reviewing this code. Please suggest improvements
    and identify any potential bugs or security issues.
    ```

## Features

- Creates a directory tree visualization of your project
- Extracts the content of source code files
- Formats everything in a clean, structured way for LLMs
- Intelligently filters files using various criteria:
  - Include only specific file types
  - Exclude specific files or directories
  - Respect gitignore patterns
- Automatically excludes binary files
- No external dependencies - just pure Go
- Supports custom LLM instructions via the `.mkctx` file

## The .mkctx File

The `.mkctx` file is a special file that can be placed in the root of your project directory. This file contains instructions or context for the LLM to follow when analyzing your code. When `mkctx` finds this file, it appends its contents to the output as a special section titled "LLM Instructions".

This is particularly useful for:

1. **Providing consistent instructions** across multiple LLM sessions about the same codebase
2. **Setting expectations** for what kind of analysis or feedback you want
3. **Giving context** about the project's purpose, constraints, or requirements
4. **Specifying a persona or role** for the LLM to adopt when analyzing the code

### Example .mkctx File

```
You are a security consultant reviewing this codebase. Focus on:
1. Potential vulnerabilities in file parsing
2. Input validation issues
3. Unsafe operations
4. Suggestions for hardening this utility

Provide specific code examples for any improvements you recommend.
```

## Default Exclusions

By default, `mkctx` excludes certain files and directories for security and relevance:

1. **Binary files** - Files with binary content or binary extensions (images, executables, etc.)
2. **Git internals** - The `.git` directory and its contents
3. **Environment files** - `.env` files and files with `.env` extension are excluded by default to prevent accidental exposure of secrets and credentials
4. **The .mkctx file itself** - This file is processed separately and appended to the output as instructions rather than included in the regular file listing

To include `.env` files, you must explicitly specify them with the `--include` flag:

    # Include only .env files
    mkctx --include ".env" /path/to/project
    
    # Include .env files along with other files
    mkctx --include "*.go" --include ".env" /path/to/project
    
    # Include all .env files in the project
    mkctx --include "*.env" /path/to/project

This protection helps prevent accidentally leaking sensitive configuration and credentials when sharing code context with LLMs.

## Open Source Status

This project is open source under the MIT License. While we welcome bug reports through the GitHub issue tracker, we are not accepting code contributions at this time. The project is being maintained by its original authors who have specific design goals and quality standards.

## Installation

### From Source

    git clone https://github.com/yourusername/mkctx.git
    cd mkctx
    go build -o mkctx

Move the built binary to a location in your PATH:

    mv mkctx /usr/local/bin/   # Linux/macOS

### Cross-platform Binaries

Pre-built binaries for multiple platforms are available in the releases section of the GitHub repository.

## Usage

### Basic Usage

At its simplest, run `mkctx` with a directory path to generate context from that directory:

    mkctx /path/to/your/project > context.md

This will generate a Markdown file with:
1. A directory tree of your project
2. The content of all non-binary files
3. Any instructions from the `.mkctx` file (if it exists)

### Filtering Files

You can control which files are included in the output using various flags:

#### Include Only Specific Files

The `--include` flag allows you to specify which files to include. Everything else will be excluded:

    # Include only Go files
    mkctx --include "*.go" /path/to/your/project > context.md
    
    # Include both Go files and Markdown files
    mkctx --include "*.go" --include "*.md" /path/to/your/project > context.md

#### Exclude Specific Files

The `--exclude` flag allows you to specify which files to exclude:

    # Exclude the vendor directory
    mkctx --exclude "vendor/*" /path/to/your/project > context.md
    
    # Exclude both vendor directory and test files
    mkctx --exclude "vendor/*" --exclude "*_test.go" /path/to/your/project > context.md

#### Using .gitignore Patterns

The `--gitignore` flag tells `mkctx` to respect patterns from your project's `.gitignore` file:

    mkctx --gitignore /path/to/your/project > context.md

#### Combining Filters

You can combine these filters to precisely control which files are included:

    # Include only Go files, exclude vendor directory, and respect gitignore
    mkctx --include "*.go" --exclude "vendor/*" --gitignore /path/to/your/project > context.md

### Filter Order

Filters are applied in the following order:

1. Include filters (if specified)
2. Exclude filters
3. Gitignore patterns

This means that even if a file matches an include pattern, it will be excluded if it also matches an exclude pattern or a gitignore pattern.

## Glob Pattern Syntax

Both `--include` and `--exclude` flags accept glob patterns:

- `*`: Matches any sequence of characters except path separators
- `?`: Matches any single character except path separators
- `[abc]`: Matches any character within the brackets
- `[a-z]`: Matches any character in the range
- `*.go`: Matches all files with the `.go` extension
- `dir/*`: Matches all files within the `dir` directory
- `**`: Matches any number of directories (useful for deep searches)

## Examples

### Small Project Example

For a small Go project, you might want to include all Go files except tests:

    mkctx --include "*.go" --exclude "*_test.go" /path/to/project > context.md

### Large Project Example

For a large project, you might want to focus on a specific component:

    mkctx --include "internal/auth/*.go" --gitignore /path/to/project > context.md

### Multiple File Types

If you're working on both code and documentation:

    mkctx --include "*.go" --include "*.md" --exclude "vendor/*" /path/to/project > context.md

## Piping to LLMs

The primary use case of `mkctx` is to generate context that can be pasted into LLM chat interfaces or used with API calls:

    # Generate context and send to Claude via API
    mkctx /path/to/project | curl -X POST https://api.anthropic.com/v1/messages \
      -H "x-api-key: $ANTHROPIC_API_KEY" \
      -H "Content-Type: application/json" \
      -d '{
        "model": "claude-3-opus-20240229",
        "messages": [{"role": "user", "content": "'"$(cat)"'"}]
      }'

## Best Practices

1. **Start with Minimal Context**: Begin with only the most relevant files to keep the context window focused.

2. **Use Specific Include Patterns**: Target only the files needed for your question or task.

3. **Exclude Generated Files**: Always exclude auto-generated code that might clutter the context.

4. **Consider File Size**: Very large files may consume too much of the context window. Consider excluding them or extracting only the relevant portions.

5. **Use --gitignore Flag**: Most repositories have sensible `.gitignore` files that exclude irrelevant directories like `node_modules` or `build` folders.

6. **Create a .mkctx File**: For projects you frequently analyze with LLMs, create a `.mkctx` file with standard instructions to maintain consistency.

## Troubleshooting

### Files Missing from Output

If expected files are not appearing in the output:

1. Check your include/exclude patterns for typos
2. Run with `--include` flag alone to see what files match
3. Remember that patterns are applied from the root of the specified directory

### Too Much Output

If the output is too large for your LLM's context window:

1. Use more specific include patterns
2. Add exclude patterns for large directories
3. Focus on a subdirectory instead of the entire project
4. Use the tool in an iterative manner, focusing on different parts of the codebase in separate prompts

### Binary Files Appearing in Tree but Not in Content

This is expected behavior. Binary files (like images or compiled binaries) are shown in the directory tree for completeness but their content is excluded from the output since they're not useful for LLMs.

## Bug Reports

We welcome bug reports through the GitHub issue tracker. When submitting a bug report, please include:

1. The version of `mkctx` you're using
2. Your operating system