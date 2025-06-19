package search

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// SearchResult represents a single search match
type SearchResult struct {
	ResourceID    string
	ResourceName  string
	ResourceType  string
	Location      string
	ResourceGroup string
	Tags          map[string]string
	MatchType     string // "name", "location", "tag", "type", "resource_group"
	MatchText     string
	MatchValue    string
	Score         int // Relevance score (higher = better match)
}

// SearchFilters defines search filtering criteria
type SearchFilters struct {
	Location      string
	ResourceType  string
	Tags          map[string]string
	ResourceGroup string
	ExcludeTypes  []string
}

// SearchQuery represents a parsed search query
type SearchQuery struct {
	RawQuery   string
	Terms      []string
	Filters    SearchFilters
	IsAdvanced bool
	Wildcards  bool
}

// SearchEngine provides search functionality across Azure resources
type SearchEngine struct {
	resources []Resource
}

// Resource represents a searchable Azure resource
type Resource struct {
	ID            string
	Name          string
	Type          string
	Location      string
	ResourceGroup string
	Status        string
	Tags          map[string]string
	Properties    map[string]interface{}
}

// NewSearchEngine creates a new search engine instance
func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		resources: make([]Resource, 0),
	}
}

// SetResources updates the searchable resource list
func (se *SearchEngine) SetResources(resources []Resource) {
	se.resources = resources
}

// Search performs a comprehensive search across all resources
func (se *SearchEngine) Search(query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return []SearchResult{}, nil
	}

	parsedQuery := se.parseQuery(query)
	results := []SearchResult{}

	for _, resource := range se.resources {
		matches := se.searchResource(resource, parsedQuery)
		results = append(results, matches...)
	}

	// Sort by relevance score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// parseQuery parses the search query and extracts filters
func (se *SearchEngine) parseQuery(query string) SearchQuery {
	sq := SearchQuery{
		RawQuery: query,
		Terms:    []string{},
		Filters:  SearchFilters{Tags: make(map[string]string)},
	}

	// Check for advanced search syntax
	if strings.Contains(query, ":") || strings.Contains(query, "AND") || strings.Contains(query, "NOT") {
		sq.IsAdvanced = true
		return se.parseAdvancedQuery(query, sq)
	}

	// Simple query - split into terms
	terms := strings.Fields(strings.ToLower(query))
	for _, term := range terms {
		if strings.Contains(term, "*") || strings.Contains(term, "?") {
			sq.Wildcards = true
		}
		sq.Terms = append(sq.Terms, term)
	}

	return sq
}

// parseAdvancedQuery handles advanced search syntax
func (se *SearchEngine) parseAdvancedQuery(query string, sq SearchQuery) SearchQuery {
	parts := strings.Fields(query)

	for _, part := range parts {
		if strings.Contains(part, ":") {
			// Handle key:value syntax
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				key := strings.ToLower(kv[0])
				value := strings.ToLower(kv[1])

				switch key {
				case "type":
					sq.Filters.ResourceType = value
				case "location", "loc":
					sq.Filters.Location = value
				case "rg", "resourcegroup", "resource-group":
					sq.Filters.ResourceGroup = value
				case "tag":
					// Handle tag:key=value or tag:key
					if strings.Contains(value, "=") {
						tagParts := strings.SplitN(value, "=", 2)
						sq.Filters.Tags[tagParts[0]] = tagParts[1]
					} else {
						sq.Filters.Tags[value] = "" // Any value
					}
				case "name":
					sq.Terms = append(sq.Terms, value)
				}
			}
		} else if strings.ToUpper(part) == "AND" || strings.ToUpper(part) == "OR" || strings.ToUpper(part) == "NOT" {
			// Handle boolean operators (simplified for now)
			continue
		} else {
			// Regular search term
			sq.Terms = append(sq.Terms, strings.ToLower(part))
		}
	}

	return sq
}

// searchResource searches a single resource for matches
func (se *SearchEngine) searchResource(resource Resource, query SearchQuery) []SearchResult {
	results := []SearchResult{}

	// Apply filters first
	if !se.matchesFilters(resource, query.Filters) {
		return results
	}

	// If this is a filter-only query (no search terms), return the resource as a match
	if len(query.Terms) == 0 && query.IsAdvanced {
		results = append(results, SearchResult{
			ResourceID:    resource.ID,
			ResourceName:  resource.Name,
			ResourceType:  resource.Type,
			Location:      resource.Location,
			ResourceGroup: resource.ResourceGroup,
			Tags:          resource.Tags,
			MatchType:     "filter",
			MatchText:     "filter match",
			MatchValue:    "matches filters",
			Score:         100,
		})
		return results
	}

	// Search in resource name
	if matches := se.searchInText(resource.Name, query.Terms, query.Wildcards); len(matches) > 0 {
		for _, match := range matches {
			results = append(results, SearchResult{
				ResourceID:    resource.ID,
				ResourceName:  resource.Name,
				ResourceType:  resource.Type,
				Location:      resource.Location,
				ResourceGroup: resource.ResourceGroup,
				Tags:          resource.Tags,
				MatchType:     "name",
				MatchText:     match,
				MatchValue:    resource.Name,
				Score:         se.calculateScore("name", match, resource.Name),
			})
		}
	}

	// Search in location
	if matches := se.searchInText(resource.Location, query.Terms, query.Wildcards); len(matches) > 0 {
		for _, match := range matches {
			results = append(results, SearchResult{
				ResourceID:    resource.ID,
				ResourceName:  resource.Name,
				ResourceType:  resource.Type,
				Location:      resource.Location,
				ResourceGroup: resource.ResourceGroup,
				Tags:          resource.Tags,
				MatchType:     "location",
				MatchText:     match,
				MatchValue:    resource.Location,
				Score:         se.calculateScore("location", match, resource.Location),
			})
		}
	}

	// Search in resource type
	if matches := se.searchInText(resource.Type, query.Terms, query.Wildcards); len(matches) > 0 {
		for _, match := range matches {
			results = append(results, SearchResult{
				ResourceID:    resource.ID,
				ResourceName:  resource.Name,
				ResourceType:  resource.Type,
				Location:      resource.Location,
				ResourceGroup: resource.ResourceGroup,
				Tags:          resource.Tags,
				MatchType:     "type",
				MatchText:     match,
				MatchValue:    resource.Type,
				Score:         se.calculateScore("type", match, resource.Type),
			})
		}
	}

	// Search in resource group
	if matches := se.searchInText(resource.ResourceGroup, query.Terms, query.Wildcards); len(matches) > 0 {
		for _, match := range matches {
			results = append(results, SearchResult{
				ResourceID:    resource.ID,
				ResourceName:  resource.Name,
				ResourceType:  resource.Type,
				Location:      resource.Location,
				ResourceGroup: resource.ResourceGroup,
				Tags:          resource.Tags,
				MatchType:     "resource_group",
				MatchText:     match,
				MatchValue:    resource.ResourceGroup,
				Score:         se.calculateScore("resource_group", match, resource.ResourceGroup),
			})
		}
	}

	// Search in tags
	for tagKey, tagValue := range resource.Tags {
		// Search in tag keys
		if matches := se.searchInText(tagKey, query.Terms, query.Wildcards); len(matches) > 0 {
			for _, match := range matches {
				results = append(results, SearchResult{
					ResourceID:    resource.ID,
					ResourceName:  resource.Name,
					ResourceType:  resource.Type,
					Location:      resource.Location,
					ResourceGroup: resource.ResourceGroup,
					Tags:          resource.Tags,
					MatchType:     "tag",
					MatchText:     match,
					MatchValue:    fmt.Sprintf("%s=%s", tagKey, tagValue),
					Score:         se.calculateScore("tag", match, tagKey),
				})
			}
		}

		// Search in tag values
		if matches := se.searchInText(tagValue, query.Terms, query.Wildcards); len(matches) > 0 {
			for _, match := range matches {
				results = append(results, SearchResult{
					ResourceID:    resource.ID,
					ResourceName:  resource.Name,
					ResourceType:  resource.Type,
					Location:      resource.Location,
					ResourceGroup: resource.ResourceGroup,
					Tags:          resource.Tags,
					MatchType:     "tag",
					MatchText:     match,
					MatchValue:    fmt.Sprintf("%s=%s", tagKey, tagValue),
					Score:         se.calculateScore("tag", match, tagValue),
				})
			}
		}
	}

	return results
}

// matchesFilters checks if a resource matches the specified filters
func (se *SearchEngine) matchesFilters(resource Resource, filters SearchFilters) bool {
	if filters.ResourceType != "" && !se.matchesResourceType(resource.Type, filters.ResourceType) {
		return false
	}

	if filters.Location != "" && !se.matchesText(resource.Location, filters.Location, false) {
		return false
	}

	if filters.ResourceGroup != "" && !se.matchesText(resource.ResourceGroup, filters.ResourceGroup, false) {
		return false
	}

	// Check tag filters
	for filterKey, filterValue := range filters.Tags {
		found := false
		for tagKey, tagValue := range resource.Tags {
			if se.matchesText(tagKey, filterKey, false) {
				if filterValue == "" || se.matchesText(tagValue, filterValue, false) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	// Check exclude types
	for _, excludeType := range filters.ExcludeTypes {
		if se.matchesText(resource.Type, excludeType, false) {
			return false
		}
	}

	return true
}

// searchInText searches for terms in a text string
func (se *SearchEngine) searchInText(text string, terms []string, wildcards bool) []string {
	matches := []string{}
	textLower := strings.ToLower(text)

	for _, term := range terms {
		if se.matchesText(textLower, term, wildcards) {
			matches = append(matches, term)
		}
	}

	return matches
}

// matchesText checks if text matches a search term
func (se *SearchEngine) matchesText(text, term string, wildcards bool) bool {
	textLower := strings.ToLower(text)
	termLower := strings.ToLower(term)

	if wildcards {
		return se.matchesWildcard(textLower, termLower)
	}

	return strings.Contains(textLower, termLower)
}

// matchesResourceType checks if a resource type matches a search term with type aliases
func (se *SearchEngine) matchesResourceType(resourceType, searchTerm string) bool {
	resourceTypeLower := strings.ToLower(resourceType)
	searchTermLower := strings.ToLower(searchTerm)

	// First try exact match or contains
	if strings.Contains(resourceTypeLower, searchTermLower) {
		return true
	}

	// Handle common type aliases
	typeAliases := map[string][]string{
		"vm":       {"Microsoft.Compute/virtualMachines", "virtualmachine", "virtualmachines"},
		"storage":  {"Microsoft.Storage/storageAccounts", "storageaccount", "storageaccounts"},
		"aks":      {"Microsoft.ContainerService/managedClusters", "managedcluster", "managedclusters"},
		"network":  {"Microsoft.Network/virtualNetworks", "virtualnetwork", "virtualnetworks"},
		"keyvault": {"Microsoft.KeyVault/vaults", "vault", "vaults"},
		"sql":      {"Microsoft.Sql/servers", "server", "servers"},
		"acr":      {"Microsoft.ContainerRegistry/registries", "registry", "registries"},
		"aci":      {"Microsoft.ContainerInstance/containerGroups", "containergroup", "containergroups"},
		"webapp":   {"Microsoft.Web/sites", "site", "sites"},
		"function": {"Microsoft.Web/sites", "functionapp", "functions"},
	}

	// Check if search term matches any aliases
	if aliases, exists := typeAliases[searchTermLower]; exists {
		for _, alias := range aliases {
			if strings.Contains(resourceTypeLower, strings.ToLower(alias)) {
				return true
			}
		}
	}

	// Check reverse - if the resource type contains the simplified term
	typeParts := strings.Split(resourceType, "/")
	if len(typeParts) > 1 {
		simpleType := strings.ToLower(typeParts[len(typeParts)-1])
		if strings.Contains(simpleType, searchTermLower) {
			return true
		}
	}

	return false
}

// matchesWildcard performs wildcard matching
func (se *SearchEngine) matchesWildcard(text, pattern string) bool {
	// Simple wildcard implementation
	// * matches any sequence of characters
	// ? matches any single character

	if pattern == "*" {
		return true
	}

	// Convert to regex-like matching
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = strings.ReplaceAll(pattern, "?", ".")

	// For simplicity, use contains for now
	if strings.Contains(pattern, ".*") {
		parts := strings.Split(pattern, ".*")
		for _, part := range parts {
			if part != "" && !strings.Contains(text, part) {
				return false
			}
		}
		return true
	}

	return strings.Contains(text, pattern)
}

// calculateScore calculates relevance score for a match
func (se *SearchEngine) calculateScore(matchType, matchTerm, fullText string) int {
	baseScore := 100

	// Boost exact matches
	if strings.EqualFold(matchTerm, fullText) {
		baseScore += 1000
	}

	// Boost prefix matches
	if strings.HasPrefix(strings.ToLower(fullText), strings.ToLower(matchTerm)) {
		baseScore += 500
	}

	// Different match types have different base scores
	switch matchType {
	case "name":
		baseScore += 800 // Names are most important
	case "type":
		baseScore += 600
	case "resource_group":
		baseScore += 400
	case "location":
		baseScore += 300
	case "tag":
		baseScore += 200
	}

	// Shorter matches score higher (more specific)
	lengthPenalty := len(fullText) / 10
	baseScore -= lengthPenalty

	if baseScore < 1 {
		baseScore = 1
	}

	return baseScore
}

// GetSuggestions provides search suggestions based on available resources
func (se *SearchEngine) GetSuggestions(partial string) []string {
	suggestions := make(map[string]bool)
	partialLower := strings.ToLower(partial)

	if len(partialLower) < 2 {
		return []string{}
	}

	for _, resource := range se.resources {
		// Suggest resource names
		if strings.HasPrefix(strings.ToLower(resource.Name), partialLower) {
			suggestions[resource.Name] = true
		}

		// Suggest locations
		if strings.HasPrefix(strings.ToLower(resource.Location), partialLower) {
			suggestions[resource.Location] = true
		}

		// Suggest resource types (simplified)
		typeParts := strings.Split(resource.Type, "/")
		if len(typeParts) > 1 {
			simpleType := strings.ToLower(typeParts[len(typeParts)-1])
			if strings.HasPrefix(simpleType, partialLower) {
				suggestions[simpleType] = true
			}
		}

		// Suggest tag keys
		for tagKey := range resource.Tags {
			if strings.HasPrefix(strings.ToLower(tagKey), partialLower) {
				suggestions[tagKey] = true
			}
		}
	}

	result := make([]string, 0, len(suggestions))
	for suggestion := range suggestions {
		result = append(result, suggestion)
	}

	sort.Strings(result)
	if len(result) > 10 {
		result = result[:10] // Limit suggestions
	}

	return result
}

// normalizeText normalizes text for better matching
func normalizeText(text string) string {
	// Convert to lowercase and remove extra spaces
	text = strings.ToLower(strings.TrimSpace(text))

	// Remove non-alphanumeric characters except spaces and common punctuation
	result := strings.Builder{}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
