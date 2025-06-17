package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	keys []string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		log.Printf("Key pressed: %s", key)
		m.keys = append(m.keys, key)
		if key == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Key Test - Press keys to test input (q to quit)\n\n"
	s += "Keys pressed:\n"
	for _, key := range m.keys {
		s += fmt.Sprintf("- %s\n", key)
	}
	s += "\nPress 'q' to quit"
	return s
}

func main() {
	// Enable debug logging
	debugFile, err := os.Create("/tmp/keytest-debug.log")
	if err != nil {
		log.Fatal(err)
	}
	defer debugFile.Close()
	log.SetOutput(debugFile)

	log.Println("Starting key test")

	m := model{keys: []string{}}
	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		log.Printf("Error: %v", err)
	}

	log.Println("Key test completed")
}
