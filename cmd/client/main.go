package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"gopherpass/internal/helpers"
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
	log     *slog.Logger
}

type ClientOption func(*client)

func WithBaseURL(url string) ClientOption {
	return func(c *client) { c.baseURL = url }
}

func WithToken(token string) ClientOption {
	return func(c *client) { c.token = token }
}

func WithLogger(log *slog.Logger) ClientOption {
	return func(c *client) { c.log = log }
}

func newClient(opts ...ClientOption) *client {
	c := &client{
		baseURL: "http://localhost:8080",
		http:    &http.Client{},
		log:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	c := newClient(
		WithBaseURL(helpers.GetEnv("GOPHERPASS_URL", "http://localhost:8080")),
		WithToken(os.Getenv("GOPHERPASS_TOKEN")),
	)

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
	body, err := json.Marshal(map[string]string{"login": login, "password": password})
	if err != nil {
		c.log.Error("failed to marshal request", "error", err)
		return
	}

	resp, err := c.http.Post(c.baseURL+"/api/register", "application/json", bytes.NewReader(body))
	if err != nil {
		c.log.Error("request failed", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusCreated {
		c.log.Info("registered successfully")
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("failed to read response", "error", err)
		return
	}
	c.log.Error("registration failed", "status", resp.StatusCode, "response", string(b))
}

func (c *client) login(login, password string) {
	body, err := json.Marshal(map[string]string{"login": login, "password": password})
	if err != nil {
		c.log.Error("failed to marshal request", "error", err)
		return
	}

	resp, err := c.http.Post(c.baseURL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		c.log.Error("request failed", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			c.log.Error("failed to read response", "error", err)
			return
		}
		c.log.Error("login failed", "status", resp.StatusCode, "response", string(b))
		return
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.log.Error("failed to decode response", "error", err)
		return
	}

	fmt.Printf("Token: %s\n", result["token"])
	fmt.Println("Set: export GOPHERPASS_TOKEN=<token>")
}

func (c *client) listSecrets() {
	req, err := http.NewRequest("GET", c.baseURL+"/api/secrets", nil)
	if err != nil {
		c.log.Error("failed to create request", "error", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		c.log.Error("request failed", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Warn("failed to close response body", "error", err)
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("failed to read response", "error", err)
		return
	}
	fmt.Println(string(b))
}

func (c *client) addSecret(secretType, payload string) {
	body, err := json.Marshal(map[string]interface{}{
		"type":    secretType,
		"payload": json.RawMessage(payload),
		"version": 0,
	})
	if err != nil {
		c.log.Error("failed to marshal request", "error", err)
		return
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/secrets", bytes.NewReader(body))
	if err != nil {
		c.log.Error("failed to create request", "error", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		c.log.Error("request failed", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Warn("failed to close response body", "error", err)
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("failed to read response", "error", err)
		return
	}
	fmt.Println(string(b))
}

func (c *client) deleteSecret(id string) {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/secrets/"+id, nil)
	if err != nil {
		c.log.Error("failed to create request", "error", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		c.log.Error("request failed", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusNoContent {
		c.log.Info("deleted successfully")
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("failed to read response", "error", err)
		return
	}
	c.log.Error("delete failed", "status", resp.StatusCode, "response", string(b))
}

func (c *client) sync(since string) {
	url := c.baseURL + "/api/sync"
	if since != "" {
		url += "?since=" + since
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.log.Error("failed to create request", "error", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		c.log.Error("request failed", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Warn("failed to close response body", "error", err)
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("failed to read response", "error", err)
		return
	}
	fmt.Println(string(b))
}
