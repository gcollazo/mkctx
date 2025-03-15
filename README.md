# mkctx

> Generate structured context from your codebase for LLMs like Claude

`mkctx` is a lightweight CLI tool that prepares your code for AI interactions by creating a formatted directory tree and
extracting file contents in a clean, structured format.


> ✨ Vibe Coding Certified™ 🧙‍♂️💻


## Installation

Visit the [releases page](https://github.com/gcollazo/mkctx/releases) to download the latest version for your platform.

```bash
# macOS (Apple Silicon) - replace X.Y.Z with the latest version
curl -L https://github.com/gcollazo/mkctx/releases/download/X.Y.Z/mkctx-X.Y.Z-darwin-arm64.tar.gz | tar xz && sudo mv mkctx /usr/local/bin/

# macOS (Intel) - replace X.Y.Z with the latest version
curl -L https://github.com/gcollazo/mkctx/releases/download/X.Y.Z/mkctx-X.Y.Z-darwin-amd64.tar.gz | tar xz && sudo mv mkctx /usr/local/bin/

# Linux - replace X.Y.Z with the latest version
curl -L https://github.com/gcollazo/mkctx/releases/download/X.Y.Z/mkctx-X.Y.Z-linux-amd64.tar.gz | tar xz && sudo mv mkctx /usr/local/bin/

# Windows (PowerShell) - replace X.Y.Z with the latest version
Invoke-WebRequest -Uri https://github.com/gcollazo/mkctx/releases/download/X.Y.Z/mkctx-X.Y.Z-windows-amd64.zip -OutFile mkctx.zip
Expand-Archive mkctx.zip -DestinationPath .
# Move mkctx.exe to a directory in your PATH
```

## Basic Usage

```bash
# Process current directory
mkctx . > context.md

# Focus on specific file types
mkctx --include "*.go" . > context.md

# Exclude specific directories
mkctx --exclude "node_modules/*" . > context.md

# Respect .gitignore patterns
mkctx --gitignore . > context.md
```

## Features

- 📂 Creates visual directory tree
- 📄 Extracts content from all non-binary files
- 🔍 Smart filtering (include/exclude patterns)
- 🚫 Auto-excludes binary files and sensitive content
- 🔒 Protects environment files (`.env`) by default
- 📝 Adds custom LLM instructions via `.mkctx` file

## File Filtering Options

### Include Only Specific Files

```bash
# Only Go files
mkctx --include "*.go" .

# Go files and Markdown files
mkctx --include "*.go" --include "*.md" .
```

### Exclude Files or Directories

```bash
# Skip the vendor directory
mkctx --exclude "vendor/*" .

# Skip tests and vendor
mkctx --exclude "*_test.go" --exclude "vendor/*" .
```

### Use .gitignore Patterns

```bash
# Respect patterns from .gitignore
mkctx --gitignore .
```

### Combine Approaches

```bash
# Only Go files, excluding tests, respecting gitignore
mkctx --include "*.go" --exclude "*_test.go" --gitignore .
```

## The `.mkctx` File

Create a `.mkctx` file in your project root to provide instructions for the LLM. Its contents will appear in the output
as a special "USER INSTRUCTIONS" section.

```
You are a code reviewer analyzing this repository. Focus on:
1. Potential security vulnerabilities
2. Performance optimizations
3. Code quality improvements

Provide specific examples when suggesting changes.
```

## Output Format

The generated output follows this structure:

    # Directory Structure
    ```
    └── project/
    ├── main.go
    ├── utils/
    │   └── helpers.go
    └── README.md
    ```

    # Source Code Files

    ## main.go
    ```go
    package main

    import "fmt"

    func main() {
        fmt.Println("Hello World")
    }
    ```

    ## USER INSTRUCTIONS

    ```
    [Contents of your .mkctx file]
    ```

## Advanced Usage

### Process Specific Subdirectories

```bash
# Focus on a particular component
mkctx --include "internal/auth/*.go" .
```

### Piping to LLMs

```bash
# Send to Claude via API
mkctx . | curl -X POST https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [{"role": "user", "content": "'"$(cat)"'"}]
  }'
```

## Usage Tips

1. **Start minimal** - Begin with only the most relevant files
2. **Use specific patterns** - Target just what you need for your question
3. **Create a `.mkctx` file** - Provide consistent instructions for analysis
4. **Use `--gitignore`** - Leverage your existing exclusion patterns
5. **Exclude generated code** - Skip auto-generated files that add noise

## Contributing

| Contribution        | Policy         |
| ------------------- | -------------- |
| 🐛 Bug reports      | Yes please! ✅ |
| ✨ Feature requests | No thanks 🙅‍♂️   |
| 👨‍💻 Code             | No thanks 🙅‍♂️   |

## License

MIT © 2025 Giovanni Collazo
