package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type simpleModel struct {
	message string
	loading bool
}

func (m simpleModel) Init() tea.Cmd {
	fmt.Println("DEBUG: Init() called")
	return func() tea.Msg {
		fmt.Println("DEBUG: Starting async function")
		time.Sleep(1 * time.Second)
		fmt.Println("DEBUG: Async function completed")
		return "loaded"
	}
}

func (m simpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	fmt.Printf("DEBUG: Update() called with msg: %T\n", msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	case string:
		if msg == "loaded" {
			m.message = "Data loaded!"
			m.loading = false
		}
	}
	return m, nil
}

func (m simpleModel) View() string {
	fmt.Println("DEBUG: View() called")
	if m.loading {
		return "Loading...\n\nPress 'q' to quit"
	}
	return fmt.Sprintf("%s\n\nPress 'q' to quit", m.message)
}

func main() {
	fmt.Println("Starting simple debug test...")

	model := simpleModel{
		message: "Waiting for data...",
		loading: true,
	}

	fmt.Println("Creating tea program...")
	p := tea.NewProgram(model)

	fmt.Println("Starting program...")
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println("Program ended.")
}
