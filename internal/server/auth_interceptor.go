package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authTokenKey struct{} // Define a private custom key type

// setAuthTokenInContext stores the authentication token in the context.
func setAuthTokenInContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authTokenKey{}, token)
}

// GetAuthTokenFromContext retrieves the authentication token from the context.
func GetAuthTokenFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value(authTokenKey{}).(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "github-token is required in metadata but not found")
	}
	return token, nil
}

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
		return nil, status.Error(codes.Unauthenticated, "github-token is required in metadata")
	}

	newCtx := setAuthTokenInContext(ctx, tokenValues[0])

	// Call the handler with the modified context
	return handler(newCtx, req)
}
