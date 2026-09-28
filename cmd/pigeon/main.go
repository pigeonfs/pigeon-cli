package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const version = "0.1.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 1
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Println("pigeon", version)
		return 0
	case "help", "--help", "-h":
		usage()
		return 0
	case "doctor":
		return doctor()
	case "emails":
		if len(args) < 2 || args[1] != "send" {
			fmt.Fprintln(os.Stderr, "usage: pigeon emails send --from ADDR --to ADDR --subject TEXT [--html HTML] [--text TEXT]")
			return 1
		}
		return sendEmail(args[2:])
	case "domains":
		if len(args) < 2 || args[1] != "list" {
			fmt.Fprintln(os.Stderr, "usage: pigeon domains list")
			return 1
		}
		return listDomains()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		usage()
		return 1
	}
}

func usage() {
	fmt.Print(`Pigeon CLI — send email from a terminal, agent, or CI job.

  pigeon emails send --from ADDR --to ADDR --subject TEXT [--html HTML] [--text TEXT]
  pigeon domains list
  pigeon doctor
  pigeon version

Auth (first match wins):
  1. --api-key
  2. PIGEON_API_KEY
  3. ~/.config/pigeon/credentials.json

Base URL: PIGEON_BASE_URL (default http://localhost:4005)
`)
}

func sendEmail(args []string) int {
	fs := flag.NewFlagSet("emails send", flag.ContinueOnError)
	from := fs.String("from", "", "from address")
	to := fs.String("to", "", "to address")
	subject := fs.String("subject", "", "subject")
	html := fs.String("html", "", "html body")
	text := fs.String("text", "", "text body")
	apiKey := fs.String("api-key", "", "API key")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if *from == "" || *to == "" || *subject == "" {
		fmt.Fprintln(os.Stderr, "--from, --to, and --subject are required")
		return 1
	}
	if *html == "" && *text == "" {
		*text = *subject
	}
	body := map[string]any{
		"from":    *from,
		"to":      []string{*to},
		"subject": *subject,
	}
	if *html != "" {
		body["html"] = *html
	}
	if *text != "" {
		body["text"] = *text
	}
	out, err := request("POST", "/api/emails", body, *apiKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}

func listDomains() int {
	out, err := request("GET", "/api/domains", nil, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}

func doctor() int {
	key, source := resolveKey("")
	base := strings.TrimRight(envOr("PIGEON_BASE_URL", "http://localhost:4005"), "/")
	fmt.Println("pigeon", version)
	fmt.Println("base_url", base)
	if key == "" {
		fmt.Println("api_key missing")
		return 1
	}
	fmt.Println("api_key", source, mask(key))
	return 0
}

func request(method, path string, body any, flagKey string) ([]byte, error) {
	key, _ := resolveKey(flagKey)
	if key == "" {
		return nil, fmt.Errorf("auth_error: set --api-key, PIGEON_API_KEY, or run with a key in ~/.config/pigeon/credentials.json")
	}
	base := strings.TrimRight(envOr("PIGEON_BASE_URL", "http://localhost:4005"), "/")
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, base+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "pigeon-cli/"+version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

func resolveKey(flagKey string) (string, string) {
	if strings.TrimSpace(flagKey) != "" {
		return strings.TrimSpace(flagKey), "flag"
	}
	if v := os.Getenv("PIGEON_API_KEY"); v != "" {
		return v, "env"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ""
	}
	raw, err := os.ReadFile(filepath.Join(home, ".config", "pigeon", "credentials.json"))
	if err != nil {
		return "", ""
	}
	var file struct {
		APIKey string `json:"api_key"`
	}
	if json.Unmarshal(raw, &file) != nil {
		return "", ""
	}
	return file.APIKey, "config"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mask(key string) string {
	if len(key) < 8 {
		return "****"
	}
	return key[:4] + "…" + key[len(key)-3:]
}
