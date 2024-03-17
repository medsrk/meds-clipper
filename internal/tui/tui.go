package tui

import (
	"context"
	"github.com/charmbracelet/bubbletea"
	"log"
	"meds-clip/internal/pb"
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

// Custom message types for handling gRPC responses and errors.
type historyMsg []string
type errMsg struct{ error }

func (m Model) Init() tea.Cmd {
	return fetchHistoryCmd(m.grpcClient.client)
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
