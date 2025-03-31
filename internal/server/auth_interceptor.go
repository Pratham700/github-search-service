package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor is a gRPC interceptor that extracts the GitHub token from the metadata.
func AuthInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Retrieve the GitHub token from the incoming metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	tokenValues := md.Get("github-token")
	if len(tokenValues) == 0 {
		return nil, fmt.Errorf("github-token is required in metadata")
	}
	authToken := tokenValues[0]

	// Add the token to the context so that the service handler can access it.
	newCtx := context.WithValue(ctx, "github-token", authToken)

	// Call the handler with the modified context
	return handler(newCtx, req)
}

// GetAuthTokenFromContext retrieves the GitHub token from the context.
func GetAuthTokenFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value("github-token").(string)
	if !ok {
		return "", fmt.Errorf("github-token not found in context")
	}
	return token, nil
}
