package tui

import (
	"context"
	"github.com/charmbracelet/bubbletea"
	"log"
	"meds-clipper/internal/pb"
	"strings"
)

type Model struct {
	grpcClient *GrpcClient
	history    []string
	cursor     int
}

func InitialModel(client *GrpcClient) *Model {
	return &Model{
		grpcClient: client,
	}
}

// Cmd to fetch clipboard history from the gRPC server.
func fetchHistoryCmd(client pb.ClipboardServiceClient) tea.Cmd {
	return func() tea.Msg {
		// context.Background() is used for simplicity; consider a more appropriate context.
		resp, err := client.GetClipboardHistory(context.Background(), &pb.GetRequest{})
		if err != nil {
			log.Printf("Failed to fetch clipboard history: %v", err)
			return errMsg{err}
		}
		return historyMsg(resp.Items)
	}
}

func subscribeToUpdatesCmd(client pb.ClipboardServiceClient) tea.Cmd {
	return func() tea.Msg {
		stream, err := client.SubscribeClipboardUpdates(context.Background(), &pb.SubscriberRequest{})
		if err != nil {
			return errMsg{err}
		}

		go func() {
			for {
				update, err := stream.Recv()
				if err != nil {
					// Log or handle the error. If the`` stream is closed, you might want to break or continue based on your use case.
					continue
				}
				// Here, instead of returning, you send the update directly to the model's Update function via a channel or similar mechanism.
				// Since you can't directly return a tea.Msg from within this goroutine, you'd typically have a message listener set up in your model.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 ~

			}
		}()

		// Since we're handling updates in a separate goroutine, we return nil here to indicate no immediate message needs to be processed.
		return nil
	}
}

// Custom message types for handling gRPC responses and errors.
type historyMsg []string
type errMsg struct{ error }

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchHistoryCmd(m.grpcClient.client),
		subscribeToUpdatesCmd(m.grpcClient.client),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case historyMsg:
		m.history = msg
		return m, nil
	case errMsg:
		log.Printf("Error: %v", msg)
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	for i, item := range m.history {
		if i == m.cursor {
			b.WriteString("> ")
		} else {
			b.WriteString("  ")
		}
		b.WriteString(item + "\n")
	}
	return b.String()
}

func (m Model) Load() tea.Cmd {
	return nil
}
