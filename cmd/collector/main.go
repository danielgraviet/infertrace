package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgraviet/infertrace/internal/collector"
	infertracepb "github.com/danielgraviet/infertrace/proto"
	"google.golang.org/grpc"
)

func main() {
	const address = ":50051"
	const workerCount = 4
	const queueSize = 1024

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", address, err)
	}

	pipeline := collector.NewCollector(workerCount, queueSize, nil)
	grpcServer := grpc.NewServer()
	infertracepb.RegisterCollectorServiceServer(grpcServer, collector.NewServer(pipeline))

	fmt.Printf("collector listening on %s\n", address)

	go func() {
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			log.Printf("gRPC server stopped with error: %v", serveErr)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("shutdown signal received, stopping collector")
	grpcServer.GracefulStop()
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pipeline.Stop(stopCtx); err != nil {
		log.Printf("collector pipeline stop error: %v", err)
	}
}
