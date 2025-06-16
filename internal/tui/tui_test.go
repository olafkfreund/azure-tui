package tui_test

import (
	"testing"

	"github.com/olafkfreund/azure-tui/internal/tui"
)

func TestRenderPopup(t *testing.T) {
	msg := tui.PopupMsg{Title: "Test", Content: "Hello", Level: "info"}
	out := tui.RenderPopup(msg)
	if out == "" {
		t.Error("Expected non-empty popup output")
	}
}

func TestRenderMatrixGraph(t *testing.T) {
	msg := tui.MatrixGraphMsg{
		Title:  "Test Graph",
		Rows:   [][]string{{"A", "B"}, {"C", "D"}},
		Labels: []string{"Col1", "Col2"},
	}
	out := tui.RenderMatrixGraph(msg)
	if out == "" {
		t.Error("Expected non-empty matrix graph output")
	}
}
