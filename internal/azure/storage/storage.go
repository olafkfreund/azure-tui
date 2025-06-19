package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Container represents a blob container in a storage account
type Container struct {
	Name                  string                 `json:"name"`
	LastModified          string                 `json:"lastModified"`
	Etag                  string                 `json:"etag"`
	Lease                 map[string]interface{} `json:"lease,omitempty"`
	PublicAccess          string                 `json:"publicAccess,omitempty"`
	HasImmutabilityPolicy bool                   `json:"hasImmutabilityPolicy,omitempty"`
	HasLegalHold          bool                   `json:"hasLegalHold,omitempty"`
	Metadata              map[string]string      `json:"metadata,omitempty"`
}

// Blob represents a blob in a container
type Blob struct {
	Name            string            `json:"name"`
	Container       string            `json:"container"`
	LastModified    string            `json:"lastModified"`
	Etag            string            `json:"etag"`
	Size            int64             `json:"contentLength"`
	ContentType     string            `json:"contentType"`
	ContentEncoding string            `json:"contentEncoding,omitempty"`
	BlobType        string            `json:"blobType"`
	AccessTier      string            `json:"accessTier,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
}

// StorageAccount represents an Azure Storage Account
type StorageAccount struct {
	Name                   string `json:"name"`
	Location               string `json:"location"`
	ResourceGroup          string `json:"resourceGroup"`
	Kind                   string `json:"kind,omitempty"`
	SkuName                string `json:"skuName,omitempty"`
	ProvisioningState      string `json:"provisioningState,omitempty"`
	CreationTime           string `json:"creationTime,omitempty"`
	PrimaryLocation        string `json:"primaryLocation,omitempty"`
	StatusOfPrimary        string `json:"statusOfPrimary,omitempty"`
	EnableHttpsTrafficOnly bool   `json:"enableHttpsTrafficOnly,omitempty"`
	AllowBlobPublicAccess  bool   `json:"allowBlobPublicAccess,omitempty"`
}

// ListContainers lists all blob containers in a storage account
func ListContainers(accountName string) ([]Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "container", "list",
		"--account-name", accountName,
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for account %s: %v", accountName, err)
	}

	var containers []Container
	if err := json.Unmarshal(output, &containers); err != nil {
		return nil, fmt.Errorf("failed to parse container data: %v", err)
	}

	return containers, nil
}

// ListBlobs lists all blobs in a specific container
func ListBlobs(accountName, containerName string) ([]Blob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "blob", "list",
		"--account-name", accountName,
		"--container-name", containerName,
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list blobs in container %s: %v", containerName, err)
	}

	var blobs []Blob
	if err := json.Unmarshal(output, &blobs); err != nil {
		return nil, fmt.Errorf("failed to parse blob data: %v", err)
	}

	// Set container name for each blob
	for i := range blobs {
		blobs[i].Container = containerName
	}

	return blobs, nil
}

// CreateContainer creates a new blob container
func CreateContainer(accountName, containerName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "container", "create",
		"--account-name", accountName,
		"--name", containerName)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create container %s: %v", containerName, err)
	}

	return nil
}

// DeleteContainer deletes a blob container
func DeleteContainer(accountName, containerName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "container", "delete",
		"--account-name", accountName,
		"--name", containerName)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete container %s: %v", containerName, err)
	}

	return nil
}

// UploadBlob uploads a file to a blob container
func UploadBlob(accountName, containerName, blobName, filePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "blob", "upload",
		"--account-name", accountName,
		"--container-name", containerName,
		"--name", blobName,
		"--file", filePath,
		"--overwrite")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upload blob %s: %v", blobName, err)
	}

	return nil
}

// DeleteBlob deletes a blob from a container
func DeleteBlob(accountName, containerName, blobName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "blob", "delete",
		"--account-name", accountName,
		"--container-name", containerName,
		"--name", blobName)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete blob %s: %v", blobName, err)
	}

	return nil
}

// GetBlobProperties gets detailed properties of a specific blob
func GetBlobProperties(accountName, containerName, blobName string) (*Blob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "blob", "show",
		"--account-name", accountName,
		"--container-name", containerName,
		"--name", blobName,
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get blob properties for %s: %v", blobName, err)
	}

	var blob Blob
	if err := json.Unmarshal(output, &blob); err != nil {
		return nil, fmt.Errorf("failed to parse blob properties: %v", err)
	}

	blob.Container = containerName
	return &blob, nil
}

// RenderStorageContainersView renders the containers list view for TUI
func RenderStorageContainersView(accountName string, containers []Container) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("🗄️  Storage Containers in '%s'\n", accountName))
	content.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	if len(containers) == 0 {
		content.WriteString("📭 No containers found in this storage account.\n\n")
		content.WriteString("Available Actions:\n")
		content.WriteString("• Press 'Shift+S' to create a new container\n")
		return content.String()
	}

	content.WriteString("📋 Container Inventory:\n")
	for _, container := range containers {
		status := "🟢 Available"
		if container.Lease != nil {
			if state, ok := container.Lease["state"].(string); ok && state == "leased" {
				status = "🔒 Leased"
			}
		}

		content.WriteString(fmt.Sprintf("• %s (%s)\n", container.Name, status))
		if container.LastModified != "" {
			content.WriteString(fmt.Sprintf("  Last Modified: %s\n", container.LastModified))
		}
		if container.PublicAccess != "" && container.PublicAccess != "off" {
			content.WriteString(fmt.Sprintf("  Public Access: %s\n", container.PublicAccess))
		}
		if len(container.Metadata) > 0 {
			content.WriteString("  Metadata: ")
			var metaPairs []string
			for key, value := range container.Metadata {
				metaPairs = append(metaPairs, fmt.Sprintf("%s=%s", key, value))
			}
			content.WriteString(strings.Join(metaPairs, ", "))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString("Available Actions:\n")
	content.WriteString("• Press 'B' to list blobs in a container\n")
	content.WriteString("• Press 'Shift+S' to create a new container\n")
	content.WriteString("• Press 'Ctrl+X' to delete a container\n")

	return content.String()
}

// RenderStorageBlobsView renders the blobs list view for TUI
func RenderStorageBlobsView(accountName, containerName string, blobs []Blob) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("📁 Blobs in Container '%s' (Account: %s)\n", containerName, accountName))
	content.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	if len(blobs) == 0 {
		content.WriteString("📭 No blobs found in this container.\n\n")
		content.WriteString("Available Actions:\n")
		content.WriteString("• Press 'U' to upload a blob\n")
		content.WriteString("• Press 'Esc' to go back to containers\n")
		return content.String()
	}

	content.WriteString("📋 Blob Inventory:\n")
	for _, blob := range blobs {
		// Format size
		var sizeStr string
		if blob.Size < 1024 {
			sizeStr = fmt.Sprintf("%d B", blob.Size)
		} else if blob.Size < 1024*1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(blob.Size)/1024)
		} else if blob.Size < 1024*1024*1024 {
			sizeStr = fmt.Sprintf("%.1f MB", float64(blob.Size)/(1024*1024))
		} else {
			sizeStr = fmt.Sprintf("%.1f GB", float64(blob.Size)/(1024*1024*1024))
		}

		// Blob type icon
		typeIcon := "📄"
		switch blob.BlobType {
		case "BlockBlob":
			typeIcon = "🧱"
		case "PageBlob":
			typeIcon = "📄"
		case "AppendBlob":
			typeIcon = "📝"
		}

		content.WriteString(fmt.Sprintf("%s %s (%s)\n", typeIcon, blob.Name, sizeStr))
		if blob.ContentType != "" {
			content.WriteString(fmt.Sprintf("   Type: %s\n", blob.ContentType))
		}
		if blob.LastModified != "" {
			content.WriteString(fmt.Sprintf("   Modified: %s\n", blob.LastModified))
		}
		if blob.AccessTier != "" {
			content.WriteString(fmt.Sprintf("   Access Tier: %s\n", blob.AccessTier))
		}
		if len(blob.Tags) > 0 {
			content.WriteString("   Tags: ")
			var tagPairs []string
			for key, value := range blob.Tags {
				tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", key, value))
			}
			content.WriteString(strings.Join(tagPairs, ", "))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString("Available Actions:\n")
	content.WriteString("• Press 'U' to upload a new blob\n")
	content.WriteString("• Press 'Ctrl+X' to delete a blob\n")
	content.WriteString("• Press 'Esc' to go back to containers\n")

	return content.String()
}

// RenderBlobDetails renders detailed view of a specific blob
func RenderBlobDetails(blob *Blob) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("📄 Blob Details: %s\n", blob.Name))
	content.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Format size
	var sizeStr string
	if blob.Size < 1024 {
		sizeStr = fmt.Sprintf("%d B", blob.Size)
	} else if blob.Size < 1024*1024 {
		sizeStr = fmt.Sprintf("%.1f KB", float64(blob.Size)/1024)
	} else if blob.Size < 1024*1024*1024 {
		sizeStr = fmt.Sprintf("%.1f MB", float64(blob.Size)/(1024*1024))
	} else {
		sizeStr = fmt.Sprintf("%.1f GB", float64(blob.Size)/(1024*1024*1024))
	}

	content.WriteString(fmt.Sprintf("Name: %s\n", blob.Name))
	content.WriteString(fmt.Sprintf("Container: %s\n", blob.Container))
	content.WriteString(fmt.Sprintf("Size: %s\n", sizeStr))
	content.WriteString(fmt.Sprintf("Type: %s\n", blob.BlobType))
	if blob.ContentType != "" {
		content.WriteString(fmt.Sprintf("Content Type: %s\n", blob.ContentType))
	}
	if blob.ContentEncoding != "" {
		content.WriteString(fmt.Sprintf("Content Encoding: %s\n", blob.ContentEncoding))
	}

	content.WriteString("\n📅 Timestamps:\n")
	if blob.LastModified != "" {
		content.WriteString(fmt.Sprintf("Last Modified: %s\n", blob.LastModified))
	}
	if blob.Etag != "" {
		content.WriteString(fmt.Sprintf("ETag: %s\n", blob.Etag))
	}

	if blob.AccessTier != "" {
		content.WriteString(fmt.Sprintf("\n🏷️  Access Tier: %s\n", blob.AccessTier))
	}

	if len(blob.Tags) > 0 {
		content.WriteString("\n🏷️  Tags:\n")
		for key, value := range blob.Tags {
			content.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	if len(blob.Metadata) > 0 {
		content.WriteString("\n📋 Metadata:\n")
		for key, value := range blob.Metadata {
			content.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	content.WriteString("\n💡 Tip: Use Azure Storage Explorer or az CLI for downloading blobs\n")

	return content.String()
}

// ListStorageAccounts lists all storage accounts (existing function enhanced)
func ListStorageAccounts() ([]StorageAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list storage accounts: %v", err)
	}

	var accounts []StorageAccount
	if err := json.Unmarshal(output, &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse storage account data: %v", err)
	}

	return accounts, nil
}

// CreateStorageAccount creates a new storage account (existing function)
func CreateStorageAccount(name, group, location string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "account", "create",
		"--name", name,
		"--resource-group", group,
		"--location", location,
		"--sku", "Standard_LRS")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create storage account %s: %v", name, err)
	}

	return nil
}

// DeleteStorageAccount deletes a storage account (existing function)
func DeleteStorageAccount(name, group string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "storage", "account", "delete",
		"--name", name,
		"--resource-group", group,
		"--yes")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete storage account %s: %v", name, err)
	}

	return nil
}
