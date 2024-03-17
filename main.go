package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"meds-clip/internal/clipboard"
	"meds-clip/internal/pb"
	"meds-clip/internal/store"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type server struct {
	pb.UnimplementedClipboardServiceServer
	store store.Store
}

// SaveClipboardItem saves a new clipboard item to the store and returns success status.
func (s *server) SaveClipboardItem(ctx context.Context, in *pb.SaveRequest) (*pb.SaveReply, error) {
	err := s.store.SaveClipboardItem(in.GetItem())
	if err != nil {
		return nil, err
	}
	return &pb.SaveReply{Success: true}, nil
}

// GetClipboardHistory retrieves the clipboard history from the store.
func (s *server) GetClipboardHistory(ctx context.Context, in *pb.GetRequest) (*pb.GetReply, error) {
	items, err := s.store.GetClipboardHistory()
	if err != nil {
		return nil, err
	}
	return &pb.GetReply{Items: items}, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	boltStore, err := store.NewBoltStore("clipper.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create boltStore: %v\n", err)
		os.Exit(1)
	}

	go func() {
		lis, err := net.Listen("tcp", "localhost:50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterClipboardServiceServer(s, &server{store: boltStore})

		log.Println("gRPC server listening on localhost:50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Setup the gRPC client connection.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	clipboardClient := pb.NewClipboardServiceClient(conn)

	onChange := func(content string) {
		req, err := clipboardClient.SaveClipboardItem(context.Background(), &pb.SaveRequest{Item: content})
		if err != nil {
			log.Printf("failed to save clipboard item via gRPC: %v", err)
		} else {
			log.Printf("saved clipboard item via gRPC: %v", req)
			rep, err := clipboardClient.GetClipboardHistory(context.Background(), &pb.GetRequest{})
			if err != nil {
				log.Printf("failed to get clipboard history via gRPC: %v", err)
			} else {
				log.Printf("got clipboard history via gRPC: %v", rep)
			}
		}
	}

	if err := clipboard.StartClipboardWatcher(ctx, onChange); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start clipboard watcher: %v\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
	log.Println("Shutting down...")
}
