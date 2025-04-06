package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"strconv"

	"github.com/Pratham700/github-search-service/internal/github"
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
	githubParams := make(map[string]string)

	if err := s.processSearchParameters(req, githubParams); err != nil {
		return nil, err
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

// processSearchParameters extracts and validates enum parameters.
func (s *GithubSearchServer) processSearchParameters(req *pb.SearchRequest, githubParams map[string]string) error {
	// Helper function to handle enum to string mapping
	mapEnumToString := func(enumValue int32, mapping map[int32]string, paramName string) error {
		if stringValue, ok := mapping[enumValue]; ok {
			githubParams[paramName] = stringValue
			return nil
		} else if enumValue != 0 {
			log.Printf("Received invalid %s option: %v", paramName, enumValue)
			return status.Errorf(codes.InvalidArgument, "invalid %s option: %v", paramName, enumValue)
		}
		return nil // No error for UNSPECIFIED
	}

	// Define the mappings for SortOption and OrderOption
	sortMapping := map[int32]string{
		int32(pb.SortOption_SORT_INDEXED): "indexed",
	}

	orderMapping := map[int32]string{
		int32(pb.OrderOption_ORDER_ASC):  "asc",
		int32(pb.OrderOption_ORDER_DESC): "desc",
	}

	// Process SortOption
	if err := mapEnumToString(int32(req.GetSort()), sortMapping, "sort"); err != nil {
		return err
	}

	// Process OrderOption
	if err := mapEnumToString(int32(req.GetOrder()), orderMapping, "order"); err != nil {
		return err
	}

	// Handle per_page
	perPage := req.GetPerPage()
	if perPage > 0 {
		if perPage < 1 || perPage > 100 {
			return status.Errorf(codes.InvalidArgument, "invalid value for 'per_page': must be an integer between 1 and 100")
		}
		githubParams["per_page"] = strconv.Itoa(int(perPage))
	}

	// Handle page
	page := req.GetPage()
	if page > 0 {
		if page < 1 {
			return status.Errorf(codes.InvalidArgument, "invalid value for 'page': must be an integer greater than or equal to 1")
		}
		githubParams["page"] = strconv.Itoa(int(page))
	}
	return nil
}
