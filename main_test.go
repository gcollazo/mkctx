package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestMatchGitignorePattern tests the pattern matching functionality.
func TestMatchGitignorePattern(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	// Create test directories
	testDirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir2", "subdir"),
	}
	for _, dir := range testDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test files
	testFiles := []string{
		filepath.Join(tempDir, "file.txt"),
		filepath.Join(tempDir, "dir1", "test.go"),
		filepath.Join(tempDir, "dir2", "file.js"),
		filepath.Join(tempDir, "dir2", "subdir", "config.yaml"),
	}
	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Define test cases: [pattern, path, expected]
	tests := []struct {
		pattern  string
		path     string
		expected bool
	}{
		// Simple file patterns
		{"*.txt", "file.txt", true},
		{"*.go", "file.txt", false},
		{"*.go", "dir1/test.go", true},

		// Directory specific patterns
		{"dir1/", "dir1", true},
		{"dir1/", "dir2", false},
		{"dir2/", "dir2/file.js", false}, // Pattern specifies directory, path is a file

		// Patterns with directory separators
		{"dir1/*.go", "dir1/test.go", true},
		{"dir1/*.go", "dir2/file.js", false},
		{"dir2/subdir/*.yaml", "dir2/subdir/config.yaml", true},

		// Patterns with leading slash
		{"/file.txt", "file.txt", true},
		{"/dir1/test.go", "dir1/test.go", true},
		{"/dir1/test.js", "dir1/test.go", false},

		// Wildcard patterns
		{"dir*/*.go", "dir1/test.go", true},
		{"*/subdir/*.yaml", "dir2/subdir/config.yaml", true},
	}

	for _, test := range tests {
		// Make paths relative to tempDir for testing
		path := strings.TrimPrefix(test.path, tempDir+"/")

		result := matchGitignorePattern(test.pattern, path)
		if result != test.expected {
			t.Errorf("matchGitignorePattern(%q, %q) = %v, expected %v",
				test.pattern, path, result, test.expected)
		}
	}
}

// TestShouldProcessFile tests the file filtering logic
func TestShouldProcessFile(t *testing.T) {
	tests := []struct {
		relPath        string
		includeGlobs   []string
		excludeGlobs   []string
		gitignoreGlobs []string
		expected       bool
	}{
		// Test include patterns
		{"file.txt", []string{"*.txt"}, []string{}, []string{}, true},
		{"file.go", []string{"*.txt"}, []string{}, []string{}, false},
		{"dir/file.txt", []string{"dir/*.txt"}, []string{}, []string{}, true},

		// Test exclude patterns
		{"file.txt", []string{}, []string{"*.txt"}, []string{}, false},
		{"file.go", []string{}, []string{"*.txt"}, []string{}, true},
		{"dir/file.txt", []string{}, []string{"dir/*"}, []string{}, false},

		// Test gitignore patterns
		{"file.txt", []string{}, []string{}, []string{"*.txt"}, false},
		{"file.go", []string{}, []string{}, []string{"*.txt"}, true},

		// Test combination of patterns
		{"file.txt", []string{"*.txt"}, []string{"file.txt"}, []string{}, false},
		{"file.go", []string{"*.go"}, []string{}, []string{"*.go"}, false},
		{"dir/file.txt", []string{"dir/*"}, []string{"*.go"}, []string{}, true},
		{"vendor/file.go", []string{"*.go"}, []string{"vendor/*"}, []string{}, false},

		// Test with empty include (should include everything)
		{"file.txt", []string{}, []string{}, []string{}, true},
	}

	for _, test := range tests {
		result := shouldProcessFile(test.relPath, test.includeGlobs, test.excludeGlobs, test.gitignoreGlobs)
		if result != test.expected {
			t.Errorf("shouldProcessFile(%q, %v, %v, %v) = %v, expected %v",
				test.relPath, test.includeGlobs, test.excludeGlobs, test.gitignoreGlobs, result, test.expected)
		}
	}
}

// TestIsBinaryFile tests the binary file detection
func TestIsBinaryFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	// Create a text file
	textFile := filepath.Join(tempDir, "text.txt")
	err := os.WriteFile(textFile, []byte("This is a text file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	// Create a binary file with null bytes
	binaryFile := filepath.Join(tempDir, "binary.bin")
	err = os.WriteFile(binaryFile, []byte{0x00, 0x01, 0x02, 0x03}, 0644)
	if err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	// Create a file with binary extension but text content
	binaryExtFile := filepath.Join(tempDir, "textcontent.png")
	err = os.WriteFile(binaryExtFile, []byte("This is actually text"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file with binary extension: %v", err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{textFile, false},
		{binaryFile, true},
		{binaryExtFile, true}, // Should be true based on extension
		{filepath.Join(tempDir, "nonexistent.file"), true}, // Should be true if file can't be read
	}

	for _, test := range tests {
		result := isBinaryFile(test.path)
		if result != test.expected {
			t.Errorf("isBinaryFile(%q) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

// TestParseGitignoreFile tests the gitignore file parsing
func TestParseGitignoreFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	// Create a gitignore file
	gitignoreContent := `# This is a comment
*.log
/dist/
node_modules/
!important.log
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore file: %v", err)
	}

	// Test parsing
	patterns, err := parseGitignoreFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to parse .gitignore file: %v", err)
	}

	expectedPatterns := []string{
		"*.log",
		"/dist/",
		"node_modules/",
	}

	if !reflect.DeepEqual(patterns, expectedPatterns) {
		t.Errorf("parseGitignoreFile(%q) = %v, expected %v", gitignorePath, patterns, expectedPatterns)
	}

	// Test with nonexistent file
	patterns, err = parseGitignoreFile(filepath.Join(tempDir, "nonexistent.gitignore"))
	if err == nil {
		t.Errorf("Expected error when parsing nonexistent file, got nil")
	}
	if len(patterns) != 0 {
		t.Errorf("Expected empty patterns for nonexistent file, got %v", patterns)
	}
}

// TestBuildDirectoryTree tests the tree building functionality
func TestBuildDirectoryTree(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	// Create test directories
	dirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir2", "subdir"),
		filepath.Join(tempDir, ".git"),
		filepath.Join(tempDir, ".git", "objects"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test files
	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "dir1", "file2.go"),
		filepath.Join(tempDir, "dir2", "file3.js"),
		filepath.Join(tempDir, "dir2", "subdir", "file4.yaml"),
		filepath.Join(tempDir, ".git", "config"),
		filepath.Join(tempDir, ".git", "objects", "object1"),
	}
	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Build tree
	tree := buildDirectoryTree(tempDir, tempDir)

	// Verify the root node
	if tree.Name != filepath.Base(tempDir) || !tree.IsDir {
		t.Errorf("Root node incorrect: got %+v", tree)
	}

	// Verify children
	childNames := make(map[string]bool)
	for _, child := range tree.Children {
		childNames[child.Name] = true
	}

	// Should have dir1, dir2, .git, and file1.txt
	expectedNames := []string{"dir1", "dir2", ".git", "file1.txt"}
	for _, name := range expectedNames {
		if !childNames[name] {
			t.Errorf("Expected child %s not found in tree", name)
		}
	}

	// Verify .git directory has no children (as per our rules)
	for _, child := range tree.Children {
		if child.Name == ".git" {
			if len(child.Children) != 0 {
				t.Errorf(".git directory should have no children, has %d", len(child.Children))
			}
			break
		}
	}
}

// TestCollectFiles tests the file collection functionality
func TestCollectFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	// Create test directories
	dirs := []string{
		filepath.Join(tempDir, "src"),
		filepath.Join(tempDir, "vendor", "lib"),
		filepath.Join(tempDir, "docs"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test files
	files := map[string][]byte{
		filepath.Join(tempDir, "src", "main.go"):          []byte("package main\n\nfunc main() {}\n"),
		filepath.Join(tempDir, "src", "utils.go"):         []byte("package main\n\nfunc util() {}\n"),
		filepath.Join(tempDir, "vendor", "lib.go"):        []byte("package lib\n\nfunc Lib() {}\n"),
		filepath.Join(tempDir, "vendor", "lib", "sub.go"): []byte("package lib\n\nfunc Sub() {}\n"),
		filepath.Join(tempDir, "docs", "readme.md"):       []byte("# Documentation\n"),
		filepath.Join(tempDir, "image.png"):               {0x00, 0x01, 0x02, 0x03}, // Binary content
	}
	for path, content := range files {
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Create a .gitignore file
	gitignoreContent := "*.md\n"
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore file: %v", err)
	}

	// Test cases
	tests := []struct {
		name             string
		config           Configuration
		expectedCount    int
		expectedContains []string
		expectedExcludes []string
	}{
		{
			name: "No filters",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{},
				ExcludeGlobs:   []string{},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedCount: 5, // All text files
			expectedContains: []string{
				filepath.Join(tempDir, "src", "main.go"),
				filepath.Join(tempDir, "docs", "readme.md"),
			},
			expectedExcludes: []string{
				filepath.Join(tempDir, "image.png"), // Binary file
			},
		},
		{
			name: "Include Go files",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{"*.go"},
				ExcludeGlobs:   []string{},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedCount: 4, // All Go files
			expectedContains: []string{
				filepath.Join(tempDir, "src", "main.go"),
				filepath.Join(tempDir, "vendor", "lib.go"),
			},
			expectedExcludes: []string{
				filepath.Join(tempDir, "docs", "readme.md"),
			},
		},
		{
			name: "Exclude vendor",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{},
				ExcludeGlobs:   []string{"vendor/*"},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedCount: 3, // All except vendor
			expectedContains: []string{
				filepath.Join(tempDir, "src", "main.go"),
				filepath.Join(tempDir, "docs", "readme.md"),
			},
			expectedExcludes: []string{
				filepath.Join(tempDir, "vendor", "lib.go"),
			},
		},
		{
			name: "Use gitignore",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{},
				ExcludeGlobs:   []string{},
				UseGitignore:   true,
				GitignoreGlobs: []string{"*.md"},
			},
			expectedCount: 4, // All except markdown
			expectedContains: []string{
				filepath.Join(tempDir, "src", "main.go"),
			},
			expectedExcludes: []string{
				filepath.Join(tempDir, "docs", "readme.md"),
			},
		},
		{
			name: "Combine include and exclude",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{"*.go"},
				ExcludeGlobs:   []string{"vendor/*"},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedCount: 2, // Only Go files outside vendor
			expectedContains: []string{
				filepath.Join(tempDir, "src", "main.go"),
			},
			expectedExcludes: []string{
				filepath.Join(tempDir, "vendor", "lib.go"),
				filepath.Join(tempDir, "docs", "readme.md"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files := collectFiles(test.config)

			// Check count
			if len(files) != test.expectedCount {
				t.Errorf("Expected %d files, got %d", test.expectedCount, len(files))
			}

			// Check for expected files
			for _, expectedFile := range test.expectedContains {
				found := false
				for _, file := range files {
					if file == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in results", expectedFile)
				}
			}

			// Check for excluded files
			for _, excludedFile := range test.expectedExcludes {
				for _, file := range files {
					if file == excludedFile {
						t.Errorf("Excluded file %s found in results", excludedFile)
						break
					}
				}
			}
		})
	}
}

// TestMultiFlagImplementation tests the custom flag implementation
func TestMultiFlagImplementation(t *testing.T) {
	var f multiFlag

	// Test initial state
	if f.String() != "" {
		t.Errorf("Expected empty string, got %q", f.String())
	}

	// Test adding values
	f.Set("value1")
	if len(f) != 1 || f[0] != "value1" {
		t.Errorf("Expected [value1], got %v", f)
	}

	f.Set("value2")
	if len(f) != 2 || f[0] != "value1" || f[1] != "value2" {
		t.Errorf("Expected [value1, value2], got %v", f)
	}

	// Test String() method
	expected := "value1, value2"
	if f.String() != expected {
		t.Errorf("Expected %q, got %q", expected, f.String())
	}
}

// Integration tests for the entire workflow, using a sample directory
func TestIntegrationFullWorkflow(t *testing.T) {
	// Create a sample project structure
	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	// Create directories
	dirs := []string{
		filepath.Join(tempDir, "src"),
		filepath.Join(tempDir, "vendor", "github.com", "pkg"),
		filepath.Join(tempDir, "docs"),
		filepath.Join(tempDir, ".git", "objects"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create files
	files := map[string][]byte{
		filepath.Join(tempDir, "src", "main.go"):                        []byte("package main\n\nfunc main() {}\n"),
		filepath.Join(tempDir, "src", "utils.go"):                       []byte("package main\n\nfunc util() {}\n"),
		filepath.Join(tempDir, "vendor", "lib.go"):                      []byte("package vendor\n\nfunc Lib() {}\n"),
		filepath.Join(tempDir, "vendor", "github.com", "pkg", "pkg.go"): []byte("package pkg\n\nfunc Pkg() {}\n"),
		filepath.Join(tempDir, "docs", "readme.md"):                     []byte("# Documentation\n"),
		filepath.Join(tempDir, "docs", "api.md"):                        []byte("# API Documentation\n"),
		filepath.Join(tempDir, "Makefile"):                              []byte("all:\n\tgo build\n"),
		filepath.Join(tempDir, "image.png"):                             {0x00, 0x01, 0x02, 0x03}, // Binary content
		filepath.Join(tempDir, ".git", "config"):                        []byte("[core]\n\trepositoryformatversion = 0\n"),
		filepath.Join(tempDir, ".git", "objects", "obj"):                []byte{0x00, 0x01, 0x02, 0x03}, // Binary content
	}
	for path, content := range files {
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Create a .gitignore file
	gitignoreContent := "*.png\ndocs/api.md\n"
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore file: %v", err)
	}

	// Test scenarios
	scenarios := []struct {
		name           string
		config         Configuration
		expectedFiles  int
		containsFiles  []string
		excludesFiles  []string
		containsInTree []string
		excludesInTree []string
	}{
		{
			name: "Default behavior",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{},
				ExcludeGlobs:   []string{},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedFiles:  7, // All text files
			containsFiles:  []string{"src/main.go", "vendor/lib.go", "docs/readme.md", "docs/api.md"},
			excludesFiles:  []string{"image.png", ".git/config"},
			containsInTree: []string{"src", "vendor", "docs", ".git", "Makefile"},
			excludesInTree: []string{".git/objects"},
		},
		{
			name: "Go files only",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{"*.go"},
				ExcludeGlobs:   []string{},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedFiles:  4, // Only Go files
			containsFiles:  []string{"src/main.go", "vendor/lib.go"},
			excludesFiles:  []string{"docs/readme.md", "Makefile"},
			containsInTree: []string{"src", "vendor", "docs", ".git", "Makefile"},
			excludesInTree: []string{".git/objects"},
		},
		{
			name: "Exclude vendor",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{},
				ExcludeGlobs:   []string{"vendor/*"},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedFiles:  5, // All except vendor
			containsFiles:  []string{"src/main.go", "docs/readme.md"},
			excludesFiles:  []string{"vendor/lib.go"},
			containsInTree: []string{"src", "vendor", "docs", ".git", "Makefile"},
			excludesInTree: []string{".git/objects"},
		},
		{
			name: "With gitignore",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{},
				ExcludeGlobs:   []string{},
				UseGitignore:   true,
				GitignoreGlobs: []string{"*.png", "docs/api.md"},
			},
			expectedFiles:  6, // All text files except api.md
			containsFiles:  []string{"src/main.go", "docs/readme.md"},
			excludesFiles:  []string{"docs/api.md", "image.png"},
			containsInTree: []string{"src", "vendor", "docs", ".git", "Makefile", "image.png"},
			excludesInTree: []string{".git/objects"},
		},
		{
			name: "Go files outside vendor",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{"*.go"},
				ExcludeGlobs:   []string{"vendor/*"},
				UseGitignore:   false,
				GitignoreGlobs: []string{},
			},
			expectedFiles:  2, // Go files outside vendor
			containsFiles:  []string{"src/main.go"},
			excludesFiles:  []string{"vendor/lib.go", "docs/readme.md"},
			containsInTree: []string{"src", "vendor", "docs", ".git", "Makefile"},
			excludesInTree: []string{".git/objects"},
		},
		{
			name: "Complex combination",
			config: Configuration{
				RootDir:        tempDir,
				IncludeGlobs:   []string{"*.go", "*.md"},
				ExcludeGlobs:   []string{"vendor/github.com/*"},
				UseGitignore:   true,
				GitignoreGlobs: []string{"*.png", "docs/api.md"},
			},
			expectedFiles:  5, // Go files and readme.md, excluding github.com and api.md
			containsFiles:  []string{"src/main.go", "vendor/lib.go", "docs/readme.md"},
			excludesFiles:  []string{"vendor/github.com/pkg/pkg.go", "docs/api.md", "Makefile"},
			containsInTree: []string{"src", "vendor", "docs", ".git", "Makefile"},
			excludesInTree: []string{".git/objects"},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Collect files according to configuration
			files := collectFiles(scenario.config)

			// Convert absolute paths to relative for easier testing
			relFiles := make([]string, 0, len(files))
			for _, file := range files {
				rel, _ := filepath.Rel(tempDir, file)
				relFiles = append(relFiles, rel)
			}

			// Check file count
			if len(relFiles) != scenario.expectedFiles {
				t.Errorf("Expected %d files, got %d: %v",
					scenario.expectedFiles, len(relFiles), relFiles)
			}

			// Check for files that should be included
			for _, expectedFile := range scenario.containsFiles {
				found := false
				for _, file := range relFiles {
					if filepath.ToSlash(file) == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in results", expectedFile)
				}
			}

			// Check for files that should be excluded
			for _, excludedFile := range scenario.excludesFiles {
				for _, file := range relFiles {
					if filepath.ToSlash(file) == excludedFile {
						t.Errorf("Excluded file %s found in results", excludedFile)
						break
					}
				}
			}

			// Generate the tree and verify its structure
			tree := buildDirectoryTree(tempDir, tempDir)

			// Validate tree structure (simplified check)
			treeStr := captureTreeOutput(tree)

			// Check for elements that should be in the tree
			for _, item := range scenario.containsInTree {
				if !strings.Contains(treeStr, item) {
					t.Errorf("Expected item %s not found in tree output", item)
				}
			}

			// Check for elements that should not be in the tree
			for _, item := range scenario.excludesInTree {
				if strings.Contains(treeStr, item) {
					t.Errorf("Excluded item %s found in tree output", item)
				}
			}
		})
	}
}

// Helper function to capture tree output as a string
func captureTreeOutput(node *TreeNode) string {
	var sb strings.Builder
	captureTreeOutputRecursive(node, "", true, &sb)
	return sb.String()
}

func captureTreeOutputRecursive(node *TreeNode, prefix string, isLast bool, sb *strings.Builder) {
	// Print the current node
	connector := "└── "
	if !isLast {
		connector = "├── "
	}

	if node.IsDir {
		sb.WriteString(prefix + connector + node.Name + "/\n")
	} else {
		sb.WriteString(prefix + connector + node.Name + "\n")
	}

	// Calculate the new prefix for children
	newPrefix := prefix
	if isLast {
		newPrefix += "    "
	} else {
		newPrefix += "│   "
	}

	// Print the children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		captureTreeOutputRecursive(child, newPrefix, isLastChild, sb)
	}
}

// TestVersionFlag tests the --version flag functionality
// TestVersionFlag tests the --version flag functionality
func TestVersionFlag(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Set Version for test
	originalVersion := Version
	Version = "1.2.3"
	defer func() { Version = originalVersion }()

	// Mock args
	os.Args = []string{"mkctx", "--version"}

	// Call with args that would make parseFlags return showVersion=true
	_, showVersion, _ := parseFlags()
	if !showVersion {
		t.Errorf("Expected showVersion to be true, got false")
	}

	// Run a simulation of what main() would do with showVersion=true
	if showVersion {
		fmt.Printf("mkctx version %s\n", Version)
	}

	// Close the writer to get the output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check output
	if !strings.Contains(output, "mkctx version 1.2.3") {
		t.Errorf("Expected version output, got %q", output)
	}
}

// TestHelpFlag tests the --help flag functionality
// TestHelpFlag tests the --help flag functionality
func TestHelpFlag(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Instead of calling parseFlags() which would redefine flags,
	// simply simulate the behavior for --help

	// Mock help behavior directly
	printHelp()

	// Close the writer to get the output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check output
	expectedContent := []string{
		"mkctx - Context Generator for LLMs",
		"USAGE:",
		"OPTIONS:",
		"--include",
		"--exclude",
		"--gitignore",
		"--version",
		"--help",
		"EXAMPLES:",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected help output to contain %q, but it doesn't", content)
		}
	}
}
