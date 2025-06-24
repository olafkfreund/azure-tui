package devops

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DevOpsClient handles Azure DevOps API interactions
type DevOpsClient struct {
	baseURL      string
	organization string
	project      string
	token        string
	httpClient   *http.Client
}

// NewDevOpsClient creates a new Azure DevOps API client
func NewDevOpsClient(config DevOpsConfig) *DevOpsClient {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://dev.azure.com"
	}

	return &DevOpsClient{
		baseURL:      baseURL,
		organization: config.Organization,
		project:      config.Project,
		token:        config.PersonalAccessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an authenticated request to Azure DevOps API
func (c *DevOpsClient) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add authentication header
	req.SetBasicAuth("", c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// ListOrganizations retrieves all accessible organizations
func (c *DevOpsClient) ListOrganizations() ([]Organization, error) {
	url := "https://app.vssps.visualstudio.com/_apis/accounts"

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("listing organizations: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var response struct {
		Value []Organization `json:"value"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing organizations: %w", err)
	}

	return response.Value, nil
}

// ListProjects retrieves all projects in the organization
func (c *DevOpsClient) ListProjects() ([]Project, error) {
	url := fmt.Sprintf("%s/%s/_apis/projects", c.baseURL, c.organization)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var response struct {
		Value []Project `json:"value"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing projects: %w", err)
	}

	return response.Value, nil
}

// ListBuildPipelines retrieves all build pipelines in the project
func (c *DevOpsClient) ListBuildPipelines() ([]Pipeline, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines", c.baseURL, c.organization, c.project)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("listing build pipelines: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var response struct {
		Value []Pipeline `json:"value"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing build pipelines: %w", err)
	}

	// Mark as build pipelines
	for i := range response.Value {
		response.Value[i].Type = "build"
	}

	return response.Value, nil
}

// ListReleasePipelines retrieves all release pipelines in the project
func (c *DevOpsClient) ListReleasePipelines() ([]Pipeline, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/release/definitions", c.baseURL, c.organization, c.project)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("listing release pipelines: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var response struct {
		Value []Pipeline `json:"value"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing release pipelines: %w", err)
	}

	// Mark as release pipelines
	for i := range response.Value {
		response.Value[i].Type = "release"
	}

	return response.Value, nil
}

// RunPipeline starts a new pipeline run
func (c *DevOpsClient) RunPipeline(pipelineID int, parameters map[string]interface{}) (*PipelineRun, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines/%d/runs", c.baseURL, c.organization, c.project, pipelineID)

	requestBody := map[string]interface{}{
		"templateParameters": parameters,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.makeRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("running pipeline: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var run PipelineRun
	if err := json.Unmarshal(body, &run); err != nil {
		return nil, fmt.Errorf("parsing pipeline run: %w", err)
	}

	return &run, nil
}

// GetPipelineRun retrieves details of a specific pipeline run
func (c *DevOpsClient) GetPipelineRun(pipelineID, runID int) (*PipelineRun, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines/%d/runs/%d", c.baseURL, c.organization, c.project, pipelineID, runID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("getting pipeline run: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var run PipelineRun
	if err := json.Unmarshal(body, &run); err != nil {
		return nil, fmt.Errorf("parsing pipeline run: %w", err)
	}

	// Calculate duration if both times are available
	if !run.StartTime.IsZero() && !run.FinishTime.IsZero() {
		run.Duration = run.FinishTime.Sub(run.StartTime)
	}

	return &run, nil
}

// ListPipelineRuns retrieves recent runs for a pipeline
func (c *DevOpsClient) ListPipelineRuns(pipelineID int, top int) ([]PipelineRun, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines/%d/runs?$top=%d", c.baseURL, c.organization, c.project, pipelineID, top)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("listing pipeline runs: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var response struct {
		Value []PipelineRun `json:"value"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing pipeline runs: %w", err)
	}

	// Calculate durations
	for i := range response.Value {
		run := &response.Value[i]
		if !run.StartTime.IsZero() && !run.FinishTime.IsZero() {
			run.Duration = run.FinishTime.Sub(run.StartTime)
		}
	}

	return response.Value, nil
}

// CancelPipelineRun cancels a running pipeline
func (c *DevOpsClient) CancelPipelineRun(pipelineID, runID int) error {
	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines/%d/runs/%d", c.baseURL, c.organization, c.project, pipelineID, runID)

	requestBody := map[string]interface{}{
		"state": "canceling",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.makeRequest("PATCH", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("canceling pipeline run: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetBuildLogs retrieves logs for a build
func (c *DevOpsClient) GetBuildLogs(buildID int) ([]LogEntry, error) {
	// First get the list of logs
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/logs", c.baseURL, c.organization, c.project, buildID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("getting build logs list: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var logsResponse struct {
		Value []struct {
			ID  int    `json:"id"`
			URL string `json:"url"`
		} `json:"value"`
	}

	if err := json.Unmarshal(body, &logsResponse); err != nil {
		return nil, fmt.Errorf("parsing logs list: %w", err)
	}

	var allLogs []LogEntry

	// Get content for each log
	for _, log := range logsResponse.Value {
		logURL := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/logs/%d",
			c.baseURL, c.organization, c.project, buildID, log.ID)

		logResp, err := c.makeRequest("GET", logURL, nil)
		if err != nil {
			continue // Skip failed log requests
		}

		logContent, err := io.ReadAll(logResp.Body)
		logResp.Body.Close()
		if err != nil {
			continue
		}

		// Parse log content into entries
		lines := strings.Split(string(logContent), "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				allLogs = append(allLogs, LogEntry{
					LineNumber: i + 1,
					Timestamp:  time.Now(), // Azure DevOps logs don't always have timestamps
					Message:    line,
					Level:      "INFO",
				})
			}
		}
	}

	return allLogs, nil
}

// TestConnection tests the connection to Azure DevOps
func (c *DevOpsClient) TestConnection() error {
	url := fmt.Sprintf("%s/%s/_apis/projects", c.baseURL, c.organization)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("testing connection: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
