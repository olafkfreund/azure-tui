package main

import (
	"testing"
	"time"
)

// TestBasicFunctionality tests basic application functionality
func TestBasicFunctionality(t *testing.T) {
	t.Run("Model_Creation", testModelCreation)
	t.Run("Resource_Validation", testResourceValidation)
	t.Run("Health_Calculation", testHealthCalculation)
}

func testModelCreation(t *testing.T) {
	model := initModel()

	if model.resourceStatusCache == nil {
		t.Error("Resource status cache should be initialized")
	}

	if model.treeView == nil {
		t.Error("Tree view should be initialized")
	}

	view := model.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	t.Log("Model creation test passed")
}

func testResourceValidation(t *testing.T) {
	resource := &EnhancedAzureResource{
		ID:           "/subscriptions/test/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/test-vm",
		Name:         "test-vm",
		Type:         "Microsoft.Compute/virtualMachines",
		Location:     "eastus",
		HealthStatus: "Healthy",
		LastUpdated:  time.Now(),
	}

	if resource.ID == "" {
		t.Error("Resource ID should not be empty")
	}

	if resource.Name == "" {
		t.Error("Resource name should not be empty")
	}

	t.Log("Resource validation test passed")
}

func testHealthCalculation(t *testing.T) {
	resources := []EnhancedAzureResource{
		{HealthStatus: "Healthy"},
		{HealthStatus: "Warning"},
		{HealthStatus: "Critical"},
		{HealthStatus: "Unknown"},
		{HealthStatus: "Healthy"},
	}

	healthy, warning, critical, unknown := calculateHealthCounts(resources)

	if healthy != 2 {
		t.Errorf("Expected 2 healthy resources, got %d", healthy)
	}

	if warning != 1 {
		t.Errorf("Expected 1 warning resource, got %d", warning)
	}

	if critical != 1 {
		t.Errorf("Expected 1 critical resource, got %d", critical)
	}

	if unknown != 1 {
		t.Errorf("Expected 1 unknown resource, got %d", unknown)
	}

	t.Log("Health calculation test passed")
}
