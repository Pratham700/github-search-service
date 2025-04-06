package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"strconv"

	"github.com/Pratham700/github-search-service/internal/github"
	"github.com/Pratham700/github-search-service/internal/util"
	pb "github.com/Pratham700/github-search-service/proto/proto"
)

type GithubSearchServer struct {
	pb.UnimplementedGithubSearchServiceServer
	gitHubClient *github.GitHubClient
}

// NewGithubSearchServer creates a new GithubSearchServer.
func NewGithubSearchServer() (*GithubSearchServer, error) {
	// Read the GitHub API base URL from an environment variable (optional)
	baseURL := os.Getenv("GITHUB_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.github.com" // Default base URL
	}

	return &GithubSearchServer{
		gitHubClient: github.NewGitHubClient(baseURL),
	}, nil
}

// Search implements the Search gRPC method.
func (s *GithubSearchServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	log.Printf("Received Search request: SearchTerm=%s, User=%s", req.SearchTerm, req.User)

	// Retrieve the GitHub token from the context using the helper function
	authToken, err := GetAuthTokenFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get github token from context: %w", err)
	}

	// Extract GitHub API parameters from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	githubParams := make(map[string]string)
	if ok {
		var err error
		githubParams, err = extractGithubSearchParams(md)
		if err != nil {
			return nil, err // Return the gRPC error
		}
	}

	// Call the GitHub API to search for files
	files, err := s.gitHubClient.SearchFiles(ctx, req.SearchTerm, req.User, authToken, githubParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search files on GitHub: %w", err)
	}

	// Prepare the response
	var results []*pb.Result
	for _, file := range files {
		fileURL := github.ExtractFileURL(file)
		repoName := github.ExtractRepoUrl(file)
		if fileURL != "" && repoName != "" {
			results = append(results, &pb.Result{
				FileUrl: fileURL,
				Repo:    repoName,
			})
		}
	}

	log.Printf("Found %d results", len(results))

	return &pb.SearchResponse{
		Results: results,
	}, nil
}

// extractGithubSearchParams extracts optional GitHub Search API parameters from gRPC metadata.
func extractGithubSearchParams(md metadata.MD) (map[string]string, error) {
	params := make(map[string]string)

	// Define validation functions
	validators := map[string]func(string) error{
		"sort": func(value string) error {
			if value != "indexed" {
				return fmt.Errorf("allowed value: 'indexed'")
			}
			return nil
		},
		"order": func(value string) error {
			if value != "asc" && value != "desc" {
				return fmt.Errorf("allowed values: 'asc', 'desc'")
			}
			return nil
		},
		"per_page": func(value string) error {
			perPage, err := strconv.Atoi(value)
			if err != nil || perPage < 1 || perPage > 100 {
				return fmt.Errorf("must be an integer between 1 and 100")
			}
			return nil
		},
		"page": func(value string) error {
			page, err := strconv.Atoi(value)
			if err != nil || page < 1 {
				return fmt.Errorf("must be an integer greater than or equal to 1")
			}
			return nil
		},
	}

	// Helper function to validate and extract metadata values
	extractAndValidate := func(key string, validate func(string) error) error {
		if value, ok := util.ExtractMetadataValue(md, key); ok {
			if err := validate(value); err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid value for '%s': %s", key, err)
			}
			params[key] = value
		}
		return nil
	}

	// Iterate over the validators map
	for key, validate := range validators {
		if err := extractAndValidate(key, validate); err != nil {
			return nil, err
		}
	}

	return params, nil
}
