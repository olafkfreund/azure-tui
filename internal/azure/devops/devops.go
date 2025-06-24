// Package devops provides Azure DevOps integration for the Azure TUI.
// It implements a borderless, tree-based interface for managing Azure DevOps
// organizations, projects, and pipelines, following vim-like navigation patterns.
package devops

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// DefaultConfig returns a default DevOps configuration
func DefaultConfig() DevOpsConfig {
	return DevOpsConfig{
		PersonalAccessToken: os.Getenv("AZURE_DEVOPS_PAT"),
		Organization:        os.Getenv("AZURE_DEVOPS_ORG"),
		Project:             os.Getenv("AZURE_DEVOPS_PROJECT"),
		BaseURL:             "https://dev.azure.com",
	}
}

// ValidateConfig validates the DevOps configuration
func ValidateConfig(config DevOpsConfig) error {
	if config.PersonalAccessToken == "" {
		return fmt.Errorf("Personal Access Token is required (set AZURE_DEVOPS_PAT environment variable)")
	}

	if config.Organization == "" {
		return fmt.Errorf("Organization is required (set AZURE_DEVOPS_ORG environment variable)")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://dev.azure.com"
	}

	return nil
}

// CreateManager creates a new DevOps manager with the given configuration
func CreateManager(config DevOpsConfig, width, height int) (*DevOpsManager, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	manager := NewDevOpsManager(config, width, height)

	if err := manager.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize DevOps manager: %w", err)
	}

	return manager, nil
}

// GetStatusIcon returns an icon for the given status
func GetStatusIcon(status string) string {
	switch status {
	case "succeeded", "success":
		return "✅"
	case "failed", "error":
		return "❌"
	case "running", "inprogress":
		return "🔄"
	case "canceled", "cancelled":
		return "⏹️"
	case "queued", "pending":
		return "⏳"
	default:
		return "❓"
	}
}

// GetPipelineTypeIcon returns an icon for the given pipeline type
func GetPipelineTypeIcon(pipelineType string) string {
	switch pipelineType {
	case "build":
		return "🔧"
	case "release":
		return "🚀"
	default:
		return "⚙️"
	}
}

// GetFormattedStatus returns a formatted status string with icon
func GetFormattedStatus(status string) string {
	return fmt.Sprintf("%s %s", GetStatusIcon(status), status)
}

// GetFormattedPipelineType returns a formatted pipeline type with icon
func GetFormattedPipelineType(pipelineType string) string {
	return fmt.Sprintf("%s %s", GetPipelineTypeIcon(pipelineType), pipelineType)
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(duration time.Duration) string {
	if duration == 0 {
		return "Unknown"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// IsValidStatus checks if a status is valid
func IsValidStatus(status string) bool {
	validStatuses := []string{"succeeded", "success", "failed", "error", "running", "inprogress", "canceled", "cancelled", "queued", "pending"}
	for _, valid := range validStatuses {
		if strings.ToLower(status) == valid {
			return true
		}
	}
	return false
}

// GetLastRunInfo returns formatted information about the last pipeline run
func GetLastRunInfo(pipeline Pipeline) string {
	if pipeline.LastRun == nil {
		return "No runs"
	}

	duration := FormatDuration(pipeline.LastRun.Duration)
	return fmt.Sprintf("Last: %s (%s)", GetFormattedStatus(pipeline.LastRun.Status), duration)
}
