package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GitHubClient is a client for interacting with the GitHub API.
type GitHubClient struct {
	client *http.Client

	// Base URL for the GitHub API
	baseURL string
}

type GithubSearchCodeResponse struct {
	TotalCount        int                `json:"total_count"`
	IncompleteResults bool               `json:"incomplete_results"`
	Items             []GitHubSearchItem `json:"items"`
}

type GitHubSearchItem struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	HTMLURL    string `json:"html_url"`
	Repository GitHubRepository
}

type GitHubRepository struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
}

// NewGitHubClient creates a new GitHubClient.
// It now takes the base URL as argument.
func NewGitHubClient(baseURL string) *GitHubClient {
	return &GitHubClient{
		client:  &http.Client{Timeout: time.Second * 5},
		baseURL: baseURL,
	}
}

// SearchFiles searches for files on GitHub based on the provided search term and user.
func (c *GitHubClient) SearchFiles(ctx context.Context, searchTerm string, user string, authToken string, githubParams map[string]string) ([]GitHubSearchItem, error) {
	// Construct the API URL
	const relativeURL = "/search/code" // Relative URL for the search endpoint
	apiURL := c.baseURL + relativeURL

	queryParams := url.Values{}
	queryParams.Set("q", searchTerm)
	if user != "" {
		queryParams.Set("q", queryParams.Get("q")+" user:"+user)
	}

	for key, value := range githubParams {
		queryParams.Set(key, value)
	}

	apiURL = apiURL + "?" + queryParams.Encode()

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the Accept header to specify the desired API version
	req.Header.Set("Accept", "application/vnd.github+json")

	// Add the Authorization header with the Personal Access Token
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28") // Add the API version header

	// Make the API request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		// Extract the response body for logging and debugging
		var responseBody string
		if bodyBytes, err := io.ReadAll(resp.Body); err == nil {
			responseBody = string(bodyBytes)
		} else {
			responseBody = "failed to read response body"
		}
		return nil, fmt.Errorf("GitHub API returned an error: %s (status code: %d, response: %s)", resp.Status, resp.StatusCode, responseBody)
	}

	bodyBytes, err := io.ReadAll(resp.Body) // Read the entire response body again
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for decoding: %w", err)
	}

	// Parse the response
	var result GithubSearchCodeResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Items, nil
}

// ExtractFileURL extracts the file URL from the search result.
func ExtractFileURL(item GitHubSearchItem) string {
	return item.HTMLURL
}

// ExtractRepoUrl extracts the repository name from the search result.
func ExtractRepoUrl(item GitHubSearchItem) string {
	return "https://github.com/" + item.Repository.FullName // Access using dot notation
}
