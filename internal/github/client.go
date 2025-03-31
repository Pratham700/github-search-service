package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GitHubClient is a client for interacting with the GitHub API.
type GitHubClient struct {
	client    *http.Client
	authToken string // GitHub Personal Access Token
	baseURL   string // Base URL for the GitHub API
}

// NewGitHubClient creates a new GitHubClient.
// It now takes the GitHub Personal Access Token and the base URL as arguments.
func NewGitHubClient(baseURL string) *GitHubClient {
	if baseURL == "" {
		baseURL = "https://api.github.com" // Default base URL
	}
	return &GitHubClient{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

// SearchFiles searches for files on GitHub based on the provided search term and user.
func (c *GitHubClient) SearchFiles(ctx context.Context, searchTerm string, user string, authToken string, githubParams map[string]string) ([]map[string]interface{}, error) {
	// Construct the API URL
	relativeURL := "/search/code" // Relative URL for the search endpoint
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
		return nil, fmt.Errorf("GitHub API returned an error: %s", resp.Status)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the items from the response
	items, ok := result["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse items from response")
	}

	var files []map[string]interface{}
	for _, item := range items {
		file, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

// ExtractFileURL extracts the file URL from the search result.
func ExtractFileURL(file map[string]interface{}) string {
	htmlURL, ok := file["html_url"].(string)
	if !ok {
		return ""
	}
	return htmlURL
}

// ExtractRepoName extracts the repository name from the search result.
func ExtractRepoName(file map[string]interface{}) string {
	repo, ok := file["repository"].(map[string]interface{})
	if !ok {
		return ""
	}
	fullName, ok := repo["full_name"].(string)
	if !ok {
		return ""
	}
	return fullName
}
