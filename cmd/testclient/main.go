package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	proto "github.com/danielgraviet/infertrace/proto"
)

func main() {
	conn, err := grpc.NewClient("localhost:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewCollectorServiceClient(conn)

	resp, err := client.SendSpan(context.Background(), &proto.SendSpanRequest{
		span := NewSpan(
			
		)
	})
	if err != nil {
		log.Fatalf("SendSpan failed: %v", err)
	}

	fmt.Printf("accepted: %v\n", resp.Accepted)
}