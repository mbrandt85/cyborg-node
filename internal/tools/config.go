package tools

import "cyborg-node/internal/config"

var globalGitTokens []config.GitToken

func SetGitTokens(tokens []config.GitToken) {
	globalGitTokens = tokens
}

// GetTokenByType returns the token for a specific provider (GITHUB/GITLAB)
func GetTokenByType(provider string) string {
	for _, t := range globalGitTokens {
		if t.Type == provider {
			return t.Token
		}
	}
	return ""
}
