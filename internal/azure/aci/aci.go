package aci

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/olafkfreund/azure-tui/internal/tui"
)

// Enhanced Container Instance structure matching Azure CLI output
type ContainerInstance struct {
	Name                 string                 `json:"name"`
	Location             string                 `json:"location"`
	ResourceGroup        string                 `json:"resourceGroup"`
	ID                   string                 `json:"id"`
	Type                 string                 `json:"type"`
	ProvisioningState    string                 `json:"provisioningState"`
	OSType               string                 `json:"osType"`
	RestartPolicy        string                 `json:"restartPolicy"`
	SKU                  string                 `json:"sku"`
	IPAddress            *IPAddress             `json:"ipAddress"`
	Containers           []Container            `json:"containers"`
	InitContainers       []Container            `json:"initContainers"`
	Volumes              []Volume               `json:"volumes"`
	Diagnostics          *Diagnostics           `json:"diagnostics"`
	NetworkProfile       *NetworkProfile        `json:"networkProfile"`
	DnsConfig            *DnsConfig             `json:"dnsConfig"`
	EncryptionProperties *EncryptionProperties  `json:"encryptionProperties"`
	Tags                 map[string]interface{} `json:"tags"`
	Priority             string                 `json:"priority"`
	Zones                []string               `json:"zones"`
}

type IPAddress struct {
	Type         string `json:"type"`
	IP           string `json:"ip"`
	FQDN         string `json:"fqdn"`
	DnsNameLabel string `json:"dnsNameLabel"`
	Ports        []Port `json:"ports"`
}

type Port struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type Container struct {
	Name                 string                `json:"name"`
	Image                string                `json:"image"`
	Command              []string              `json:"command"`
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables"`
	Resources            Resources             `json:"resources"`
	Ports                []Port                `json:"ports"`
	VolumeMounts         []VolumeMount         `json:"volumeMounts"`
	LivenessProbe        *Probe                `json:"livenessProbe"`
	ReadinessProbe       *Probe                `json:"readinessProbe"`
	InstanceView         *InstanceView         `json:"instanceView"`
}

type EnvironmentVariable struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	SecureValue string `json:"secureValue"`
}

type Resources struct {
	Requests *ResourceRequests `json:"requests"`
	Limits   *ResourceLimits   `json:"limits"`
}

type ResourceRequests struct {
	CPU        float64      `json:"cpu"`
	MemoryInGB float64      `json:"memoryInGb"`
	GPU        *GPUResource `json:"gpu"`
}

type ResourceLimits struct {
	CPU        float64      `json:"cpu"`
	MemoryInGB float64      `json:"memoryInGb"`
	GPU        *GPUResource `json:"gpu"`
}

type GPUResource struct {
	Count int    `json:"count"`
	SKU   string `json:"sku"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	ReadOnly  bool   `json:"readOnly"`
}

type Volume struct {
	Name      string     `json:"name"`
	AzureFile *AzureFile `json:"azureFile"`
	EmptyDir  *EmptyDir  `json:"emptyDir"`
	Secret    *Secret    `json:"secret"`
	GitRepo   *GitRepo   `json:"gitRepo"`
}

type AzureFile struct {
	ShareName          string `json:"shareName"`
	StorageAccountName string `json:"storageAccountName"`
	StorageAccountKey  string `json:"storageAccountKey"`
	ReadOnly           bool   `json:"readOnly"`
}

type EmptyDir struct{}

type Secret struct {
	SecretType string            `json:"secretType"`
	Items      map[string]string `json:"items"`
}

type GitRepo struct {
	Repository string `json:"repository"`
	Directory  string `json:"directory"`
	Revision   string `json:"revision"`
}

type Probe struct {
	HTTPGet             *HTTPGetAction `json:"httpGet"`
	Exec                *ExecAction    `json:"exec"`
	InitialDelaySeconds int            `json:"initialDelaySeconds"`
	PeriodSeconds       int            `json:"periodSeconds"`
	FailureThreshold    int            `json:"failureThreshold"`
	SuccessThreshold    int            `json:"successThreshold"`
	TimeoutSeconds      int            `json:"timeoutSeconds"`
}

type HTTPGetAction struct {
	Path   string `json:"path"`
	Port   int    `json:"port"`
	Scheme string `json:"scheme"`
}

type ExecAction struct {
	Command []string `json:"command"`
}

type InstanceView struct {
	RestartCount  int              `json:"restartCount"`
	CurrentState  *ContainerState  `json:"currentState"`
	PreviousState *ContainerState  `json:"previousState"`
	Events        []ContainerEvent `json:"events"`
}

type ContainerState struct {
	State        string    `json:"state"`
	StartTime    time.Time `json:"startTime"`
	ExitCode     int       `json:"exitCode"`
	FinishTime   time.Time `json:"finishTime"`
	DetailStatus string    `json:"detailStatus"`
}

type ContainerEvent struct {
	Count          int       `json:"count"`
	FirstTimestamp time.Time `json:"firstTimestamp"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
	Name           string    `json:"name"`
	Message        string    `json:"message"`
	Type           string    `json:"type"`
}

type Diagnostics struct {
	LogAnalytics *LogAnalytics `json:"logAnalytics"`
}

type LogAnalytics struct {
	WorkspaceID         string                 `json:"workspaceId"`
	WorkspaceKey        string                 `json:"workspaceKey"`
	LogType             string                 `json:"logType"`
	Metadata            map[string]interface{} `json:"metadata"`
	WorkspaceResourceID string                 `json:"workspaceResourceId"`
}

type NetworkProfile struct {
	ID string `json:"id"`
}

type DnsConfig struct {
	NameServers   []string `json:"nameServers"`
	SearchDomains []string `json:"searchDomains"`
	Options       string   `json:"options"`
}

type EncryptionProperties struct {
	VaultBaseURL string `json:"vaultBaseUrl"`
	KeyName      string `json:"keyName"`
	KeyVersion   string `json:"keyVersion"`
}

// =============================================================================
// CONTAINER INSTANCE MANAGEMENT FUNCTIONS
// =============================================================================

func ListContainerInstances() ([]ContainerInstance, error) {
	cmd := exec.Command("az", "container", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var containers []ContainerInstance
	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, err
	}
	return containers, nil
}

func GetContainerInstanceDetails(name, resourceGroup string) (*ContainerInstance, error) {
	cmd := exec.Command("az", "container", "show", "--name", name, "--resource-group", resourceGroup, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var container ContainerInstance
	if err := json.Unmarshal(out, &container); err != nil {
		return nil, err
	}
	return &container, nil
}

func CreateContainerInstance(name, group, location, image string) error {
	return exec.Command("az", "container", "create", "--name", name, "--resource-group", group, "--location", location, "--image", image).Run()
}

func DeleteContainerInstance(name, group string) error {
	return exec.Command("az", "container", "delete", "--name", name, "--resource-group", group, "--yes").Run()
}

func StartContainerInstance(name, resourceGroup string) error {
	return exec.Command("az", "container", "start", "--name", name, "--resource-group", resourceGroup).Run()
}

func StopContainerInstance(name, resourceGroup string) error {
	return exec.Command("az", "container", "stop", "--name", name, "--resource-group", resourceGroup).Run()
}

func RestartContainerInstance(name, resourceGroup string) error {
	return exec.Command("az", "container", "restart", "--name", name, "--resource-group", resourceGroup).Run()
}

func GetContainerLogs(name, resourceGroup string, containerName string, tail int) (string, error) {
	args := []string{"container", "logs", "--name", name, "--resource-group", resourceGroup, "--output", "tsv"}
	if containerName != "" {
		args = append(args, "--container-name", containerName)
	}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	cmd := exec.Command("az", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ExecIntoContainer(name, resourceGroup, containerName, command string) error {
	args := []string{"container", "exec", "--exec-command", command, "--name", name, "--resource-group", resourceGroup}
	if containerName != "" {
		args = append(args, "--container-name", containerName)
	}

	cmd := exec.Command("az", args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func AttachToContainer(name, resourceGroup, containerName string) error {
	args := []string{"container", "attach", "--name", name, "--resource-group", resourceGroup}
	if containerName != "" {
		args = append(args, "--container-name", containerName)
	}

	cmd := exec.Command("az", args...)
	return cmd.Run()
}

func UpdateContainerInstance(name, resourceGroup string, cpu float64, memory float64) error {
	args := []string{"container", "update", "--name", name, "--resource-group", resourceGroup}
	if cpu > 0 {
		args = append(args, "--cpu", fmt.Sprintf("%.1f", cpu))
	}
	if memory > 0 {
		args = append(args, "--memory", fmt.Sprintf("%.1f", memory))
	}

	return exec.Command("az", args...).Run()
}

// =============================================================================
// CONTAINER INSTANCE ANALYSIS AND RENDERING
// =============================================================================

func RenderContainerInstanceDetails(name, resourceGroup string) string {
	container, err := GetContainerInstanceDetails(name, resourceGroup)
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "Container Instance Details Error",
			Content: err.Error(),
			Level:   "error",
		})
	}

	rows := [][]string{}
	rows = append(rows, []string{"Property", "Value"})

	// Basic Information
	rows = append(rows, []string{"Name", container.Name})
	rows = append(rows, []string{"Resource Group", container.ResourceGroup})
	rows = append(rows, []string{"Location", container.Location})
	rows = append(rows, []string{"Provisioning State", container.ProvisioningState})
	rows = append(rows, []string{"OS Type", container.OSType})
	rows = append(rows, []string{"SKU", container.SKU})
	rows = append(rows, []string{"Restart Policy", container.RestartPolicy})

	// IP Address Information
	if container.IPAddress != nil {
		rows = append(rows, []string{"", ""}) // Spacer
		rows = append(rows, []string{"IP ADDRESS", ""})
		rows = append(rows, []string{"Type", container.IPAddress.Type})
		if container.IPAddress.IP != "" {
			rows = append(rows, []string{"Public IP", container.IPAddress.IP})
		}
		if container.IPAddress.FQDN != "" {
			rows = append(rows, []string{"FQDN", container.IPAddress.FQDN})
		}
		if container.IPAddress.DnsNameLabel != "" {
			rows = append(rows, []string{"DNS Name Label", container.IPAddress.DnsNameLabel})
		}

		// Ports
		if len(container.IPAddress.Ports) > 0 {
			portStrings := []string{}
			for _, port := range container.IPAddress.Ports {
				portStrings = append(portStrings, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
			}
			rows = append(rows, []string{"Exposed Ports", strings.Join(portStrings, ", ")})
		}
	}

	// Container Information
	if len(container.Containers) > 0 {
		rows = append(rows, []string{"", ""}) // Spacer
		rows = append(rows, []string{"CONTAINERS", ""})

		for i, cont := range container.Containers {
			prefix := fmt.Sprintf("Container %d", i+1)
			rows = append(rows, []string{fmt.Sprintf("%s Name", prefix), cont.Name})
			rows = append(rows, []string{fmt.Sprintf("%s Image", prefix), cont.Image})

			if cont.Resources.Requests != nil {
				rows = append(rows, []string{fmt.Sprintf("%s CPU", prefix), fmt.Sprintf("%.1f cores", cont.Resources.Requests.CPU)})
				rows = append(rows, []string{fmt.Sprintf("%s Memory", prefix), fmt.Sprintf("%.1f GB", cont.Resources.Requests.MemoryInGB)})
			}

			if len(cont.Ports) > 0 {
				portStrings := []string{}
				for _, port := range cont.Ports {
					portStrings = append(portStrings, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
				}
				rows = append(rows, []string{fmt.Sprintf("%s Ports", prefix), strings.Join(portStrings, ", ")})
			}

			if len(cont.EnvironmentVariables) > 0 {
				envStrings := []string{}
				for _, env := range cont.EnvironmentVariables {
					if env.SecureValue != "" {
						envStrings = append(envStrings, fmt.Sprintf("%s=***", env.Name))
					} else {
						envStrings = append(envStrings, fmt.Sprintf("%s=%s", env.Name, env.Value))
					}
				}
				rows = append(rows, []string{fmt.Sprintf("%s Environment", prefix), strings.Join(envStrings, ", ")})
			}

			// Instance View (if available)
			if cont.InstanceView != nil && cont.InstanceView.CurrentState != nil {
				rows = append(rows, []string{fmt.Sprintf("%s State", prefix), cont.InstanceView.CurrentState.State})
				if cont.InstanceView.RestartCount > 0 {
					rows = append(rows, []string{fmt.Sprintf("%s Restarts", prefix), fmt.Sprintf("%d", cont.InstanceView.RestartCount)})
				}
			}
		}
	}

	// Volumes
	if len(container.Volumes) > 0 {
		rows = append(rows, []string{"", ""}) // Spacer
		rows = append(rows, []string{"VOLUMES", ""})

		for _, volume := range container.Volumes {
			rows = append(rows, []string{"Volume Name", volume.Name})
			if volume.AzureFile != nil {
				rows = append(rows, []string{"Type", "Azure File"})
				rows = append(rows, []string{"Share Name", volume.AzureFile.ShareName})
				rows = append(rows, []string{"Storage Account", volume.AzureFile.StorageAccountName})
			} else if volume.EmptyDir != nil {
				rows = append(rows, []string{"Type", "Empty Directory"})
			} else if volume.Secret != nil {
				rows = append(rows, []string{"Type", "Secret"})
			} else if volume.GitRepo != nil {
				rows = append(rows, []string{"Type", "Git Repository"})
				rows = append(rows, []string{"Repository", volume.GitRepo.Repository})
			}
		}
	}

	// Diagnostics
	if container.Diagnostics != nil && container.Diagnostics.LogAnalytics != nil {
		rows = append(rows, []string{"", ""}) // Spacer
		rows = append(rows, []string{"DIAGNOSTICS", ""})
		rows = append(rows, []string{"Log Analytics", "Enabled"})
		rows = append(rows, []string{"Workspace ID", container.Diagnostics.LogAnalytics.WorkspaceID})
		rows = append(rows, []string{"Log Type", container.Diagnostics.LogAnalytics.LogType})
	}

	return tui.RenderMatrixGraph(tui.MatrixGraphMsg{
		Title:  fmt.Sprintf("🐳 Container Instance: %s", name),
		Rows:   rows,
		Labels: []string{"Property", "Value"},
	})
}
