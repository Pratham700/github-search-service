package util

import "google.golang.org/grpc/metadata"

// ExtractMetadataValue extracts a single value from gRPC metadata for a given key.
// It returns the extracted value and a boolean indicating whether the key was found.
func ExtractMetadataValue(md metadata.MD, key string) (string, bool) {
	if values := md.Get(key); len(values) > 0 {
		return values[0], true
	}
	return "", false
}
