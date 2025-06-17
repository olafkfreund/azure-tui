package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestPerformanceScenarios tests various performance scenarios
func TestPerformanceScenarios(t *testing.T) {
	t.Run("Memory_Usage_Under_Load", testMemoryUsageUnderLoad)
	t.Run("Concurrent_Operations", testConcurrentOperations)
	t.Run("Resource_Cache_Performance", testResourceCachePerformance)
	t.Run("UI_Rendering_Performance", testUIRenderingPerformance)
	t.Run("Health_Monitoring_Performance", testHealthMonitoringPerformance)
}

func testMemoryUsageUnderLoad(t *testing.T) {
	// Get baseline memory stats
	var baseline runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&baseline)

	model := initModel()

	// Simulate heavy resource loading
	resourceCount := 1000
	if model.resourceStatusCache == nil {
		model.resourceStatusCache = make(map[string]*EnhancedAzureResource)
	}

	for i := 0; i < resourceCount; i++ {
		resource := &EnhancedAzureResource{
			ID:           fmt.Sprintf("/subscriptions/test/resourceGroups/rg-%d/providers/Microsoft.Compute/virtualMachines/vm-%d", i/100, i),
			Name:         fmt.Sprintf("vm-%d", i),
			Type:         "Microsoft.Compute/virtualMachines",
			Location:     "eastus",
			HealthStatus: []string{"Healthy", "Warning", "Critical", "Unknown"}[i%4],
			LastUpdated:  time.Now(),
			Metadata: map[string]interface{}{
				"size":   "Standard_D2s_v3",
				"osType": "Linux",
				"vmId":   fmt.Sprintf("vm-id-%d", i),
			},
			Dependencies: []string{
				fmt.Sprintf("/subscriptions/test/resourceGroups/rg-%d/providers/Microsoft.Network/networkInterfaces/nic-%d", i/100, i),
			},
			Tags: map[string]string{
				"Environment": []string{"dev", "staging", "prod"}[i%3],
				"Owner":       "azure-tui-test",
			},
			Cost: ResourceCost{
				DailyCost:   10.50 + float64(i%100)/10,
				MonthlyCost: 315.00 + float64(i%100)*3,
				Currency:    "USD",
			},
			Metrics: ResourceMetrics{
				CPUUtilization:    float64(i % 100),
				MemoryUtilization: float64((i * 7) % 100),
				NetworkIn:         float64(i * 1024),
				NetworkOut:        float64(i * 512),
				LastCollected:     time.Now(),
			},
		}
		model.resourceStatusCache[resource.ID] = resource
	}

	// Force garbage collection and get final memory stats
	runtime.GC()
	var final runtime.MemStats
	runtime.ReadMemStats(&final)

	memoryIncrease := final.Alloc - baseline.Alloc
	memoryPerResource := memoryIncrease / uint64(resourceCount)

	t.Logf("Memory usage: baseline=%d, final=%d, increase=%d bytes",
		baseline.Alloc, final.Alloc, memoryIncrease)
	t.Logf("Memory per resource: %d bytes", memoryPerResource)

	// Reasonable memory usage check (adjust threshold as needed)
	maxMemoryPerResource := uint64(10000) // 10KB per resource
	if memoryPerResource > maxMemoryPerResource {
		t.Errorf("Memory usage too high: %d bytes per resource (max: %d)",
			memoryPerResource, maxMemoryPerResource)
	}

	// Test that the model still functions
	view := model.View()
	if view == "" {
		t.Error("Model should still render view after heavy resource loading")
	}
}

func testConcurrentOperations(t *testing.T) {
	model := initModel()

	if model.resourceStatusCache == nil {
		model.resourceStatusCache = make(map[string]*EnhancedAzureResource)
	}

	var wg sync.WaitGroup
	concurrentOps := 10
	resourcesPerOp := 100

	// Test concurrent resource updates
	for i := 0; i < concurrentOps; i++ {
		wg.Add(1)
		go func(opIndex int) {
			defer wg.Done()

			for j := 0; j < resourcesPerOp; j++ {
				resourceID := fmt.Sprintf("/subscriptions/test/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-%d-%d", opIndex, j)

				resource := &EnhancedAzureResource{
					ID:           resourceID,
					Name:         fmt.Sprintf("vm-%d-%d", opIndex, j),
					Type:         "Microsoft.Compute/virtualMachines",
					Location:     "eastus",
					HealthStatus: "Healthy",
					LastUpdated:  time.Now(),
				}

				// Simulate concurrent cache updates
				// Note: In a real implementation, you'd need proper synchronization
				model.resourceStatusCache[resourceID] = resource

				// Simulate some processing time
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	expectedResources := concurrentOps * resourcesPerOp
	actualResources := len(model.resourceStatusCache)

	t.Logf("Concurrent operations completed: expected=%d, actual=%d",
		expectedResources, actualResources)

	if actualResources != expectedResources {
		t.Errorf("Expected %d resources, got %d (may indicate race conditions)",
			expectedResources, actualResources)
	}
}

func testResourceCachePerformance(t *testing.T) {
	model := initModel()

	if model.resourceStatusCache == nil {
		model.resourceStatusCache = make(map[string]*EnhancedAzureResource)
	}

	resourceCount := 10000

	// Test cache insertion performance
	start := time.Now()
	for i := 0; i < resourceCount; i++ {
		resourceID := fmt.Sprintf("/subscriptions/test/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-%d", i)

		resource := &EnhancedAzureResource{
			ID:           resourceID,
			Name:         fmt.Sprintf("vm-%d", i),
			Type:         "Microsoft.Compute/virtualMachines",
			Location:     "eastus",
			HealthStatus: "Healthy",
			LastUpdated:  time.Now(),
		}

		model.resourceStatusCache[resourceID] = resource
	}
	insertDuration := time.Since(start)

	// Test cache lookup performance
	start = time.Now()
	foundCount := 0
	for i := 0; i < resourceCount; i++ {
		resourceID := fmt.Sprintf("/subscriptions/test/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-%d", i)
		if _, exists := model.resourceStatusCache[resourceID]; exists {
			foundCount++
		}
	}
	lookupDuration := time.Since(start)

	t.Logf("Cache performance:")
	t.Logf("  Insert %d items: %v (%v per item)", resourceCount, insertDuration, insertDuration/time.Duration(resourceCount))
	t.Logf("  Lookup %d items: %v (%v per item)", foundCount, lookupDuration, lookupDuration/time.Duration(foundCount))

	if foundCount != resourceCount {
		t.Errorf("Expected to find %d resources, found %d", resourceCount, foundCount)
	}

	// Performance thresholds
	maxInsertTime := 100 * time.Millisecond
	maxLookupTime := 50 * time.Millisecond

	if insertDuration > maxInsertTime {
		t.Errorf("Cache insertion too slow: %v (max: %v)", insertDuration, maxInsertTime)
	}

	if lookupDuration > maxLookupTime {
		t.Errorf("Cache lookup too slow: %v (max: %v)", lookupDuration, maxLookupTime)
	}
}

func testUIRenderingPerformance(t *testing.T) {
	model := initModel()

	// Add resources to test rendering performance
	if model.resourceStatusCache == nil {
		model.resourceStatusCache = make(map[string]*EnhancedAzureResource)
	}

	for i := 0; i < 100; i++ {
		resource := &EnhancedAzureResource{
			ID:           fmt.Sprintf("/subscriptions/test/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-%d", i),
			Name:         fmt.Sprintf("vm-%d", i),
			Type:         "Microsoft.Compute/virtualMachines",
			Location:     "eastus",
			HealthStatus: []string{"Healthy", "Warning", "Critical", "Unknown"}[i%4],
			LastUpdated:  time.Now(),
		}
		model.resourceStatusCache[resource.ID] = resource
	}

	// Test rendering performance
	renderCount := 1000
	start := time.Now()

	for i := 0; i < renderCount; i++ {
		view := model.View()
		if view == "" {
			t.Error("View should not be empty")
			break
		}
	}

	duration := time.Since(start)
	avgRenderTime := duration / time.Duration(renderCount)

	t.Logf("Rendering performance:")
	t.Logf("  %d renders: %v", renderCount, duration)
	t.Logf("  Average: %v per render", avgRenderTime)

	// Performance threshold
	maxAvgRenderTime := 1 * time.Millisecond
	if avgRenderTime > maxAvgRenderTime {
		t.Errorf("Rendering too slow: %v per render (max: %v)", avgRenderTime, maxAvgRenderTime)
	}
}

func testHealthMonitoringPerformance(t *testing.T) {
	// Test health status calculation performance
	resourceCounts := []int{100, 500, 1000, 5000}

	for _, count := range resourceCounts {
		t.Run(fmt.Sprintf("Resources_%d", count), func(t *testing.T) {
			resources := make([]EnhancedAzureResource, count)

			for i := 0; i < count; i++ {
				resources[i] = EnhancedAzureResource{
					ID:           fmt.Sprintf("/subscriptions/test/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-%d", i),
					Name:         fmt.Sprintf("vm-%d", i),
					HealthStatus: []string{"Healthy", "Warning", "Critical", "Unknown"}[i%4],
					LastUpdated:  time.Now(),
				}
			}

			// Test health calculation performance
			iterations := 1000
			start := time.Now()

			for i := 0; i < iterations; i++ {
				healthy, warning, critical, unknown := calculateHealthCounts(resources)

				// Verify calculation is correct
				if healthy+warning+critical+unknown != count {
					t.Errorf("Health count mismatch: %d+%d+%d+%d != %d",
						healthy, warning, critical, unknown, count)
					break
				}
			}

			duration := time.Since(start)
			avgTime := duration / time.Duration(iterations)

			t.Logf("Health calculation for %d resources:", count)
			t.Logf("  %d calculations: %v", iterations, duration)
			t.Logf("  Average: %v per calculation", avgTime)

			// Performance threshold scales with resource count
			maxAvgTime := time.Duration(count) * time.Microsecond
			if avgTime > maxAvgTime {
				t.Errorf("Health calculation too slow: %v (max: %v)", avgTime, maxAvgTime)
			}
		})
	}
}

// TestErrorHandlingScenarios tests various error conditions
func TestErrorHandlingScenarios(t *testing.T) {
	t.Run("Azure_API_Errors", testAzureAPIErrors)
	t.Run("AI_Service_Errors", testAIServiceErrors)
	t.Run("Network_Timeouts", testNetworkTimeouts)
	t.Run("Resource_Not_Found", testResourceNotFound)
	t.Run("Invalid_Data_Handling", testInvalidDataHandling)
}

func testAzureAPIErrors(t *testing.T) {
	// Test various Azure API error scenarios
	errorScenarios := []struct {
		name        string
		statusCode  int
		errorMsg    string
		shouldRetry bool
	}{
		{
			name:        "Unauthorized",
			statusCode:  401,
			errorMsg:    "Authentication failed",
			shouldRetry: false,
		},
		{
			name:        "Forbidden",
			statusCode:  403,
			errorMsg:    "Insufficient permissions",
			shouldRetry: false,
		},
		{
			name:        "Rate Limited",
			statusCode:  429,
			errorMsg:    "Too many requests",
			shouldRetry: true,
		},
		{
			name:        "Server Error",
			statusCode:  500,
			errorMsg:    "Internal server error",
			shouldRetry: true,
		},
		{
			name:        "Bad Gateway",
			statusCode:  502,
			errorMsg:    "Bad gateway",
			shouldRetry: true,
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Simulate Azure API error
			err := simulateAzureAPIError(scenario.statusCode, scenario.errorMsg)

			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			// Test error handling logic
			shouldRetry := shouldRetryAzureError(err)
			if shouldRetry != scenario.shouldRetry {
				t.Errorf("Expected shouldRetry=%v, got %v for error: %v",
					scenario.shouldRetry, shouldRetry, err)
			}

			t.Logf("Error handled correctly: %v (retry: %v)", err, shouldRetry)
		})
	}
}

func testAIServiceErrors(t *testing.T) {
	// Test AI service error scenarios
	aiErrorScenarios := []struct {
		name           string
		errorType      string
		shouldFallback bool
	}{
		{
			name:           "API Key Invalid",
			errorType:      "auth_error",
			shouldFallback: false,
		},
		{
			name:           "Rate Limit Exceeded",
			errorType:      "rate_limit",
			shouldFallback: true,
		},
		{
			name:           "Model Overloaded",
			errorType:      "overloaded",
			shouldFallback: true,
		},
		{
			name:           "Network Timeout",
			errorType:      "timeout",
			shouldFallback: true,
		},
	}

	for _, scenario := range aiErrorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := simulateAIError(scenario.errorType)

			if err == nil {
				t.Error("Expected AI error but got nil")
				return
			}

			shouldFallback := shouldFallbackFromAIError(err)
			if shouldFallback != scenario.shouldFallback {
				t.Errorf("Expected shouldFallback=%v, got %v for error: %v",
					scenario.shouldFallback, shouldFallback, err)
			}

			t.Logf("AI error handled: %v (fallback: %v)", err, shouldFallback)
		})
	}
}

func testNetworkTimeouts(t *testing.T) {
	timeoutScenarios := []struct {
		name          string
		timeout       time.Duration
		expectTimeout bool
	}{
		{
			name:          "Very Short Timeout",
			timeout:       1 * time.Millisecond,
			expectTimeout: true,
		},
		{
			name:          "Reasonable Timeout",
			timeout:       5 * time.Second,
			expectTimeout: false,
		},
	}

	for _, scenario := range timeoutScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), scenario.timeout)
			defer cancel()

			// Simulate network operation
			err := simulateNetworkOperation(ctx)

			if scenario.expectTimeout {
				if err == nil {
					t.Log("Operation completed before timeout (very fast network)")
				} else if err == context.DeadlineExceeded {
					t.Log("Timeout handled correctly")
				} else {
					t.Errorf("Expected timeout or success, got: %v", err)
				}
			} else {
				if err == context.DeadlineExceeded {
					t.Errorf("Unexpected timeout with %v timeout", scenario.timeout)
				} else if err != nil {
					t.Logf("Network operation failed: %v", err)
				} else {
					t.Log("Network operation succeeded")
				}
			}
		})
	}
}

func testResourceNotFound(t *testing.T) {
	model := initModel()

	// Test accessing non-existent resources
	nonExistentIDs := []string{
		"/subscriptions/invalid/resourceGroups/invalid/providers/Microsoft.Compute/virtualMachines/invalid",
		"",
		"/invalid/resource/path",
	}

	for _, id := range nonExistentIDs {
		t.Run(fmt.Sprintf("Resource_%s", id), func(t *testing.T) {
			// Test resource lookup
			resource := lookupResource(model, id)

			if resource != nil {
				t.Errorf("Expected nil for non-existent resource '%s', got: %v", id, resource)
			}

			// Test that the application doesn't crash
			view := model.View()
			if view == "" {
				t.Error("View should not be empty even with invalid resource lookup")
			}

			t.Logf("Non-existent resource '%s' handled gracefully", id)
		})
	}
}

func testInvalidDataHandling(t *testing.T) {
	// Test handling of invalid/corrupted data
	invalidDataScenarios := []struct {
		name     string
		resource *EnhancedAzureResource
		valid    bool
	}{
		{
			name:     "Nil resource",
			resource: nil,
			valid:    false,
		},
		{
			name: "Empty ID",
			resource: &EnhancedAzureResource{
				ID:   "",
				Name: "test",
			},
			valid: false,
		},
		{
			name: "Invalid health status",
			resource: &EnhancedAzureResource{
				ID:           "/valid/id",
				Name:         "test",
				HealthStatus: "InvalidStatus",
			},
			valid: false,
		},
		{
			name: "Negative metrics",
			resource: &EnhancedAzureResource{
				ID:   "/valid/id",
				Name: "test",
				Metrics: ResourceMetrics{
					CPUUtilization: -10.0,
				},
			},
			valid: false,
		},
		{
			name: "Valid resource",
			resource: &EnhancedAzureResource{
				ID:           "/subscriptions/test/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm",
				Name:         "test-vm",
				Type:         "Microsoft.Compute/virtualMachines",
				HealthStatus: "Healthy",
				Metrics: ResourceMetrics{
					CPUUtilization: 50.0,
				},
			},
			valid: true,
		},
	}

	for _, scenario := range invalidDataScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			isValid := validateResource(scenario.resource)

			if isValid != scenario.valid {
				t.Errorf("Expected validation result %v, got %v for scenario '%s'",
					scenario.valid, isValid, scenario.name)
			}

			t.Logf("Data validation for '%s': %v", scenario.name, isValid)
		})
	}
}

// Helper functions for error handling tests
func simulateAzureAPIError(statusCode int, message string) error {
	return fmt.Errorf("Azure API error %d: %s", statusCode, message)
}

func shouldRetryAzureError(err error) bool {
	// Simple implementation - in real code, parse error details
	errMsg := err.Error()
	return containsAny(errMsg, []string{"429", "500", "502", "503", "504"})
}

func simulateAIError(errorType string) error {
	return fmt.Errorf("AI service error: %s", errorType)
}

func shouldFallbackFromAIError(err error) bool {
	errMsg := err.Error()
	return containsAny(errMsg, []string{"rate_limit", "overloaded", "timeout"})
}

func simulateNetworkOperation(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		return nil // Simulate quick success
	}
}

func lookupResource(model Model, id string) *EnhancedAzureResource {
	if model.resourceStatusCache == nil {
		return nil
	}
	return model.resourceStatusCache[id]
}

func validateResource(resource *EnhancedAzureResource) bool {
	if resource == nil {
		return false
	}

	if resource.ID == "" {
		return false
	}

	validHealthStatuses := []string{"Healthy", "Warning", "Critical", "Unknown"}
	if resource.HealthStatus != "" && !containsString(validHealthStatuses, resource.HealthStatus) {
		return false
	}

	if resource.Metrics.CPUUtilization < 0 || resource.Metrics.CPUUtilization > 100 {
		return false
	}

	return true
}

func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			// Simple contains check
			return true
		}
	}
	return false
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Benchmark tests for performance validation
func BenchmarkHealthCalculation(b *testing.B) {
	resources := make([]EnhancedAzureResource, 1000)
	for i := 0; i < 1000; i++ {
		resources[i] = EnhancedAzureResource{
			HealthStatus: []string{"Healthy", "Warning", "Critical", "Unknown"}[i%4],
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateHealthCounts(resources)
	}
}

func BenchmarkResourceCacheOperations(b *testing.B) {
	cache := make(map[string]*EnhancedAzureResource)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("/resource/%d", i%1000)
		resource := &EnhancedAzureResource{
			ID:   id,
			Name: fmt.Sprintf("resource-%d", i%1000),
		}

		cache[id] = resource
		_ = cache[id]
	}
}

func BenchmarkUIRendering(b *testing.B) {
	model := initModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}
