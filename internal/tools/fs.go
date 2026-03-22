package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SearchFilesArgs struct {
	RootPath string `json:"root_path"`
	Pattern  string `json:"pattern"`
}

func SearchFiles(args SearchFilesArgs) (string, error) {
	var result strings.Builder
	count := 0
	err := filepath.Walk(args.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() { return nil }
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(args.Pattern)) {
			result.WriteString(fmt.Sprintf("%s\n", path))
			count++
		}
		if count > 100 { return filepath.SkipDir }
		return nil
	})
	if err != nil { return "", err }
	if count == 0 { return "No matches found.", nil }
	return result.String(), nil
}

type RemovePathArgs struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

func RemovePath(args RemovePathArgs) (string, error) {
	if args.Recursive {
		err := os.RemoveAll(args.Path)
		if err != nil { return "", err }
		return fmt.Sprintf("Successfully removed (recursive): %s", args.Path), nil
	}
	err := os.Remove(args.Path)
	if err != nil { return "", err }
	return fmt.Sprintf("Successfully removed: %s", args.Path), nil
}

type ListDirectoryArgs struct {
	Path string `json:"path"`
}

func ListDirectory(args ListDirectoryArgs) (string, error) {
	entries, err := os.ReadDir(args.Path)
	if err != nil { return "", err }
	var result strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			result.WriteString(fmt.Sprintf("[DIR]  %s\n", entry.Name()))
		} else {
			result.WriteString(fmt.Sprintf("[FILE] %s\n", entry.Name()))
		}
	}
	return result.String(), nil
}

type ReadMultipleFilesArgs struct {
	Paths []string `json:"paths"`
}

func ReadMultipleFiles(args ReadMultipleFilesArgs) (string, error) {
	var result strings.Builder
	for _, path := range args.Paths {
		content, err := os.ReadFile(path)
		result.WriteString(fmt.Sprintf("--- FILE: %s ---\n", path))
		if err != nil {
			result.WriteString(fmt.Sprintf("ERROR: %v\n\n", err))
		} else {
			result.WriteString(string(content))
			result.WriteString("\n\n")
		}
	}
	return result.String(), nil
}

type ReadFileArgs struct {
	Path      string `json:"path"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
}

type ReadFileResult struct {
	Content    string `json:"content"`
	TotalLines int    `json:"total_lines"`
}

func ReadFile(args ReadFileArgs) (*ReadFileResult, error) {
	file, err := os.Open(args.Path)
	if err != nil { return nil, err }
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
		if args.LineStart > 0 && count < args.LineStart { continue }
		if args.LineEnd > 0 && count > args.LineEnd { continue }
		lines = append(lines, fmt.Sprintf("%d | %s", count, scanner.Text()))
	}
	return &ReadFileResult{Content: strings.Join(lines, "\n"), TotalLines: count}, scanner.Err()
}

type WriteFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func WriteFile(args WriteFileArgs) (string, error) {
	err := os.WriteFile(args.Path, []byte(args.Content), 0644)
	if err != nil { return "", err }
	diff, _ := GetGitDiff(filepath.Dir(args.Path)) // Skicka mappen till diff
	return diff, nil
}

type EditFileArgs struct {
	Path    string `json:"path"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

func EditFile(args EditFileArgs) (string, error) {
	content, err := os.ReadFile(args.Path)
	if err != nil { return "", err }
	original := string(content)
	newContent := strings.Replace(original, args.OldText, args.NewText, 1)
	err = os.WriteFile(args.Path, []byte(newContent), 0644)
	if err != nil { return "", err }
	diff, _ := GetGitDiff(filepath.Dir(args.Path)) // Skicka mappen till diff
	return diff, nil
}

type GrepSearchArgs struct {
	RootPath string `json:"root_path"`
	Pattern  string `json:"pattern"`
	Suffix   string `json:"suffix"` // Valfritt: t.ex. ".go" eller ".ts"
}

func GrepSearch(args GrepSearchArgs) (string, error) {
	var result strings.Builder
	count := 0
	err := filepath.Walk(args.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() { return nil }
		if args.Suffix != "" && !strings.HasSuffix(path, args.Suffix) { return nil }
		
		content, err := os.ReadFile(path)
		if err != nil { return nil }
		
		if strings.Contains(string(content), args.Pattern) {
			result.WriteString(fmt.Sprintf("MATCH: %s\n", path))
			count++
		}
		if count > 50 { return filepath.SkipDir }
		return nil
	})
	if err != nil { return "", err }
	if count == 0 { return "No matches found.", nil }
	return result.String(), nil
}

