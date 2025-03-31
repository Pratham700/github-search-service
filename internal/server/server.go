package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"os"

	"github.com/Pratham700/github-search-service/internal/github"
	"github.com/Pratham700/github-search-service/internal/util"
	pb "github.com/Pratham700/github-search-service/proto/proto"
)

type GithubSearchServer struct {
	pb.UnimplementedGithubSearchServiceServer
	GitHubClient *github.GitHubClient
}

// NewGithubSearchServer creates a new GithubSearchServer.
func NewGithubSearchServer() *GithubSearchServer {
	// Read the GitHub API base URL from an environment variable (optional)
	baseURL := os.Getenv("GITHUB_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.github.com" // Default base URL
	}

	return &GithubSearchServer{
		GitHubClient: github.NewGitHubClient(baseURL),
	}
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
	files, err := s.GitHubClient.SearchFiles(ctx, req.SearchTerm, req.User, authToken, githubParams)
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
	// Extract optional GitHub Search API parameters safely
	var ok bool
	var value string

	// Validate 'sort'
	if value, ok = util.ExtractMetadataValue(md, "sort"); ok {
		if value != "indexed" {
			return nil, status.Errorf(codes.InvalidArgument, "invalid value for 'sort': %s. Allowed value: 'indexed'", value)
		}
		params["sort"] = value
	}
	// Validate 'order'
	if value, ok = util.ExtractMetadataValue(md, "order"); ok {
		if value != "asc" && value != "desc" {
			return nil, status.Errorf(codes.InvalidArgument, "invalid value for 'order': %s. Allowed values: 'asc', 'desc'", value)
		}
		params["order"] = value
	}
	// Validate 'per_page'
	if value, ok = util.ExtractMetadataValue(md, "per_page"); ok {
		var perPage int
		if _, err := fmt.Sscan(value, &perPage); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid value for 'per_page': %s. Must be an integer", value)
		}
		if perPage < 1 || perPage > 100 {
			return nil, status.Errorf(codes.InvalidArgument, "invalid value for 'per_page': %s. Must be between 1 and 100", value)
		}
		params["per_page"] = value
	}
	// Validate 'page'
	if value, ok = util.ExtractMetadataValue(md, "page"); ok {
		var page int
		if _, err := fmt.Sscan(value, &page); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid value for 'page': %s. Must be an integer", value)
		}
		if page < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "invalid value for 'page': %s. Must be greater than or equal to 1", value)
		}
		params["page"] = value
	}
	return params, nil
}
