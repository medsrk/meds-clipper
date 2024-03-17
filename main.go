package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"meds-clipper/internal/clipboard"
	"meds-clipper/internal/pb"
	"meds-clipper/internal/store"
	"net"
	"os"
	"os/signal"
	"sync"
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

type subscriber struct {
	id      string
	updates chan<- *pb.ClipboardUpdate
}

var (
	subscribers = make(map[string]subscriber)
	mu          sync.Mutex
)

func (s *server) subscribeClipboardUpdates(req *pb.SubscriberRequest, stream pb.ClipboardService_SubscribeClipboardUpdatesServer) error {
	updates := make(chan *pb.ClipboardUpdate)
	sub := subscriber{id: "someUniqueID", updates: updates}

	mu.Lock()
	subscribers[sub.id] = sub
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(subscribers, sub.id)
		mu.Unlock()
		close(updates)
	}()

	for update := range updates {
		if err := stream.Send(update); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) notifySubscribers(items []string) {
	update := &pb.ClipboardUpdate{Items: items}

	mu.Lock()
	defer mu.Unlock()
	for _, sub := range subscribers {
		select {
		case sub.updates <- update:
		default:
			log.Println("Failed to send update to a subscriber")
		}
	}
}

func (s *server) NotifyClipboardUpdate(ctx context.Context, in *pb.NotifyRequest) (*pb.NotifyReply, error) {
	items, err := s.store.GetClipboardHistory()
	if err != nil {
		return nil, err
	}
	s.notifySubscribers(items)

	fmt.Printf("Notified %d subscribers\n", len(subscribers))
	return &pb.NotifyReply{Success: true}, nil
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
		_, err := clipboardClient.SaveClipboardItem(context.Background(), &pb.SaveRequest{Item: content})
		if err != nil {
			log.Printf("failed to save clipboard item via gRPC: %v", err)
			return
		}

		// Now trigger subscriber notifications via gRPC
		_, err = clipboardClient.NotifyClipboardUpdate(context.Background(), &pb.NotifyRequest{})
		if err != nil {
			log.Printf("failed to notify subscribers via gRPC: %v", err)
		}
	}

	if err := clipboard.StartClipboardWatcher(ctx, onChange); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start clipboard watcher: %v\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
	log.Println("Shutting down...")
}
