package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tshubham2/catalog-proj/internal/services"
	pb "github.com/tshubham2/catalog-proj/proto/product/v1"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPath := spannerDatabasePath()
	client, err := spanner.NewClient(ctx, dbPath)
	if err != nil {
		log.Fatalf("failed to create spanner client: %v", err)
	}
	defer client.Close()

	container := services.NewContainer(client)

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, container.Handler)
	reflection.Register(grpcServer)

	port := envOrDefault("PORT", "50051")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("shutting down gRPC server...")
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Printf("gRPC server listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}

func spannerDatabasePath() string {
	project := envOrDefault("SPANNER_PROJECT", "test-project")
	instance := envOrDefault("SPANNER_INSTANCE", "test-instance")
	database := envOrDefault("SPANNER_DATABASE", "test-database")
	return fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, database)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
