gen:
	@protoc \
	  --go_out=proto --go_opt=paths=source_relative \
	  --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
	  proto/github_search_service.proto

run-server:
	@go run cmd/server/main.go