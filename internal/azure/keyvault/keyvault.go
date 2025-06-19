package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type KeyVault struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

type Secret struct {
	Name        string            `json:"name"`
	ID          string            `json:"id"`
	Enabled     bool              `json:"enabled"`
	Created     string            `json:"created"`
	Updated     string            `json:"updated"`
	ContentType string            `json:"contentType"`
	Tags        map[string]string `json:"tags"`
	Attributes  SecretAttributes  `json:"attributes"`
}

type SecretAttributes struct {
	Enabled   bool   `json:"enabled"`
	Created   int64  `json:"created"`
	Updated   int64  `json:"updated"`
	NotBefore *int64 `json:"notBefore"`
	Expires   *int64 `json:"expires"`
}

func ListKeyVaults() ([]KeyVault, error) {
	cmd := exec.Command("az", "keyvault", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var vaults []KeyVault
	if err := json.Unmarshal(out, &vaults); err != nil {
		return nil, err
	}
	return vaults, nil
}

func CreateKeyVault(name, group, location string) error {
	return exec.Command("az", "keyvault", "create", "--name", name, "--resource-group", group, "--location", location).Run()
}

func DeleteKeyVault(name, group string) error {
	return exec.Command("az", "keyvault", "delete", "--name", name, "--resource-group", group).Run()
}

// =============================================================================
// SECRET MANAGEMENT FUNCTIONS
// =============================================================================

// ListSecrets lists all secrets in a Key Vault
func ListSecrets(vaultName string) ([]Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "keyvault", "secret", "list",
		"--vault-name", vaultName, "--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %v", err)
	}

	var secrets []Secret
	if err := json.Unmarshal(output, &secrets); err != nil {
		return nil, fmt.Errorf("failed to parse secrets: %v", err)
	}

	return secrets, nil
}

// CreateSecret creates a new secret in a Key Vault
func CreateSecret(vaultName, secretName, secretValue string, tags map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := []string{"keyvault", "secret", "set",
		"--vault-name", vaultName,
		"--name", secretName,
		"--value", secretValue}

	// Add tags if provided
	if len(tags) > 0 {
		var tagStrings []string
		for key, value := range tags {
			tagStrings = append(tagStrings, fmt.Sprintf("%s=%s", key, value))
		}
		args = append(args, "--tags", strings.Join(tagStrings, " "))
	}

	cmd := exec.CommandContext(ctx, "az", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}

	return nil
}

// DeleteSecret deletes a secret from a Key Vault
func DeleteSecret(vaultName, secretName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "keyvault", "secret", "delete",
		"--vault-name", vaultName,
		"--name", secretName)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete secret: %v", err)
	}

	return nil
}

// GetSecretMetadata gets detailed metadata about a secret (without the value)
func GetSecretMetadata(vaultName, secretName string) (*Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "keyvault", "secret", "show",
		"--vault-name", vaultName,
		"--name", secretName,
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get secret metadata: %v", err)
	}

	var secret Secret
	if err := json.Unmarshal(output, &secret); err != nil {
		return nil, fmt.Errorf("failed to parse secret metadata: %v", err)
	}

	return &secret, nil
}

// RenderKeyVaultSecretsView renders a formatted view of Key Vault secrets
func RenderKeyVaultSecretsView(vaultName string, secrets []Secret) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("🔐 Key Vault Secrets: %s\n", vaultName))
	content.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	if len(secrets) == 0 {
		content.WriteString("No secrets found in this Key Vault.\n\n")
		content.WriteString("💡 Press 'C' to create a new secret\n")
		return content.String()
	}

	content.WriteString(fmt.Sprintf("Found %d secret(s):\n\n", len(secrets)))

	for i, secret := range secrets {
		status := "🟢 Enabled"
		if !secret.Enabled {
			status = "🔴 Disabled"
		}

		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, secret.Name))
		content.WriteString(fmt.Sprintf("   Status: %s\n", status))
		content.WriteString(fmt.Sprintf("   ID: %s\n", secret.ID))

		if secret.ContentType != "" {
			content.WriteString(fmt.Sprintf("   Content Type: %s\n", secret.ContentType))
		}

		if secret.Created != "" {
			content.WriteString(fmt.Sprintf("   Created: %s\n", secret.Created))
		}

		if secret.Updated != "" {
			content.WriteString(fmt.Sprintf("   Updated: %s\n", secret.Updated))
		}

		if len(secret.Tags) > 0 {
			content.WriteString("   Tags: ")
			var tagPairs []string
			for key, value := range secret.Tags {
				tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", key, value))
			}
			content.WriteString(strings.Join(tagPairs, ", "))
			content.WriteString("\n")
		}

		content.WriteString("\n")
	}

	content.WriteString("Available Actions:\n")
	content.WriteString("• Press 'C' to create a new secret\n")
	content.WriteString("• Press 'D' to delete a selected secret\n")
	content.WriteString("• Press 'R' to refresh the list\n")
	content.WriteString("• Press 'Enter' to view secret details\n")

	return content.String()
}

// RenderSecretDetails renders detailed information about a specific secret
func RenderSecretDetails(secret *Secret) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("🔐 Secret Details: %s\n", secret.Name))
	content.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	status := "🟢 Enabled"
	if !secret.Enabled {
		status = "🔴 Disabled"
	}

	content.WriteString(fmt.Sprintf("Name: %s\n", secret.Name))
	content.WriteString(fmt.Sprintf("Status: %s\n", status))
	content.WriteString(fmt.Sprintf("ID: %s\n", secret.ID))

	if secret.ContentType != "" {
		content.WriteString(fmt.Sprintf("Content Type: %s\n", secret.ContentType))
	}

	content.WriteString("\n📅 Timestamps:\n")
	if secret.Created != "" {
		content.WriteString(fmt.Sprintf("Created: %s\n", secret.Created))
	}
	if secret.Updated != "" {
		content.WriteString(fmt.Sprintf("Updated: %s\n", secret.Updated))
	}

	if secret.Attributes.NotBefore != nil {
		content.WriteString(fmt.Sprintf("Not Before: %s\n",
			time.Unix(*secret.Attributes.NotBefore, 0).Format(time.RFC3339)))
	}
	if secret.Attributes.Expires != nil {
		content.WriteString(fmt.Sprintf("Expires: %s\n",
			time.Unix(*secret.Attributes.Expires, 0).Format(time.RFC3339)))
	}

	if len(secret.Tags) > 0 {
		content.WriteString("\n🏷️ Tags:\n")
		for key, value := range secret.Tags {
			content.WriteString(fmt.Sprintf("• %s: %s\n", key, value))
		}
	}

	content.WriteString("\n🔒 Security Note:\n")
	content.WriteString("Secret values are not displayed for security reasons.\n")
	content.WriteString("Use Azure CLI or Azure Portal to view secret values if needed.\n")

	return content.String()
}
