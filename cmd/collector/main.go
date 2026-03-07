package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgraviet/infertrace/internal/collector"
	"github.com/danielgraviet/infertrace/internal/queryapi"
	"github.com/danielgraviet/infertrace/internal/store"
	infertracepb "github.com/danielgraviet/infertrace/proto"
	"google.golang.org/grpc"
)

func main() {
	const grpcAddress = ":4317"
	const httpAddress = ":8080"
	const workerCount = 4
	const queueSize = 1024

	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcAddress, err)
	}

	latencyStore := store.NewLatencyStore(30*time.Minute, 10000)
	pipeline := collector.NewCollector(workerCount, queueSize, func(span collector.Span) {
		latencyStore.Add(span.ModelName, span.StartTimeUnixNano, span.DurationNanos)
	})
	grpcServer := grpc.NewServer()
	infertracepb.RegisterCollectorServiceServer(grpcServer, collector.NewServer(pipeline))
	httpServer := &http.Server{
		Addr:    httpAddress,
		Handler: queryapi.NewServer(latencyStore).Handler(),
	}

	fmt.Printf("collector listening on %s (gRPC)\n", grpcAddress)
	fmt.Printf("query API listening on %s (HTTP)\n", httpAddress)

	go func() {
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			log.Printf("gRPC server stopped with error: %v", serveErr)
		}
	}()
	go func() {
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Printf("HTTP server stopped with error: %v", serveErr)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("shutdown signal received, stopping collector")
	grpcServer.GracefulStop()
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()
	if err := httpServer.Shutdown(httpCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pipeline.Stop(stopCtx); err != nil {
		log.Printf("collector pipeline stop error: %v", err)
	}
}
