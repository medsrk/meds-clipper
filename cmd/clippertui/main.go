package main

import (
	"log"
	"meds-clip/internal/tui"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	grpcClient, err := tui.NewGrpcClient("localhost:50051")
	if err != nil {
		log.Fatalf("failed to create grpc client: %v", err)
	}

	m := tui.InitialModel(grpcClient)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatalf("error running program: %v", err)
	}

}
