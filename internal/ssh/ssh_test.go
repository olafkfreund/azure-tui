package ssh

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSSHManager(t *testing.T) {
	manager := NewSSHManager()

	if manager == nil {
		t.Fatal("NewSSHManager returned nil")
	}

	if manager.connections == nil {
		t.Error("SSH manager connections map is nil")
	}

	if manager.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", manager.timeout)
	}
}

func TestDiscoverSSHKeys(t *testing.T) {
	keys := discoverSSHKeys()

	// Test should pass even if no keys are found
	if keys == nil {
		t.Error("discoverSSHKeys returned nil")
	}

	// If we're in a test environment, we might not have SSH keys
	t.Logf("Found %d SSH keys", len(keys))
	for i, key := range keys {
		t.Logf("Key %d: %s", i+1, key)
	}
}

func TestAddKeyPath(t *testing.T) {
	manager := NewSSHManager()

	// Test adding a non-existent key path
	err := manager.AddKeyPath("/nonexistent/key")
	if err == nil {
		t.Error("Expected error for non-existent key path")
	}

	// Test adding a valid path (create temporary file)
	tmpDir := t.TempDir()
	tmpKeyPath := filepath.Join(tmpDir, "test_key")

	// Create a temporary file
	file, err := os.Create(tmpKeyPath)
	if err != nil {
		t.Fatalf("Failed to create temporary key file: %v", err)
	}
	file.Close()

	// Test adding the valid key path
	err = manager.AddKeyPath(tmpKeyPath)
	if err != nil {
		t.Errorf("Failed to add valid key path: %v", err)
	}

	// Verify the key was added
	found := false
	for _, key := range manager.keyPaths {
		if key == tmpKeyPath {
			found = true
			break
		}
	}
	if !found {
		t.Error("Key path was not added to manager")
	}

	// Test adding the same key path again (should not duplicate)
	initialCount := len(manager.keyPaths)
	err = manager.AddKeyPath(tmpKeyPath)
	if err != nil {
		t.Errorf("Failed to add key path again: %v", err)
	}
	if len(manager.keyPaths) != initialCount {
		t.Error("Key path was duplicated")
	}
}

func TestGetAvailableKeys(t *testing.T) {
	manager := NewSSHManager()

	keys := manager.GetAvailableKeys()
	if keys == nil {
		t.Error("GetAvailableKeys returned nil")
	}

	// Test that we get a copy, not the original slice
	if len(keys) > 0 {
		original := keys[0]
		keys[0] = "modified"
		managerKeys := manager.GetAvailableKeys()
		if len(managerKeys) > 0 && managerKeys[0] == "modified" {
			t.Error("GetAvailableKeys returned original slice instead of copy")
		}
		// Restore for other tests
		keys[0] = original
	}
}

func TestSetTimeout(t *testing.T) {
	manager := NewSSHManager()

	newTimeout := 60 * time.Second
	manager.SetTimeout(newTimeout)

	if manager.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, manager.timeout)
	}
}

func TestSSHConnection(t *testing.T) {
	// Test SSHConnection struct
	conn := &SSHConnection{
		Host:     "test.example.com",
		User:     "testuser",
		KeyPath:  "/test/key",
		ConnTime: time.Now(),
	}

	if conn.Host != "test.example.com" {
		t.Errorf("Expected host 'test.example.com', got '%s'", conn.Host)
	}

	if conn.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", conn.User)
	}
}

func TestGetActiveConnections(t *testing.T) {
	manager := NewSSHManager()

	// Test with no connections
	active := manager.GetActiveConnections()
	if active == nil {
		t.Error("GetActiveConnections returned nil")
	}
	if len(active) != 0 {
		t.Errorf("Expected 0 active connections, got %d", len(active))
	}
}

func TestCloseConnection(t *testing.T) {
	manager := NewSSHManager()

	// Test closing non-existent connection
	err := manager.CloseConnection("nonexistent.com", "user")
	if err != nil {
		t.Errorf("CloseConnection returned error for non-existent connection: %v", err)
	}
}

func TestCloseAllConnections(t *testing.T) {
	manager := NewSSHManager()

	// Test with no connections
	err := manager.CloseAllConnections()
	if err != nil {
		t.Errorf("CloseAllConnections returned error with no connections: %v", err)
	}
}

// TestSSHKeyDiscoveryPaths tests that SSH key discovery looks in the right places
func TestSSHKeyDiscoveryPaths(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Skip("Cannot get current user, skipping SSH key discovery test")
	}

	sshDir := filepath.Join(currentUser.HomeDir, ".ssh")
	expectedKeys := []string{
		filepath.Join(sshDir, "id_rsa"),
		filepath.Join(sshDir, "id_ed25519"),
		filepath.Join(sshDir, "id_ecdsa"),
		filepath.Join(sshDir, "azure_key"),
		filepath.Join(sshDir, "azure_rsa"),
	}

	discoveredKeys := discoverSSHKeys()

	// Check that discovered keys are from expected paths
	for _, discovered := range discoveredKeys {
		found := false
		for _, expected := range expectedKeys {
			if discovered == expected {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Discovered unexpected key path: %s", discovered)
		}
	}

	// Log what we found for debugging
	t.Logf("SSH directory: %s", sshDir)
	t.Logf("Expected key paths: %v", expectedKeys)
	t.Logf("Discovered keys: %v", discoveredKeys)
}

// Benchmark tests
func BenchmarkNewSSHManager(b *testing.B) {
	for i := 0; i < b.N; i++ {
		manager := NewSSHManager()
		_ = manager
	}
}

func BenchmarkDiscoverSSHKeys(b *testing.B) {
	for i := 0; i < b.N; i++ {
		keys := discoverSSHKeys()
		_ = keys
	}
}
