package tui

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"meds-clip/internal/pb"
	"time"
)

type GrpcClient struct {
	client pb.ClipboardServiceClient
}

func NewGrpcClient(serverAddr string) (*GrpcClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	client := pb.NewClipboardServiceClient(conn)

	return &GrpcClient{client: client}, nil
}

func (c *GrpcClient) FetchClipboardHistory(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetClipboardHistory(ctx, &pb.GetRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}
