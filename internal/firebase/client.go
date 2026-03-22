package firebase

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"cyborg-node/internal/config"
	"cyborg-node/internal/tools"
)

type Client struct {
	BaseURL string
	Token   string
	nodeId  string
}

// GitIdentity is a safe representation of a GitToken without the secret string
type GitIdentity struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	URLPattern string `json:"url_pattern"`
}

type Command struct {
	ID        string                 `json:"id"`
	NodeID    string                 `json:"node"`
	SessionID string                 `json:"session"`
	Branch    string                 `json:"branch"`
	Name      string                 `json:"name"`
	Args      map[string]interface{} `json:"args"`
	Status    string                 `json:"status"`
}

func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{
		BaseURL: baseURL,
		Token:   token,
	}
}

func (c *Client) RegisterNode(cfg *config.Config) (string, error) {
	return c.UpdateStatus(cfg, "online")
}

func (c *Client) UpdateStatus(cfg *config.Config, status string) (string, error) {
	url := fmt.Sprintf("%s/node/pulse", c.BaseURL)
	
	// 1. Create a safe list of identities without secret tokens
	identities := make([]GitIdentity, len(cfg.GitTokens))
	for i, t := range cfg.GitTokens {
		identities[i] = GitIdentity{
			Name:       t.Name,
			Type:       t.Type,
			URLPattern: t.URLPattern,
		}
	}

	// 2. Scan for local repositories if search paths are defined
	var localRepos []tools.LocalRepo
	if len(cfg.SearchPaths) > 0 {
		var err error
		localRepos, err = tools.ScanForRepositories(cfg.SearchPaths)
		if err != nil {
			log.Printf("⚠️ Repository scan failed: %v", err)
		}
	}

	data := map[string]interface{}{
		"name":           cfg.NodeName,
		"status":         status,
		"git_identities": identities,
		"local_repos":    localRepos,
	}
	
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cyborg-Node-Token", c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("status update failed with code %d", resp.StatusCode)
	}

	var res struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	c.nodeId = res.ID
	return c.nodeId, nil
}

func (c *Client) StreamCommands(handler func(cmd Command)) error {
	if c.nodeId == "" {
		return fmt.Errorf("node not registered")
	}

	url := fmt.Sprintf("%s/node/stream", c.BaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Cyborg-Node-Token", c.Token)
	req.Header.Set("Accept", "text/event-stream")

	log.Printf("📡 Opening real-time stream: %s", url)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("stream failed with status %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("stream connection lost: %v", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data:")
		if strings.TrimSpace(data) == ":heartbeat" {
			continue
		}

		var cmd Command
		if err := json.Unmarshal([]byte(data), &cmd); err != nil {
			log.Printf("⚠️ Error unmarshaling streamed command: %v", err)
			continue
		}

		go handler(cmd)
	}
}

func (c *Client) UpdateCommand(cmdId, status, output, errMsg string) error {
	url := fmt.Sprintf("%s/node/action/update", c.BaseURL)
	data := map[string]interface{}{
		"id":     cmdId,
		"status": status,
		"output": output,
		"error":  errMsg,
	}
	
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cyborg-Node-Token", c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("update failed with status %d", resp.StatusCode)
	}
	return nil
}
