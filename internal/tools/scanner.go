package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LocalRepo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	RemoteURL string `json:"remote_url"`
}

func ScanForRepositories(paths []string) ([]LocalRepo, error) {
	var repos []LocalRepo

	for _, root := range paths {
		// Clean the path (handle ~ if needed, though we expect absolute paths from config)
		root = filepath.Clean(root)

		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue walking
			}

			// If we find a .git directory, we've found a repo
			if info.IsDir() && info.Name() == ".git" {
				repoPath := filepath.Dir(path)
				
				// Get remote URL
				remoteURL := getRemoteURL(repoPath)
				
				repos = append(repos, LocalRepo{
					Name:      filepath.Base(repoPath),
					Path:      repoPath,
					RemoteURL: remoteURL,
				})
				
				return filepath.SkipDir // Don't look inside .git
			}
			return nil
		})
	}

	return repos, nil
}

func getRemoteURL(path string) string {
	cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
