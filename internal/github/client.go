package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/somaz94/go-git-commit-action/internal/errors"
)

const (
	apiBaseURL     = "https://api.github.com"
	apiVersion     = "2022-11-28"
	acceptHeader   = "application/vnd.github+json"
	requestTimeout = 30 * time.Second
)

// Client handles GitHub API interactions.
type Client struct {
	token      string
	repo       string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		repo:       os.Getenv("GITHUB_REPOSITORY"),
		baseURL:    apiBaseURL,
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

// Post sends a POST request to the GitHub API.
func (c *Client) Post(ctx context.Context, endpoint string, data interface{}) (map[string]interface{}, error) {
	return c.request(ctx, http.MethodPost, endpoint, data)
}

// Patch sends a PATCH request to the GitHub API.
func (c *Client) Patch(ctx context.Context, endpoint string, data interface{}) (map[string]interface{}, error) {
	return c.request(ctx, http.MethodPatch, endpoint, data)
}

// GetArray sends a GET request to the GitHub API and returns an array response.
func (c *Client) GetArray(ctx context.Context, endpoint string) ([]map[string]interface{}, error) {
	body, statusCode, err := c.do(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.New("GitHub API GET", err)
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

// do performs an HTTP request against the GitHub API and returns the response
// body and status code. payload is nil for requests without a body.
func (c *Client) do(ctx context.Context, method, endpoint string, payload []byte) ([]byte, int, error) {
	url := fmt.Sprintf("%s/repos/%s%s", c.baseURL, c.repo, endpoint)

	var reqBody io.Reader
	if payload != nil {
		reqBody = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", acceptHeader)
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

// request sends a POST/PATCH request with a JSON body to the GitHub API.
func (c *Client) request(ctx context.Context, method, endpoint string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("marshal request data", err)
	}

	body, statusCode, err := c.do(ctx, method, endpoint, jsonData)
	if err != nil {
		return nil, errors.New("GitHub API "+method, err)
	}

	// For client/server errors, try to parse the JSON body so the caller
	// can inspect API error details (e.g., "A pull request already exists").
	if statusCode < 200 || statusCode >= 300 {
		var errResult map[string]interface{}
		if json.Unmarshal(body, &errResult) == nil {
			// Return the parsed error response — caller checks for "message" key.
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
