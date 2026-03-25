package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/somaz94/go-git-commit-action/internal/errors"
)

const (
	apiBaseURL     = "https://api.github.com"
	apiVersion     = "2022-11-28"
	acceptHeader   = "application/vnd.github+json"
	curlMaxTimeSec = "30"
)

// Client handles GitHub API interactions.
type Client struct {
	token string
	repo  string
}

// NewClient creates a new GitHub API client.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		repo:  os.Getenv("GITHUB_REPOSITORY"),
	}
}

// Post sends a POST request to the GitHub API.
func (c *Client) Post(endpoint string, data interface{}) (map[string]interface{}, error) {
	return c.request("POST", endpoint, data)
}

// Patch sends a PATCH request to the GitHub API.
func (c *Client) Patch(endpoint string, data interface{}) (map[string]interface{}, error) {
	return c.request("PATCH", endpoint, data)
}

// GetArray sends a GET request to the GitHub API and returns an array response.
func (c *Client) GetArray(endpoint string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s%s", apiBaseURL, c.repo, endpoint)

	cmd := exec.Command("curl", "-s", "--max-time", curlMaxTimeSec,
		"-w", "\n%{http_code}",
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.token),
		"-H", fmt.Sprintf("Accept: %s", acceptHeader),
		"-H", fmt.Sprintf("X-GitHub-Api-Version: %s", apiVersion),
		url)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New("GitHub API GET", err)
	}

	body, statusCode, err := parseHTTPResponse(output)
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, errors.NewAPIError("GitHub API GET", fmt.Sprintf("HTTP %d", statusCode))
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.New("parse GitHub API response", err)
	}

	return result, nil
}

// Repo returns the GitHub repository name.
func (c *Client) Repo() string {
	return c.repo
}

// parseHTTPResponse splits curl output (with -w "\n%{http_code}") into body and status code.
func parseHTTPResponse(output []byte) ([]byte, int, error) {
	raw := strings.TrimSpace(string(output))
	lastNewline := strings.LastIndex(raw, "\n")
	if lastNewline == -1 {
		// Only status code, no body
		code, err := strconv.Atoi(raw)
		if err != nil {
			return output, 0, errors.New("parse HTTP status code", err)
		}
		return nil, code, nil
	}

	body := raw[:lastNewline]
	codeStr := strings.TrimSpace(raw[lastNewline+1:])

	// Validate status code is numeric (3 digits)
	if !regexp.MustCompile(`^\d{3}$`).MatchString(codeStr) {
		return []byte(body), 0, errors.New("parse HTTP status code", fmt.Errorf("invalid status code: %q", codeStr))
	}

	code, _ := strconv.Atoi(codeStr)
	return []byte(body), code, nil
}

// request sends an HTTP request to the GitHub API.
func (c *Client) request(method, endpoint string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("marshal request data", err)
	}

	url := fmt.Sprintf("%s/repos/%s%s", apiBaseURL, c.repo, endpoint)

	cmd := exec.Command("curl", "-s", "--max-time", curlMaxTimeSec,
		"-w", "\n%{http_code}",
		"-X", method,
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.token),
		"-H", fmt.Sprintf("Accept: %s", acceptHeader),
		"-H", "Content-Type: application/json",
		url,
		"-d", string(jsonData))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New("GitHub API "+method, err)
	}

	body, statusCode, err := parseHTTPResponse(output)
	if err != nil {
		return nil, err
	}

	// For client/server errors, try to parse JSON body so the caller
	// can inspect API error details (e.g., "A pull request already exists").
	if statusCode < 200 || statusCode >= 300 {
		var errResult map[string]interface{}
		if json.Unmarshal(body, &errResult) == nil {
			// Return the parsed error response — caller checks for "message" key
			return errResult, nil
		}
		return nil, errors.NewAPIError("GitHub API "+method, fmt.Sprintf("HTTP %d", statusCode))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// Some endpoints (e.g., labels API) return an array on success.
		// If array parse succeeds, treat as successful with no map result.
		var arrayResult []interface{}
		if json.Unmarshal(body, &arrayResult) == nil {
			return nil, nil
		}
		return nil, errors.New("parse GitHub API response", err)
	}

	return result, nil
}
