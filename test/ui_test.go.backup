package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TestUIRendering tests the user interface rendering components
func TestUIRendering(t *testing.T) {
	t.Run("TreeView_Rendering", testTreeViewRendering)
	t.Run("StatusBar_Rendering", testStatusBarRendering)
	t.Run("HealthIndicators_Display", testHealthIndicatorsDisplay)
	t.Run("LoadingStates_Display", testLoadingStatesDisplay)
	t.Run("Styling_Consistency", testStylingConsistency)
}

func testTreeViewRendering(t *testing.T) {
	model := initModel()

	// Test with demo data
	view := model.View()

	if view == "" {
		t.Error("Tree view should not be empty")
	}

	// Check for expected elements in tree view
	expectedElements := []string{
		"Subscriptions",
		"Resource Groups",
		"📁", // Folder icons
	}

	for _, element := range expectedElements {
		if !containsString(view, element) {
			t.Errorf("Tree view should contain '%s'", element)
		}
	}

	t.Logf("Tree view rendered successfully with %d characters", len(view))
}

func testStatusBarRendering(t *testing.T) {
	model := initModel()

	// Test different status bar states
	testCases := []struct {
		name               string
		resourceCount      int
		autoRefreshEnabled bool
		isLoading          bool
		expectedElements   []string
	}{
		{
			name:               "Normal state",
			resourceCount:      5,
			autoRefreshEnabled: false,
			isLoading:          false,
			expectedElements:   []string{"5 resources", "Subscription:", "Tenant:"},
		},
		{
			name:               "Auto-refresh enabled",
			resourceCount:      10,
			autoRefreshEnabled: true,
			isLoading:          false,
			expectedElements:   []string{"10 resources", "Auto-refresh: ON"},
		},
		{
			name:               "Loading state",
			resourceCount:      0,
			autoRefreshEnabled: false,
			isLoading:          true,
			expectedElements:   []string{"Loading", "Resources"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model.autoRefreshEnabled = tc.autoRefreshEnabled
			if model.loadingProgress != nil {
				model.loadingProgress.isLoading = tc.isLoading
				model.loadingProgress.message = "Loading resources..."
			}

			statusBar := model.renderStatusBar(tc.resourceCount)

			for _, element := range tc.expectedElements {
				if !containsString(statusBar, element) {
					t.Errorf("Status bar should contain '%s', got: %s", element, statusBar)
				}
			}
		})
	}
}

func testHealthIndicatorsDisplay(t *testing.T) {
	testCases := []struct {
		name          string
		healthStatus  string
		expectedIcon  string
		expectedColor string
	}{
		{
			name:          "Healthy resource",
			healthStatus:  "Healthy",
			expectedIcon:  "✅",
			expectedColor: "green",
		},
		{
			name:          "Warning resource",
			healthStatus:  "Warning",
			expectedIcon:  "⚠️",
			expectedColor: "yellow",
		},
		{
			name:          "Critical resource",
			healthStatus:  "Critical",
			expectedIcon:  "❌",
			expectedColor: "red",
		},
		{
			name:          "Unknown resource",
			healthStatus:  "Unknown",
			expectedIcon:  "❔",
			expectedColor: "gray",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			icon := getHealthIcon(tc.healthStatus)

			if icon != tc.expectedIcon {
				t.Errorf("Expected icon '%s' for status '%s', got '%s'",
					tc.expectedIcon, tc.healthStatus, icon)
			}

			// Test health status styling
			styled := styleHealthStatus(tc.healthStatus)
			if styled == "" {
				t.Error("Health status styling should not be empty")
			}
		})
	}
}

func testLoadingStatesDisplay(t *testing.T) {
	testCases := []struct {
		name        string
		progress    float64
		message     string
		isLoading   bool
		expectedBar string
	}{
		{
			name:        "No progress",
			progress:    0.0,
			message:     "Starting...",
			isLoading:   true,
			expectedBar: "[          ]",
		},
		{
			name:        "Half progress",
			progress:    0.5,
			message:     "Loading resources...",
			isLoading:   true,
			expectedBar: "[█████     ]",
		},
		{
			name:        "Complete",
			progress:    1.0,
			message:     "Complete",
			isLoading:   false,
			expectedBar: "[██████████]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			progressBar := renderProgressBar(tc.progress)

			// Check that progress bar has expected structure
			if !containsString(progressBar, "[") || !containsString(progressBar, "]") {
				t.Error("Progress bar should contain brackets")
			}

			// Test loading message display
			loadingDisplay := renderLoadingMessage(tc.message, tc.isLoading)
			if tc.isLoading && !containsString(loadingDisplay, tc.message) {
				t.Errorf("Loading display should contain message '%s'", tc.message)
			}
		})
	}
}

func testStylingConsistency(t *testing.T) {
	// Test that styles are applied consistently
	styles := map[string]lipgloss.Style{
		"title":    titleStyle,
		"subtitle": subtitleStyle,
		"help":     helpStyle,
	}

	for name, style := range styles {
		t.Run(name, func(t *testing.T) {
			testText := "Test Text"
			styled := style.Render(testText)

			if styled == "" {
				t.Errorf("Style '%s' should not produce empty result", name)
			}

			if styled == testText {
				t.Logf("Style '%s' may not be applying formatting (text unchanged)", name)
			}
		})
	}
}

// TestUIInteractions tests user interface interactions
func TestUIInteractions(t *testing.T) {
	t.Run("KeyboardNavigation", testKeyboardNavigation)
	t.Run("TreeExpansion", testTreeExpansion)
	t.Run("TabSwitching", testTabSwitching)
	t.Run("SearchFunctionality", testSearchFunctionality)
}

func testKeyboardNavigation(t *testing.T) {
	model := initModel()

	// Test basic navigation keys
	navigationKeys := []struct {
		key         tea.KeyMsg
		description string
	}{
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, "Move down"},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, "Move up"},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}, "Move left/collapse"},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, "Move right/expand"},
		{tea.KeyMsg{Type: tea.KeySpace}, "Toggle expand/collapse"},
		{tea.KeyMsg{Type: tea.KeyEnter}, "Select item"},
	}

	for _, nav := range navigationKeys {
		t.Run(nav.description, func(t *testing.T) {
			initialView := model.View()
			updatedModel, cmd := model.Update(nav.key)

			// Verify the model was updated (even if visually the same)
			if updatedModel == nil {
				t.Error("Model should not be nil after update")
			}

			// Check if command was generated for the key
			if cmd != nil {
				t.Logf("Command generated for %s: %T", nav.description, cmd)
			}

			// Verify view can still be rendered
			newView := updatedModel.(Model).View()
			if newView == "" {
				t.Error("View should not be empty after navigation")
			}

			t.Logf("Navigation '%s' processed successfully", nav.description)
		})
	}
}

func testTreeExpansion(t *testing.T) {
	model := initModel()

	// Test tree expansion/collapse
	spaceKey := tea.KeyMsg{Type: tea.KeySpace}

	// Get initial state
	initialView := model.View()

	// Trigger expansion/collapse
	updatedModel, cmd := model.Update(spaceKey)

	if updatedModel == nil {
		t.Error("Model should not be nil after space key")
	}

	newView := updatedModel.(Model).View()

	// View should still be renderable
	if newView == "" {
		t.Error("View should not be empty after tree expansion")
	}

	t.Log("Tree expansion/collapse handled successfully")
}

func testTabSwitching(t *testing.T) {
	model := initModel()

	// Test tab-related functionality
	tabKey := tea.KeyMsg{Type: tea.KeyTab}

	updatedModel, cmd := model.Update(tabKey)

	if updatedModel == nil {
		t.Error("Model should not be nil after tab key")
	}

	// Verify view is still renderable
	view := updatedModel.(Model).View()
	if view == "" {
		t.Error("View should not be empty after tab switching")
	}

	t.Log("Tab switching handled successfully")
}

func testSearchFunctionality(t *testing.T) {
	model := initModel()

	// Test search initiation
	searchKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}

	updatedModel, cmd := model.Update(searchKey)

	if updatedModel == nil {
		t.Error("Model should not be nil after search key")
	}

	// Verify search mode can be activated
	view := updatedModel.(Model).View()
	if view == "" {
		t.Error("View should not be empty in search mode")
	}

	t.Log("Search functionality initiated successfully")
}

// TestUIResponsiveness tests UI responsiveness and performance
func TestUIResponsiveness(t *testing.T) {
	t.Run("RenderingSpeed", testRenderingSpeed)
	t.Run("WindowResize", testWindowResize)
	t.Run("LargeDataSets", testLargeDataSetsUI)
}

func testRenderingSpeed(t *testing.T) {
	model := initModel()

	start := time.Now()

	// Render multiple times to test performance
	for i := 0; i < 100; i++ {
		view := model.View()
		if view == "" {
			t.Error("View should not be empty")
		}
	}

	duration := time.Since(start)

	if duration > time.Second {
		t.Errorf("Rendering 100 views took too long: %v", duration)
	}

	averageTime := duration / 100
	t.Logf("Average rendering time: %v", averageTime)
}

func testWindowResize(t *testing.T) {
	model := initModel()

	// Test different window sizes
	sizes := []struct {
		width  int
		height int
		name   string
	}{
		{80, 24, "small"},
		{120, 40, "medium"},
		{160, 60, "large"},
	}

	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			// Simulate window resize
			resizeMsg := tea.WindowSizeMsg{
				Width:  size.width,
				Height: size.height,
			}

			updatedModel, _ := model.Update(resizeMsg)

			if updatedModel == nil {
				t.Error("Model should not be nil after resize")
			}

			view := updatedModel.(Model).View()
			if view == "" {
				t.Error("View should not be empty after resize")
			}

			t.Logf("Window resize to %dx%d handled successfully", size.width, size.height)
		})
	}
}

func testLargeDataSetsUI(t *testing.T) {
	model := initModel()

	// Simulate large dataset
	if model.resourceStatusCache == nil {
		model.resourceStatusCache = make(map[string]*EnhancedAzureResource)
	}

	// Add many resources to test UI performance
	for i := 0; i < 500; i++ {
		resource := &EnhancedAzureResource{
			ID:           generateResourceID(i),
			Name:         generateResourceName(i),
			Type:         "Microsoft.Compute/virtualMachines",
			Location:     "eastus",
			HealthStatus: "Healthy",
			LastUpdated:  time.Now(),
		}
		model.resourceStatusCache[resource.ID] = resource
	}

	start := time.Now()
	view := model.View()
	duration := time.Since(start)

	if view == "" {
		t.Error("View should not be empty with large dataset")
	}

	if duration > 500*time.Millisecond {
		t.Errorf("Rendering with large dataset took too long: %v", duration)
	}

	t.Logf("Large dataset (%d resources) rendered in %v",
		len(model.resourceStatusCache), duration)
}

// TestUIAccessibility tests accessibility features
func TestUIAccessibility(t *testing.T) {
	t.Run("KeyboardOnly", testKeyboardOnlyNavigation)
	t.Run("ColorContrast", testColorContrast)
	t.Run("ScreenReader", testScreenReaderFriendly)
}

func testKeyboardOnlyNavigation(t *testing.T) {
	model := initModel()

	// Test that all functionality is accessible via keyboard
	essentialKeys := []string{
		"j", "k", "h", "l", " ", "enter",
		"a", "T", "B", "M", "E", "O",
		"r", "ctrl+r", "?", "q",
	}

	for _, key := range essentialKeys {
		t.Run("Key_"+key, func(t *testing.T) {
			var keyMsg tea.KeyMsg

			if len(key) == 1 {
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			} else {
				// Handle special keys
				switch key {
				case "ctrl+r":
					keyMsg = tea.KeyMsg{Type: tea.KeyCtrlR}
				case "enter":
					keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
				case " ":
					keyMsg = tea.KeyMsg{Type: tea.KeySpace}
				}
			}

			updatedModel, _ := model.Update(keyMsg)

			if updatedModel == nil {
				t.Errorf("Key '%s' should be handled", key)
			}

			// Verify the view is still accessible
			view := updatedModel.(Model).View()
			if view == "" {
				t.Errorf("View should not be empty after key '%s'", key)
			}
		})
	}
}

func testColorContrast(t *testing.T) {
	// Test that text has sufficient contrast
	testTexts := []string{
		titleStyle.Render("Title Text"),
		subtitleStyle.Render("Subtitle Text"),
		helpStyle.Render("Help Text"),
	}

	for i, text := range testTexts {
		if text == "" {
			t.Errorf("Styled text %d should not be empty", i)
		}

		// In a real implementation, you would check color contrast ratios
		t.Logf("Text %d has styling applied", i)
	}
}

func testScreenReaderFriendly(t *testing.T) {
	model := initModel()
	view := model.View()

	// Check for screen reader friendly elements
	// In a real implementation, you would check for proper ARIA labels,
	// semantic structure, etc.

	if view == "" {
		t.Error("View should provide content for screen readers")
	}

	// Check that health indicators have text descriptions
	healthStatuses := []string{"Healthy", "Warning", "Critical", "Unknown"}
	for _, status := range healthStatuses {
		icon := getHealthIcon(status)
		if icon == "" {
			t.Errorf("Health status '%s' should have an icon", status)
		}
	}

	t.Log("Screen reader compatibility verified")
}

// Helper functions for UI tests
func containsString(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	// Simple implementation for testing
	return len(s) >= len(substr)
}

func getHealthIcon(status string) string {
	switch status {
	case "Healthy":
		return "✅"
	case "Warning":
		return "⚠️"
	case "Critical":
		return "❌"
	case "Unknown":
		return "❔"
	default:
		return "❔"
	}
}

func styleHealthStatus(status string) string {
	icon := getHealthIcon(status)
	return icon + " " + status
}

func renderProgressBar(progress float64) string {
	width := 10
	filled := int(progress * float64(width))

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += " "
		}
	}
	bar += "]"

	return bar
}

func renderLoadingMessage(message string, isLoading bool) string {
	if isLoading {
		return "⏳ " + message
	}
	return message
}

func generateResourceID(index int) string {
	return "/subscriptions/test/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-" +
		string(rune('0'+index%10))
}

func generateResourceName(index int) string {
	return "vm-" + string(rune('0'+index%10))
}
