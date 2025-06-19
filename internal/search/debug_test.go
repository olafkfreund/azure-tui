package search

import (
	"fmt"
	"testing"
)

// Debug test to understand tag parsing
func TestTagParsing(t *testing.T) {
	engine := NewSearchEngine()

	// Test query parsing
	query := "tag:env=production"
	parsedQuery := engine.parseQuery(query)

	fmt.Printf("Raw query: %s\n", parsedQuery.RawQuery)
	fmt.Printf("Is advanced: %t\n", parsedQuery.IsAdvanced)
	fmt.Printf("Terms: %v\n", parsedQuery.Terms)
	fmt.Printf("Filters.Tags: %+v\n", parsedQuery.Filters.Tags)

	// Test with sample resource
	resource := Resource{
		ID:            "test",
		Name:          "test-vm",
		Type:          "Microsoft.Compute/virtualMachines",
		Location:      "eastus",
		ResourceGroup: "test-rg",
		Tags:          map[string]string{"env": "production", "app": "web"},
	}

	// Test filter matching
	matches := engine.matchesFilters(resource, parsedQuery.Filters)
	fmt.Printf("Resource matches filters: %t\n", matches)

	// Test full search
	engine.SetResources([]Resource{resource})
	results, err := engine.Search(query)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Search results count: %d\n", len(results))
	for _, result := range results {
		fmt.Printf("Result: %s, MatchType: %s, Tags: %+v\n", result.ResourceName, result.MatchType, result.Tags)
	}
}
