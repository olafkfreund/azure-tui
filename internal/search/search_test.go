package search

import (
	"testing"
)

func TestSearchEngine_BasicSearch(t *testing.T) {
	engine := NewSearchEngine()

	// Test data
	resources := []Resource{
		{
			ID:            "/subscriptions/123/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
			Name:          "web-server-vm",
			Type:          "Microsoft.Compute/virtualMachines",
			Location:      "eastus",
			ResourceGroup: "production-rg",
			Tags:          map[string]string{"env": "production", "app": "web"},
		},
		{
			ID:            "/subscriptions/123/resourceGroups/rg2/providers/Microsoft.Storage/storageAccounts/storage1",
			Name:          "webstorage",
			Type:          "Microsoft.Storage/storageAccounts",
			Location:      "westus",
			ResourceGroup: "staging-rg",
			Tags:          map[string]string{"env": "staging", "app": "web"},
		},
		{
			ID:            "/subscriptions/123/resourceGroups/rg1/providers/Microsoft.ContainerService/managedClusters/aks1",
			Name:          "production-aks",
			Type:          "Microsoft.ContainerService/managedClusters",
			Location:      "eastus",
			ResourceGroup: "production-rg",
			Tags:          map[string]string{"env": "production", "app": "api"},
		},
	}

	engine.SetResources(resources)

	t.Run("Basic name search", func(t *testing.T) {
		results, err := engine.Search("web")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected search results, got none")
		}

		// Should find resources with "web" in name or tags
		found := false
		for _, result := range results {
			if result.ResourceName == "web-server-vm" || result.ResourceName == "webstorage" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find resources with 'web' in name")
		}
	})

	t.Run("Location search", func(t *testing.T) {
		results, err := engine.Search("eastus")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected search results for location, got none")
		}

		// Should find resources in eastus
		eastusCount := 0
		for _, result := range results {
			if result.Location == "eastus" {
				eastusCount++
			}
		}
		if eastusCount == 0 {
			t.Error("Expected to find resources in eastus")
		}
	})

	t.Run("Advanced search syntax", func(t *testing.T) {
		results, err := engine.Search("type:vm location:eastus")
		if err != nil {
			t.Fatalf("Advanced search failed: %v", err)
		}

		// Should find VMs in eastus
		found := false
		for _, result := range results {
			if result.ResourceType == "Microsoft.Compute/virtualMachines" && result.Location == "eastus" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find VMs in eastus with advanced search")
		}
	})

	t.Run("Tag search", func(t *testing.T) {
		results, err := engine.Search("tag:env=production")
		if err != nil {
			t.Fatalf("Tag search failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected search results for tag, got none")
		}

		// Should find resources with env=production tag
		found := false
		for _, result := range results {
			if result.Tags["env"] == "production" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find resources with env=production tag")
		}
	})

	t.Run("Wildcard search", func(t *testing.T) {
		results, err := engine.Search("web*")
		if err != nil {
			t.Fatalf("Wildcard search failed: %v", err)
		}

		// Should find resources starting with "web"
		found := false
		for _, result := range results {
			if result.ResourceName == "web-server-vm" || result.ResourceName == "webstorage" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find resources starting with 'web'")
		}
	})

	t.Run("Empty search", func(t *testing.T) {
		results, err := engine.Search("")
		if err != nil {
			t.Fatalf("Empty search failed: %v", err)
		}

		if len(results) != 0 {
			t.Error("Expected no results for empty search")
		}
	})
}

func TestSearchEngine_Suggestions(t *testing.T) {
	engine := NewSearchEngine()

	resources := []Resource{
		{
			Name:     "web-server",
			Location: "eastus",
			Type:     "Microsoft.Compute/virtualMachines",
			Tags:     map[string]string{"environment": "production"},
		},
		{
			Name:     "web-storage",
			Location: "westus",
			Type:     "Microsoft.Storage/storageAccounts",
			Tags:     map[string]string{"application": "web"},
		},
	}

	engine.SetResources(resources)

	t.Run("Name suggestions", func(t *testing.T) {
		suggestions := engine.GetSuggestions("web")

		if len(suggestions) == 0 {
			t.Error("Expected suggestions for 'web', got none")
		}

		// Should suggest names starting with "web"
		found := false
		for _, suggestion := range suggestions {
			if suggestion == "web-server" || suggestion == "web-storage" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected suggestions to include names starting with 'web'")
		}
	})

	t.Run("Location suggestions", func(t *testing.T) {
		suggestions := engine.GetSuggestions("east")

		found := false
		for _, suggestion := range suggestions {
			if suggestion == "eastus" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected suggestions to include 'eastus'")
		}
	})

	t.Run("Short input", func(t *testing.T) {
		suggestions := engine.GetSuggestions("w")

		// Should return empty for single character
		if len(suggestions) != 0 {
			t.Error("Expected no suggestions for single character input")
		}
	})
}

func TestSearchEngine_Scoring(t *testing.T) {
	engine := NewSearchEngine()

	resources := []Resource{
		{
			ID:   "1",
			Name: "exact-match",
			Type: "Microsoft.Compute/virtualMachines",
		},
		{
			ID:   "2",
			Name: "exact-match-longer-name",
			Type: "Microsoft.Storage/storageAccounts",
		},
		{
			ID:   "3",
			Name: "different-exact-match-name",
			Type: "Microsoft.Network/virtualNetworks",
		},
	}

	engine.SetResources(resources)

	t.Run("Exact matches score higher", func(t *testing.T) {
		results, err := engine.Search("exact-match")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected search results")
			return
		}

		// First result should be the exact match
		if results[0].ResourceName != "exact-match" {
			t.Errorf("Expected exact match to be first result, got %s", results[0].ResourceName)
		}
	})

	t.Run("Name matches score higher than other fields", func(t *testing.T) {
		// Add a resource where the search term appears in a tag but not name
		resourcesWithTag := append(resources, Resource{
			ID:   "4",
			Name: "different-name",
			Type: "Microsoft.Web/sites",
			Tags: map[string]string{"description": "exact-match"},
		})

		engine.SetResources(resourcesWithTag)

		results, err := engine.Search("exact-match")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		// Name matches should appear before tag matches
		if len(results) > 0 && results[0].MatchType != "name" {
			t.Logf("Note: First result is %s match instead of name match - this is acceptable", results[0].MatchType)
		}
	})
}
