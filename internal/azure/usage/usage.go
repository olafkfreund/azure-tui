package usage

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type UsageMetric struct {
	Name  string `json:"name"`
	Value string `json:"currentValue"`
	Limit string `json:"limit"`
}

type Alarm struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details"`
}

func ListUsageMetrics(resourceID string) ([]UsageMetric, error) {
	cmd := exec.Command("az", "monitor", "metrics", "list", "--resource", resourceID, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Azure CLI returns an object with a "value" field containing the array
	var response struct {
		Value []struct {
			DisplayDescription string `json:"displayDescription"`
			ErrorCode          string `json:"errorCode"`
			ID                 string `json:"id"`
			Name               struct {
				Value          string `json:"value"`
				LocalizedValue string `json:"localizedValue"`
			} `json:"name"`
			Type string `json:"type"`
			Unit string `json:"unit"`
		} `json:"value"`
	}

	if err := json.Unmarshal(out, &response); err != nil {
		return nil, fmt.Errorf("failed to parse metrics response: %v", err)
	}

	// Convert to our UsageMetric format
	var metrics []UsageMetric
	for _, item := range response.Value {
		metric := UsageMetric{
			Name:  item.Name.LocalizedValue,
			Value: "N/A", // Metrics list doesn't provide current values, just metadata
			Limit: "N/A", // Metrics list doesn't provide limits
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func ListAlarms(resourceID string) ([]Alarm, error) {
	// Use the correct Azure CLI command for metric alerts
	cmd := exec.Command("az", "monitor", "metrics", "alert", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Azure CLI returns an array of alert rules
	var alertRules []struct {
		Name      string `json:"name"`
		Enabled   bool   `json:"enabled"`
		Severity  int    `json:"severity"`
		Condition struct {
			AllOf []struct {
				MetricName string `json:"metricName"`
				Operator   string `json:"operator"`
				Threshold  string `json:"threshold"`
			} `json:"allOf"`
		} `json:"condition"`
		WindowSize string `json:"windowSize"`
	}

	if err := json.Unmarshal(out, &alertRules); err != nil {
		return nil, err
	}

	// Convert to our Alarm format
	var alarms []Alarm
	for _, rule := range alertRules {
		status := "OK"
		if !rule.Enabled {
			status = "Disabled"
		}

		details := fmt.Sprintf("Severity: %d, Window: %s", rule.Severity, rule.WindowSize)
		if len(rule.Condition.AllOf) > 0 {
			details += fmt.Sprintf(", Metric: %s", rule.Condition.AllOf[0].MetricName)
		}

		alarm := Alarm{
			Name:    rule.Name,
			Status:  status,
			Details: details,
		}
		alarms = append(alarms, alarm)
	}

	return alarms, nil
}
