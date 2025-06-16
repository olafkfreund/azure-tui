package usage

import (
	"encoding/json"
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
	var metrics []UsageMetric
	if err := json.Unmarshal(out, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func ListAlarms(resourceID string) ([]Alarm, error) {
	cmd := exec.Command("az", "monitor", "alert", "list", "--resource", resourceID, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var alarms []Alarm
	if err := json.Unmarshal(out, &alarms); err != nil {
		return nil, err
	}
	return alarms, nil
}
