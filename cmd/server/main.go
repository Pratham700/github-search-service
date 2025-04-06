package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/Pratham700/github-search-service/internal/server"
	pb "github.com/Pratham700/github-search-service/proto/proto"
)

func main() {
	// Set the port for the gRPC server
	var port int
	flag.IntVar(&port, "port", 50051, "The port number to listen on")
	flag.Parse() // Parse the command-line flags
	log.Printf("Using gRPC server port: %d", port)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	// Create a new gRPC server
	s := grpc.NewServer(grpc.UnaryInterceptor(server.AuthInterceptor))

	githubServer, err := server.NewGithubSearchServer() // Capture the returned error
	if err != nil {
		log.Fatalf("failed to create GithubSearchServer: %v", err)
		return
	}
	pb.RegisterGithubSearchServiceServer(s, githubServer)

	// Start the gRPC server in a separate goroutine
	go func() {
		log.Printf("gRPC server listening on port %d", port)
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for a signal to quit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Stopping gRPC server...")

	// Stop the gRPC server gracefully
	s.GracefulStop()
	log.Println("gRPC server stopped")
}
