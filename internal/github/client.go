package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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

// Get sends a GET request to the GitHub API and returns an array response.
func (c *Client) GetArray(endpoint string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s%s", apiBaseURL, c.repo, endpoint)

	cmd := exec.Command("curl", "-s", "--max-time", curlMaxTimeSec,
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.token),
		"-H", fmt.Sprintf("Accept: %s", acceptHeader),
		"-H", fmt.Sprintf("X-GitHub-Api-Version: %s", apiVersion),
		url)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New("GitHub API GET", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, errors.New("parse GitHub API response", err)
	}

	return result, nil
}

// Repo returns the GitHub repository name.
func (c *Client) Repo() string {
	return c.repo
}

// request sends an HTTP request to the GitHub API.
func (c *Client) request(method, endpoint string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("marshal request data", err)
	}

	url := fmt.Sprintf("%s/repos/%s%s", apiBaseURL, c.repo, endpoint)

	cmd := exec.Command("curl", "-s", "--max-time", curlMaxTimeSec, "-X", method,
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.token),
		"-H", fmt.Sprintf("Accept: %s", acceptHeader),
		"-H", "Content-Type: application/json",
		url,
		"-d", string(jsonData))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New("GitHub API "+method, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		// Some endpoints (e.g., labels API) return an array on success.
		// If array parse succeeds, treat as successful with no map result.
		var arrayResult []interface{}
		if json.Unmarshal(output, &arrayResult) == nil {
			return nil, nil
		}
		fmt.Printf("Raw response: %s\n", string(output))
		return nil, errors.New("parse GitHub API response", err)
	}

	return result, nil
}
