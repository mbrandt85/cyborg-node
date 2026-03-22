package tools

import (
	"os/exec"
	"strings"
)

func GetGitDiff(repoPath string) (string, error) {
	// First, mark untracked files with intent-to-add so they show up in git diff
	exec.Command("git", "-C", repoPath, "add", "-N", ".").Run()

	cmd := exec.Command("git", "diff")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil { return "", err }
	return string(output), nil
}

func GitCheckout(repoPath, branch string) (string, error) {
	if branch == "" { return "No branch specified, staying on current.", nil }
	
	// Check if already on branch
	current := GetCurrentBranch(repoPath)
	if current == branch {
		return "Already on branch " + branch, nil
	}

	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try creating the branch if it doesn't exist
		cmd = exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = repoPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			return string(output), err
		}
	}
	return string(output), nil
}

func GetCurrentBranch(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil { return "unknown" }
	return strings.TrimSpace(string(output))
}

func GitStatus(repoPath string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil { return "", err }
	return string(output), nil
}

func GitCommit(repoPath, message string) (string, error) {
	cmd := exec.Command("git", "commit", "-am", message)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil { return string(output), err }
	return string(output), nil
}

func GitPush(repoPath, branch string) (string, error) {
	if branch == "" { branch = "HEAD" }
	cmd := exec.Command("git", "push", "origin", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil { return string(output), err }
	return string(output), nil
}
