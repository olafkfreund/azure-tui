package resourcedetails

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ResourceDetails represents detailed information about an Azure resource
type ResourceDetails struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Location      string                 `json:"location"`
	Tags          map[string]string      `json:"tags"`
	CreatedTime   string                 `json:"createdTime"`
	ModifiedTime  string                 `json:"modifiedTime"`
	Status        string                 `json:"status"`
	Properties    map[string]interface{} `json:"properties"`
	SKU           map[string]interface{} `json:"sku"`
	ResourceGroup string                 `json:"resourceGroup"`
}

// ResourceMetrics represents real-time metrics for a resource
type ResourceMetrics struct {
	ResourceID  string               `json:"resourceId"`
	CPUUsage    float64              `json:"cpuUsage"`
	MemoryUsage float64              `json:"memoryUsage"`
	NetworkIn   float64              `json:"networkIn"`
	NetworkOut  float64              `json:"networkOut"`
	DiskRead    float64              `json:"diskRead"`
	DiskWrite   float64              `json:"diskWrite"`
	Timestamp   time.Time            `json:"timestamp"`
	TrendData   map[string][]float64 `json:"trendData"`
}

// ResourceActions represents available actions for a resource
type ResourceActions struct {
	CanStart   bool `json:"canStart"`
	CanStop    bool `json:"canStop"`
	CanRestart bool `json:"canRestart"`
	CanConnect bool `json:"canConnect"`
	CanScale   bool `json:"canScale"`
	CanBackup  bool `json:"canBackup"`
}

// AKSDetails represents detailed information about an AKS cluster
type AKSDetails struct {
	ClusterInfo ResourceDetails        `json:"clusterInfo"`
	NodePools   []AKSNodePool          `json:"nodePools"`
	Pods        []KubernetesPod        `json:"pods"`
	Deployments []KubernetesDeployment `json:"deployments"`
	Services    []KubernetesService    `json:"services"`
	Namespaces  []string               `json:"namespaces"`
}

type AKSNodePool struct {
	Name   string `json:"name"`
	Count  int    `json:"count"`
	VMSize string `json:"vmSize"`
	OSType string `json:"osType"`
	Mode   string `json:"mode"`
	Status string `json:"provisioningState"`
}

type KubernetesPod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Ready     string `json:"ready"`
	Age       string `json:"age"`
	IP        string `json:"ip"`
}

type KubernetesDeployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	UpToDate  string `json:"upToDate"`
	Available string `json:"available"`
	Age       string `json:"age"`
}

type KubernetesService struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Type       string `json:"type"`
	ClusterIP  string `json:"clusterIP"`
	ExternalIP string `json:"externalIP"`
	Ports      string `json:"ports"`
	Age        string `json:"age"`
}

// GetResourceDetails fetches comprehensive details for a resource
func GetResourceDetails(resourceID string) (*ResourceDetails, error) {
	cmd := exec.Command("az", "resource", "show", "--ids", resourceID, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource details: %w", err)
	}

	var rawResource map[string]interface{}
	if err := json.Unmarshal(out, &rawResource); err != nil {
		return nil, fmt.Errorf("failed to parse resource details: %w", err)
	}

	details := &ResourceDetails{
		ID:       getStringValue(rawResource, "id"),
		Name:     getStringValue(rawResource, "name"),
		Type:     getStringValue(rawResource, "type"),
		Location: getStringValue(rawResource, "location"),
		Tags:     make(map[string]string),
	}

	// Extract resource group from ID
	if details.ID != "" {
		parts := strings.Split(details.ID, "/")
		for i, part := range parts {
			if part == "resourceGroups" && i+1 < len(parts) {
				details.ResourceGroup = parts[i+1]
				break
			}
		}
	}

	// Extract tags
	if tags, ok := rawResource["tags"].(map[string]interface{}); ok {
		for key, value := range tags {
			if strValue, ok := value.(string); ok {
				details.Tags[key] = strValue
			}
		}
	}

	// Extract properties
	if properties, ok := rawResource["properties"].(map[string]interface{}); ok {
		details.Properties = properties

		// Try to extract creation/modification times from properties
		if timeStr, ok := properties["timeCreated"].(string); ok {
			details.CreatedTime = timeStr
		}
		if timeStr, ok := properties["lastModified"].(string); ok {
			details.ModifiedTime = timeStr
		}

		// Extract status/provisioning state
		if status, ok := properties["provisioningState"].(string); ok {
			details.Status = status
		}
	}

	// Extract SKU information
	if sku, ok := rawResource["sku"].(map[string]interface{}); ok {
		details.SKU = sku
	}

	return details, nil
}

// GetResourceMetrics fetches real-time metrics for a resource
func GetResourceMetrics(resourceID string) (*ResourceMetrics, error) {
	// Get current metrics using Azure Monitor
	metrics := &ResourceMetrics{
		ResourceID: resourceID,
		Timestamp:  time.Now(),
		TrendData:  make(map[string][]float64),
	}

	// CPU Usage
	if cpuUsage, err := getMetricValue(resourceID, "Percentage CPU"); err == nil {
		metrics.CPUUsage = cpuUsage
	}

	// Memory Usage (if available)
	if memUsage, err := getMetricValue(resourceID, "Available Memory Bytes"); err == nil {
		// Convert to percentage (assuming standard VM sizes)
		metrics.MemoryUsage = 100 - (memUsage / (1024 * 1024 * 1024)) // Simplified calculation
	}

	// Network metrics
	if netIn, err := getMetricValue(resourceID, "Network In Total"); err == nil {
		metrics.NetworkIn = netIn / (1024 * 1024) // Convert to MB
	}
	if netOut, err := getMetricValue(resourceID, "Network Out Total"); err == nil {
		metrics.NetworkOut = netOut / (1024 * 1024) // Convert to MB
	}

	// Disk metrics
	if diskRead, err := getMetricValue(resourceID, "Disk Read Bytes"); err == nil {
		metrics.DiskRead = diskRead / (1024 * 1024) // Convert to MB
	}
	if diskWrite, err := getMetricValue(resourceID, "Disk Write Bytes"); err == nil {
		metrics.DiskWrite = diskWrite / (1024 * 1024) // Convert to MB
	}

	// Get trend data for the last 24 hours
	trends, err := getMetricTrends(resourceID)
	if err == nil {
		metrics.TrendData = trends
	}

	return metrics, nil
}

// GetResourceActions determines what actions are available for a resource
func GetResourceActions(resourceType string) ResourceActions {
	actions := ResourceActions{}

	switch {
	case strings.Contains(resourceType, "Microsoft.Compute/virtualMachines"):
		actions.CanStart = true
		actions.CanStop = true
		actions.CanRestart = true
		actions.CanConnect = true
		actions.CanBackup = true
	case strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters"):
		actions.CanStart = true
		actions.CanStop = true
		actions.CanScale = true
		actions.CanConnect = true
	case strings.Contains(resourceType, "Microsoft.Web/sites"):
		actions.CanStart = true
		actions.CanStop = true
		actions.CanRestart = true
		actions.CanScale = true
	case strings.Contains(resourceType, "Microsoft.Sql/servers"):
		actions.CanScale = true
		actions.CanBackup = true
		actions.CanConnect = true
	}

	return actions
}

// GetAKSDetails fetches comprehensive AKS cluster information
func GetAKSDetails(clusterName, resourceGroup string) (*AKSDetails, error) {
	details := &AKSDetails{}

	// Get cluster resource details
	resourceID := fmt.Sprintf("/subscriptions/{subscription}/resourceGroups/%s/providers/Microsoft.ContainerService/managedClusters/%s",
		resourceGroup, clusterName)

	clusterInfo, err := GetResourceDetails(resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster details: %w", err)
	}
	details.ClusterInfo = *clusterInfo

	// Get credentials for kubectl commands
	if err := getAKSCredentials(clusterName, resourceGroup); err != nil {
		return nil, fmt.Errorf("failed to get AKS credentials: %w", err)
	}

	// Get node pools
	if nodePools, err := getAKSNodePools(clusterName, resourceGroup); err == nil {
		details.NodePools = nodePools
	}

	// Get Kubernetes resources
	if namespaces, err := getKubernetesNamespaces(); err == nil {
		details.Namespaces = namespaces
	}

	if pods, err := getKubernetesPods(); err == nil {
		details.Pods = pods
	}

	if deployments, err := getKubernetesDeployments(); err == nil {
		details.Deployments = deployments
	}

	if services, err := getKubernetesServices(); err == nil {
		details.Services = services
	}

	return details, nil
}

// Helper functions

func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

func getMetricValue(resourceID, metricName string) (float64, error) {
	cmd := exec.Command("az", "monitor", "metrics", "list",
		"--resource", resourceID,
		"--metric", metricName,
		"--aggregation", "Average",
		"--interval", "PT1M",
		"--output", "json")

	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return 0, err
	}

	// Extract the latest metric value
	if value, exists := result["value"]; exists {
		if valueSlice, ok := value.([]interface{}); ok && len(valueSlice) > 0 {
			if metric, ok := valueSlice[0].(map[string]interface{}); ok {
				if timeseries, ok := metric["timeseries"].([]interface{}); ok && len(timeseries) > 0 {
					if ts, ok := timeseries[0].(map[string]interface{}); ok {
						if data, ok := ts["data"].([]interface{}); ok && len(data) > 0 {
							if dataPoint, ok := data[len(data)-1].(map[string]interface{}); ok {
								if average, ok := dataPoint["average"].(float64); ok {
									return average, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("metric value not found")
}

func getMetricTrends(resourceID string) (map[string][]float64, error) {
	trends := make(map[string][]float64)

	metrics := []string{"Percentage CPU", "Network In Total", "Network Out Total"}
	for _, metric := range metrics {
		cmd := exec.Command("az", "monitor", "metrics", "list",
			"--resource", resourceID,
			"--metric", metric,
			"--aggregation", "Average",
			"--interval", "PT1H",
			"--start-time", time.Now().Add(-24*time.Hour).Format(time.RFC3339),
			"--output", "json")

		if out, err := cmd.Output(); err == nil {
			var result map[string]interface{}
			if json.Unmarshal(out, &result) == nil {
				// Extract trend data points
				var values []float64
				// Parse the Azure Monitor response and extract values
				// This is a simplified version - you'd want more robust parsing
				trends[metric] = values
			}
		}
	}

	return trends, nil
}

func getAKSCredentials(clusterName, resourceGroup string) error {
	cmd := exec.Command("az", "aks", "get-credentials",
		"--name", clusterName,
		"--resource-group", resourceGroup,
		"--overwrite-existing")
	return cmd.Run()
}

func getAKSNodePools(clusterName, resourceGroup string) ([]AKSNodePool, error) {
	cmd := exec.Command("az", "aks", "nodepool", "list",
		"--cluster-name", clusterName,
		"--resource-group", resourceGroup,
		"--output", "json")

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var nodePools []AKSNodePool
	if err := json.Unmarshal(out, &nodePools); err != nil {
		return nil, err
	}

	return nodePools, nil
}

func getKubernetesNamespaces() ([]string, error) {
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	namespaces := strings.Fields(string(out))
	return namespaces, nil
}

func getKubernetesPods() ([]KubernetesPod, error) {
	cmd := exec.Command("kubectl", "get", "pods", "--all-namespaces", "-o", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
				PodIP string `json:"podIP"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	var pods []KubernetesPod
	for _, item := range result.Items {
		pod := KubernetesPod{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Status:    item.Status.Phase,
			IP:        item.Status.PodIP,
			Ready:     "1/1", // Simplified
			Age:       "1d",  // Simplified
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func getKubernetesDeployments() ([]KubernetesDeployment, error) {
	cmd := exec.Command("kubectl", "get", "deployments", "--all-namespaces", "-o", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Status struct {
				Replicas          int `json:"replicas"`
				ReadyReplicas     int `json:"readyReplicas"`
				UpdatedReplicas   int `json:"updatedReplicas"`
				AvailableReplicas int `json:"availableReplicas"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	var deployments []KubernetesDeployment
	for _, item := range result.Items {
		deployment := KubernetesDeployment{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Ready:     fmt.Sprintf("%d/%d", item.Status.ReadyReplicas, item.Status.Replicas),
			UpToDate:  fmt.Sprintf("%d", item.Status.UpdatedReplicas),
			Available: fmt.Sprintf("%d", item.Status.AvailableReplicas),
			Age:       "1d", // Simplified
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func getKubernetesServices() ([]KubernetesService, error) {
	cmd := exec.Command("kubectl", "get", "services", "--all-namespaces", "-o", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Spec struct {
				Type      string `json:"type"`
				ClusterIP string `json:"clusterIP"`
				Ports     []struct {
					Port       int    `json:"port"`
					TargetPort int    `json:"targetPort"`
					Protocol   string `json:"protocol"`
				} `json:"ports"`
			} `json:"spec"`
			Status struct {
				LoadBalancer struct {
					Ingress []struct {
						IP string `json:"ip"`
					} `json:"ingress"`
				} `json:"loadBalancer"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	var services []KubernetesService
	for _, item := range result.Items {
		service := KubernetesService{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Type:      item.Spec.Type,
			ClusterIP: item.Spec.ClusterIP,
			Age:       "1d", // Simplified
		}

		// Extract external IP
		if len(item.Status.LoadBalancer.Ingress) > 0 {
			service.ExternalIP = item.Status.LoadBalancer.Ingress[0].IP
		}

		// Extract ports
		var ports []string
		for _, port := range item.Spec.Ports {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", port.Port, port.TargetPort, port.Protocol))
		}
		service.Ports = strings.Join(ports, ",")

		services = append(services, service)
	}

	return services, nil
}
