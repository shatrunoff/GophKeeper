package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Build info (set via ldflags).
var (
	Version   = "dev"
	BuildDate = "unknown"
)

type client struct {
	baseURL string
	token   string
	http    *http.Client
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	c := &client{
		baseURL: getEnv("GOPHERPASS_URL", "http://localhost:8080"),
		token:   os.Getenv("GOPHERPASS_TOKEN"),
		http:    &http.Client{},
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("GopherPass CLI %s (built %s)\n", Version, BuildDate)
	case "register":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gopherpass register <login> <password>")
			return
		}
		c.register(os.Args[2], os.Args[3])
	case "login":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gopherpass login <login> <password>")
			return
		}
		c.login(os.Args[2], os.Args[3])
	case "list":
		c.listSecrets()
	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gopherpass add <type> <payload>")
			return
		}
		c.addSecret(os.Args[2], os.Args[3])
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gopherpass delete <id>")
			return
		}
		c.deleteSecret(os.Args[2])
	case "sync":
		since := ""
		if len(os.Args) > 2 {
			since = os.Args[2]
		}
		c.sync(since)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`GopherPass CLI

Commands:
  version                    Show version info
  register <login> <pass>    Register new user
  login <login> <pass>       Login and get token
  list                       List all secrets
  add <type> <payload>       Add secret (types: credentials, text, binary, card)
  delete <id>                Delete secret
  sync [since]               Sync secrets (since: RFC3339 timestamp)

Environment:
  GOPHERPASS_URL    Server URL (default: http://localhost:8080)
  GOPHERPASS_TOKEN  Auth token (from login command)`)
}

func (c *client) register(login, password string) {
	body, _ := json.Marshal(map[string]string{"login": login, "password": password})
	resp, err := c.http.Post(c.baseURL+"/api/register", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusCreated {
		fmt.Println("Registered successfully")
	} else {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", b)
	}
}

func (c *client) login(login, password string) {
	body, _ := json.Marshal(map[string]string{"login": login, "password": password})
	resp, err := c.http.Post(c.baseURL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		fmt.Printf("Token: %s\n", result["token"])
		fmt.Println("Set: export GOPHERPASS_TOKEN=<token>")
	} else {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", b)
	}
}

func (c *client) listSecrets() {
	req, _ := http.NewRequest("GET", c.baseURL+"/api/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))
}

func (c *client) addSecret(secretType, payload string) {
	body, _ := json.Marshal(map[string]interface{}{
		"type":    secretType,
		"payload": json.RawMessage(payload),
		"version": 0,
	})
	req, _ := http.NewRequest("POST", c.baseURL+"/api/secrets", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))
}

func (c *client) deleteSecret(id string) {
	req, _ := http.NewRequest("DELETE", c.baseURL+"/api/secrets/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("Deleted")
	} else {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", b)
	}
}

func (c *client) sync(since string) {
	url := c.baseURL + "/api/sync"
	if since != "" {
		url += "?since=" + since
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
