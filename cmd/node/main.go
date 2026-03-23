package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"cyborg-node/internal/config"
	"cyborg-node/internal/firebase"
	"cyborg-node/internal/tools"
)

var (
	mu     sync.Mutex
	nodeID string
)

func main() {
	log.Println("🦾 CYBORG NODE (Real-time Edition) initializing...")

	cfg, err := config.LoadConfig("cyborg-node-config.json")
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	if cfg.GatewayURL == "" || cfg.NodeKey == "" {
		log.Fatal("❌ Missing GatewayURL or NodeKey in config.")
	}

	tools.SetGitTokens(cfg.GitTokens)

	client := firebase.NewClient(cfg.GatewayURL, cfg.NodeKey)
	log.Printf("📡 Connected to Gateway: %s\n", cfg.GatewayURL)

	// --- 1. Graceful Shutdown Setup ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\n🛑 Received signal: %v. Cleaning up...", sig)
		
		if nodeID != "" {
			log.Println("👋 Sending offline status to hub...")
			client.UpdateStatus(cfg, "offline")
		}
		
		log.Println("🔌 Node shutdown complete. Stay safe, organic.")
		os.Exit(0)
	}()

	// --- 2. Registration & Real-time loop ---
	for {
		log.Println("🛰️ Registering node and sending pulse...")
		id, err := client.RegisterNode(cfg)
		if err != nil {
			log.Printf("⚠️ Handshake failed, retrying in 10s: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if id != "" {
			mu.Lock()
			nodeID = id
			mu.Unlock()
			log.Printf("✅ Node uplink active (ID: %s)", id)
		}

		// SSE stream will block here until disconnected
		err = client.StreamCommands(func(cmd firebase.Command) {
			executeCommand(client, cmd)
		})
		
		if err != nil {
			log.Printf("⏳ Connection lost: %v. Attempting to reconnect...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func executeCommand(client *firebase.Client, cmd firebase.Command) {
	log.Printf("\n[>] Executing Command: %s (ID: %s)", cmd.Name, cmd.ID)
	
	getPath := func(args map[string]interface{}) string {
		if path, ok := args["path"].(string); ok {
			return path
		}
		if rootPath, ok := args["root_path"].(string); ok {
			return rootPath
		}
		return ""
	}

	repoPath := getPath(cmd.Args)

	if cmd.Branch != "" && repoPath != "" {
		log.Printf("Switching to branch: %s", cmd.Branch)
		_, err := tools.GitCheckout(repoPath, cmd.Branch)
		if err != nil {
			log.Printf("⚠️ Branch switch failed: %v", err)
		}
	}

	var output string
	var err error

	mustMarshal := func(v any) []byte {
		b, _ := json.Marshal(v)
		return b
	}

	switch cmd.Name {
	case "fs.read":
		var args tools.ReadFileArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		res, e := tools.ReadFile(args)
		err = e
		if res != nil { output = res.Content }
	case "fs.read_multiple":
		var args tools.ReadMultipleFilesArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.ReadMultipleFiles(args)
	case "fs.write":
		var args tools.WriteFileArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.WriteFile(args)
	case "fs.edit":
		var args tools.EditFileArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.EditFile(args)
	case "fs.remove":
		var args tools.RemovePathArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.RemovePath(args)
	case "fs.list_directory":
		var args tools.ListDirectoryArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.ListDirectory(args)
	case "fs.search":
		var args tools.SearchFilesArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.SearchFiles(args)
	case "fs.grep_search":
		var args tools.GrepSearchArgs
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.GrepSearch(args)
	case "git.diff":
		var args map[string]string
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.GetGitDiff(args["path"])
	case "git.status":
		var args map[string]string
		json.Unmarshal(mustMarshal(cmd.Args), &args)
		output, err = tools.GitStatus(args["path"])
	default:
		err = fmt.Errorf("Unknown command: %s", cmd.Name)
	}

	errMsg := ""
	status := "completed"
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}

	if updateErr := client.UpdateCommand(cmd.ID, status, output, errMsg); updateErr != nil {
		log.Printf("❌ Failed to report status: %v", updateErr)
	}
}
