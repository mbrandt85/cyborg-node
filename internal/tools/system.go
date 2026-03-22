package tools

import (
	"errors"
	"os/exec"
	"strings"
)

type RunShellArgs struct {
	Command string `json:"command"`
}

func RunShell(args RunShellArgs) (string, error) {
	if !isCommandAllowed(args.Command) {
		return "", errors.New("security violation: this command is not allowed for autonomous execution")
	}

	parts := strings.Fields(args.Command)
	cmd := exec.Command(parts[0], parts[1:]...)

	// In a real secure node, we would set the CWD to the project root
	// and potentially other sandboxing mechanisms.
	// cmd.Dir = "/path/to/project" 

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

func isCommandAllowed(command string) bool {
	// Trim and normalize
	cmd := strings.TrimSpace(command)
	
	// Block dangerous git write operations
	if strings.HasPrefix(cmd, "git") {
		if strings.Contains(cmd, "push") || strings.Contains(cmd, "commit") || strings.Contains(cmd, "merge") {
			return false
		}
	}
	
	// Block dangerous file system operations that bypass tools
	if strings.HasPrefix(cmd, "rm ") || strings.HasPrefix(cmd, "mv ") || strings.HasPrefix(cmd, "cp ") {
		// Allow specific safe uses if needed, otherwise block all.
		// For now, we block. Agent should use fs.delete, etc.
		if !strings.Contains(cmd, "--force") && !strings.Contains(cmd, " -rf") {
			// Maybe allow simple cases, but for now, no.
		}
		// return false; // Let's be cautious for now and discuss what to allow
	}

	// Default to allowed for now, but a whitelist is safer.
	return true
}
