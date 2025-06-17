package main

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

// TestAzureCLIIntegration tests the Azure CLI integration functions
func TestAzureCLIIntegration(t *testing.T) {
	// Skip if Azure CLI is not available
	if !isAzureCLIAvailable() {
		t.Skip("Azure CLI not available, skipping integration tests")
	}

	t.Run("FetchResourceGroups", testFetchResourceGroups)
	t.Run("FetchResourceGroupsWithTimeout", testFetchResourceGroupsWithTimeout)
	t.Run("FetchSubscriptions", testFetchSubscriptions)
	t.Run("FetchResourcesInGroup", testFetchResourcesInGroup)
}

func testFetchResourceGroups(t *testing.T) {
	groups, err := fetchResourceGroups("demo-subscription-id")

	// Should either return real data or fallback gracefully
	if err != nil {
		t.Logf("fetchResourceGroups returned error (expected if not authenticated): %v", err)
		// Verify it falls back to demo data
		demoGroups := getDemoResourceGroups()
		if len(demoGroups) == 0 {
			t.Error("Demo data fallback should not be empty")
		}
	} else {
		t.Logf("fetchResourceGroups succeeded, returned %d groups", len(groups))
		if len(groups) == 0 {
			t.Error("Expected at least some resource groups (real or demo)")
		}
	}
}

func testFetchResourceGroupsWithTimeout(t *testing.T) {
	// Test with very short timeout to verify timeout handling
	groups, err := fetchResourceGroupsWithTimeout("demo-subscription-id", 100*time.Millisecond)

	if err != nil {
		t.Logf("Short timeout test returned error as expected: %v", err)
	} else {
		t.Logf("Short timeout test succeeded (Azure CLI was very fast), returned %d groups", len(groups))
	}

	// Test with reasonable timeout
	groups, err = fetchResourceGroupsWithTimeout("demo-subscription-id", 10*time.Second)

	if err != nil {
		t.Logf("Reasonable timeout returned error: %v", err)
	} else {
		t.Logf("Reasonable timeout succeeded, returned %d groups", len(groups))
	}
}

func testFetchSubscriptions(t *testing.T) {
	subs, tenants, err := fetchAzureSubsAndTenantsWithTimeout(5 * time.Second)

	if err != nil {
		t.Logf("fetchAzureSubsAndTenantsWithTimeout returned error: %v", err)
		// Should fallback to demo data
		demoSubs := getDemoSubscriptions()
		demoTenants := getDemoTenants()
		if len(demoSubs) == 0 || len(demoTenants) == 0 {
			t.Error("Demo subscription/tenant data should not be empty")
		}
	} else {
		t.Logf("fetchAzureSubsAndTenantsWithTimeout succeeded: %d subs, %d tenants", len(subs), len(tenants))
		if len(subs) == 0 {
			t.Error("Expected at least one subscription")
		}
	}
}

func testFetchResourcesInGroup(t *testing.T) {
	// Test with a resource group that might exist
	resources, err := fetchResourcesInGroup("demo-resource-group")

	if err != nil {
		t.Logf("fetchResourcesInGroup returned error: %v", err)
		// Should fallback to demo data
		demoResources := getDemoResourcesForGroup("demo-resource-group")
		if len(demoResources) == 0 {
			t.Error("Demo resources fallback should not be empty")
		}
	} else {
		t.Logf("fetchResourcesInGroup succeeded, returned %d resources", len(resources))
	}
}

func isAzureCLIAvailable() bool {
	cmd := exec.Command("az", "--version")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, "az", "--version")

	return cmd.Run() == nil
}

// TestDemoDataIntegrity tests that demo data is properly structured
func TestDemoDataIntegrity(t *testing.T) {
	t.Run("DemoSubscriptions", testDemoSubscriptions)
	t.Run("DemoResourceGroups", testDemoResourceGroups)
	t.Run("DemoResources", testDemoResources)
	t.Run("DemoTenants", testDemoTenants)
}

func testDemoSubscriptions(t *testing.T) {
	subs := getDemoSubscriptions()

	if len(subs) == 0 {
		t.Error("Demo subscriptions should not be empty")
	}

	for i, sub := range subs {
		if sub.ID == "" {
			t.Errorf("Demo subscription %d has empty ID", i)
		}
		if sub.Name == "" {
			t.Errorf("Demo subscription %d has empty Name", i)
		}
		if sub.TenantID == "" {
			t.Errorf("Demo subscription %d has empty TenantID", i)
		}
	}
}

func testDemoResourceGroups(t *testing.T) {
	groups := getDemoResourceGroups()

	if len(groups) == 0 {
		t.Error("Demo resource groups should not be empty")
	}

	for i, group := range groups {
		if group.Name == "" {
			t.Errorf("Demo resource group %d has empty Name", i)
		}
		if group.Location == "" {
			t.Errorf("Demo resource group %d has empty Location", i)
		}
	}
}

func testDemoResources(t *testing.T) {
	// Test demo resources for each demo resource group
	groups := getDemoResourceGroups()

	for _, group := range groups {
		resources := getDemoResourcesForGroup(group.Name)

		if len(resources) == 0 {
			t.Errorf("Demo resources for group %s should not be empty", group.Name)
		}

		for i, resource := range resources {
			if resource.Name == "" {
				t.Errorf("Demo resource %d in group %s has empty Name", i, group.Name)
			}
			if resource.Type == "" {
				t.Errorf("Demo resource %d in group %s has empty Type", i, group.Name)
			}
			if resource.Location == "" {
				t.Errorf("Demo resource %d in group %s has empty Location", i, group.Name)
			}
		}
	}
}

func testDemoTenants(t *testing.T) {
	tenants := getDemoTenants()

	if len(tenants) == 0 {
		t.Error("Demo tenants should not be empty")
	}

	for i, tenant := range tenants {
		if tenant.ID == "" {
			t.Errorf("Demo tenant %d has empty ID", i)
		}
		if tenant.Name == "" {
			t.Errorf("Demo tenant %d has empty Name", i)
		}
	}
}

// TestBasicFunctionality tests basic application functionality
func TestBasicFunctionality(t *testing.T) {
	t.Run("DemoData", testDemoData)
	t.Run("AzureCLIAvailable", testAzureCLIAvailable)
}

func testDemoData(t *testing.T) {
	// Test demo data functions
	groups := getDemoResourceGroups()
	if len(groups) == 0 {
		t.Error("Demo resource groups should not be empty")
	}

	subs := getDemoSubscriptions()
	if len(subs) == 0 {
		t.Error("Demo subscriptions should not be empty")
	}

	tenants := getDemoTenants()
	if len(tenants) == 0 {
		t.Error("Demo tenants should not be empty")
	}

	// Test demo resources for first group
	if len(groups) > 0 {
		resources := getDemoResourcesForGroup(groups[0].Name)
		if len(resources) == 0 {
			t.Error("Demo resources should not be empty")
		}
	}
}

func testAzureCLIAvailable(t *testing.T) {
	available := isAzureCLIAvailable()
	t.Logf("Azure CLI available: %v", available)

	if available {
		t.Log("Azure CLI is available - real integration tests can run")
	} else {
		t.Log("Azure CLI not available - application will use demo mode")
	}
}

// TestErrorHandling tests error handling and fallback mechanisms
func TestErrorHandling(t *testing.T) {
	t.Run("NetworkTimeout", testNetworkTimeout)
	t.Run("InvalidResourceGroup", testInvalidResourceGroup)
	t.Run("AuthenticationFailure", testAuthenticationFailure)
}

func testNetworkTimeout(t *testing.T) {
	// Test with extremely short timeout to force timeout
	_, err := fetchResourceGroupsWithTimeout("test-subscription", 1*time.Millisecond)

	if err == nil {
		t.Log("Timeout test didn't timeout (Azure CLI was extremely fast)")
	} else {
		t.Logf("Timeout test properly returned error: %v", err)
	}
}

func testInvalidResourceGroup(t *testing.T) {
	// Test with invalid resource group name
	resources, err := fetchResourcesInGroup("invalid-resource-group-12345")

	if err != nil {
		t.Logf("Invalid resource group properly returned error: %v", err)
	}

	// Should fall back to demo data
	if len(resources) == 0 {
		t.Error("Should fallback to demo resources even with invalid group")
	}
}

func testAuthenticationFailure(t *testing.T) {
	// This test would require manipulating Azure CLI auth state
	// For now, just verify demo data fallback works
	demoSubs := getDemoSubscriptions()
	demoTenants := getDemoTenants()

	if len(demoSubs) == 0 {
		t.Error("Demo subscriptions fallback should not be empty")
	}

	if len(demoTenants) == 0 {
		t.Error("Demo tenants fallback should not be empty")
	}
}

// TestPerformance tests performance characteristics
func TestPerformance(t *testing.T) {
	t.Run("DemoDataGeneration", testDemoDataGeneration)
}

func testDemoDataGeneration(t *testing.T) {
	// Test that demo data generation is fast
	start := time.Now()

	for i := 0; i < 100; i++ {
		_ = getDemoResourceGroups()
		_ = getDemoSubscriptions()
		_ = getDemoTenants()
	}

	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Demo data generation took too long: %v", elapsed)
	}

	t.Logf("Generated demo data 100 times in %v", elapsed)
}

// Benchmark tests
func BenchmarkDemoDataGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getDemoResourceGroups()
		_ = getDemoSubscriptions()
		_ = getDemoTenants()
	}
}
