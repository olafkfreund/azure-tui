package test

import (
	"testing"
)

// Mock model structure for testing
type mockModel struct {
	activeView      string
	navigationStack []string
}

// Implement navigation methods for testing
func (m *mockModel) pushView(newView string) {
	if m.activeView != newView {
		m.navigationStack = append(m.navigationStack, m.activeView)
		m.activeView = newView
	}
}

func (m *mockModel) popView() bool {
	if len(m.navigationStack) == 0 {
		return false
	}

	lastIndex := len(m.navigationStack) - 1
	previousView := m.navigationStack[lastIndex]
	m.navigationStack = m.navigationStack[:lastIndex]
	m.activeView = previousView

	return true
}

func (m *mockModel) clearNavigationStack() {
	m.navigationStack = []string{}
}

// TestNavigationStack tests the navigation stack functionality
func TestNavigationStack(t *testing.T) {
	model := &mockModel{
		activeView:      "welcome",
		navigationStack: []string{},
	}

	// Test pushing views
	model.pushView("dashboard")
	if model.activeView != "dashboard" {
		t.Errorf("Expected activeView to be 'dashboard', got '%s'", model.activeView)
	}
	if len(model.navigationStack) != 1 || model.navigationStack[0] != "welcome" {
		t.Errorf("Expected navigation stack to contain 'welcome', got %v", model.navigationStack)
	}

	// Test pushing another view
	model.pushView("network-dashboard")
	if model.activeView != "network-dashboard" {
		t.Errorf("Expected activeView to be 'network-dashboard', got '%s'", model.activeView)
	}
	if len(model.navigationStack) != 2 {
		t.Errorf("Expected navigation stack length to be 2, got %d", len(model.navigationStack))
	}

	// Test popping views
	success := model.popView()
	if !success {
		t.Error("Expected popView to succeed")
	}
	if model.activeView != "dashboard" {
		t.Errorf("Expected activeView to be 'dashboard' after pop, got '%s'", model.activeView)
	}
	if len(model.navigationStack) != 1 {
		t.Errorf("Expected navigation stack length to be 1 after pop, got %d", len(model.navigationStack))
	}

	// Test popping last view
	success = model.popView()
	if !success {
		t.Error("Expected popView to succeed")
	}
	if model.activeView != "welcome" {
		t.Errorf("Expected activeView to be 'welcome' after final pop, got '%s'", model.activeView)
	}
	if len(model.navigationStack) != 0 {
		t.Errorf("Expected navigation stack to be empty after final pop, got %v", model.navigationStack)
	}

	// Test popping from empty stack
	success = model.popView()
	if success {
		t.Error("Expected popView to fail when stack is empty")
	}

	// Test that pushing the same view doesn't add to stack
	model.activeView = "dashboard"
	model.navigationStack = []string{}
	model.pushView("dashboard")
	if len(model.navigationStack) != 0 {
		t.Error("Expected navigation stack to remain empty when pushing same view")
	}

	// Test clearing navigation stack
	model.pushView("network-dashboard")
	model.pushView("vnet-details")
	model.clearNavigationStack()
	if len(model.navigationStack) != 0 {
		t.Error("Expected navigation stack to be empty after clearNavigationStack")
	}

	t.Log("Navigation stack tests passed")
}
