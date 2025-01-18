package github

type GitHubFile struct {
	Path    string
	Content string
}

// LabelAnalysis represents the result of analyzing an issue for label suggestions
type LabelAnalysis struct {
	// SuggestedLabels is a map of label names to confidence scores (0.0-1.0)
	SuggestedLabels map[string]float64
	// Explanation provides reasoning for each suggested label
	Explanation string
}

// FileFilter represents the configuration for file filtering
type FileFilter struct {
	// AllowedExtensions is a list of file extensions to include (e.g. ".go", ".md")
	AllowedExtensions []string
	// AllowedFiles is a list of specific files to include without extensions (e.g. "Dockerfile", "Makefile")
	AllowedFiles []string
	// ExcludedPaths contains path patterns to exclude (e.g. "vendor/", "test/")
	ExcludedPaths []string
	// ExcludedFiles contains specific file patterns to exclude (e.g. "*_test.go")
	ExcludedFiles []string
}

// DefaultFileFilter returns the default file filter configuration
func DefaultFileFilter() FileFilter {
	return FileFilter{
		AllowedExtensions: []string{
			".go", ".js", ".ts", ".py", ".java", ".rb", ".php",
			".md", ".txt", ".yaml", ".yml", ".json", ".rs", ".sh",
		},
		AllowedFiles: []string{
			"Dockerfile",
			"Makefile",
			"README",
		},
		ExcludedPaths: []string{
			"vendor/",
			"node_modules/",
			"dist/",
			"build/",
		},
		ExcludedFiles: []string{
			"*_test.go",
			"*.test.js",
			"*.spec.ts",
		},
	}
}
