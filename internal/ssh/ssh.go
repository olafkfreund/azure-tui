package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/olafkfreund/azure-tui/internal/azure/vm"
	"golang.org/x/crypto/ssh"
)

// SSHConnection represents an active SSH connection
type SSHConnection struct {
	Client   *ssh.Client
	Host     string
	User     string
	KeyPath  string
	ConnTime time.Time
}

// SSHManager manages SSH connections to Azure VMs
type SSHManager struct {
	connections    map[string]*SSHConnection
	keyPaths       []string
	timeout        time.Duration
	vmManager      *vm.VMManager
	subscriptionID string
}

// NewSSHManager creates a new SSH manager
func NewSSHManager() *SSHManager {
	return &SSHManager{
		connections: make(map[string]*SSHConnection),
		keyPaths:    discoverSSHKeys(),
		timeout:     30 * time.Second,
	}
}

// NewSSHManagerWithVM creates a new SSH manager with VM integration
func NewSSHManagerWithVM(cred *azidentity.DefaultAzureCredential, subscriptionID string) *SSHManager {
	return &SSHManager{
		connections:    make(map[string]*SSHConnection),
		keyPaths:       discoverSSHKeys(),
		timeout:        30 * time.Second,
		vmManager:      vm.NewVMManager(cred, subscriptionID),
		subscriptionID: subscriptionID,
	}
}

// discoverSSHKeys finds SSH keys in common locations
func discoverSSHKeys() []string {
	var keyPaths []string
	
	currentUser, err := user.Current()
	if err != nil {
		return keyPaths
	}
	
	sshDir := filepath.Join(currentUser.HomeDir, ".ssh")
	commonKeys := []string{
		"id_rsa",
		"id_ed25519",
		"id_ecdsa",
		"azure_key",
		"azure_rsa",
	}
	
	for _, keyName := range commonKeys {
		keyPath := filepath.Join(sshDir, keyName)
		if _, err := os.Stat(keyPath); err == nil {
			keyPaths = append(keyPaths, keyPath)
		}
	}
	
	return keyPaths
}

// AddKeyPath adds an SSH key path to the manager
func (sm *SSHManager) AddKeyPath(keyPath string) error {
	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("SSH key not found: %s", keyPath)
	}
	
	for _, existing := range sm.keyPaths {
		if existing == keyPath {
			return nil // Already exists
		}
	}
	
	sm.keyPaths = append(sm.keyPaths, keyPath)
	return nil
}

// Connect establishes an SSH connection to a VM
func (sm *SSHManager) Connect(ctx context.Context, host, username string, keyPath ...string) (*SSHConnection, error) {
	connectionKey := fmt.Sprintf("%s@%s", username, host)
	
	// Check if connection already exists and is still valid
	if existing, exists := sm.connections[connectionKey]; exists {
		if sm.isConnectionAlive(existing) {
			return existing, nil
		}
		// Remove stale connection
		delete(sm.connections, connectionKey)
	}
	
	// Determine which keys to try
	keysToTry := sm.keyPaths
	if len(keyPath) > 0 && keyPath[0] != "" {
		keysToTry = []string{keyPath[0]}
	}
	
	var lastErr error
	for _, kp := range keysToTry {
		client, err := sm.connectWithKey(ctx, host, username, kp)
		if err != nil {
			lastErr = err
			continue
		}
		
		connection := &SSHConnection{
			Client:   client,
			Host:     host,
			User:     username,
			KeyPath:  kp,
			ConnTime: time.Now(),
		}
		
		sm.connections[connectionKey] = connection
		return connection, nil
	}
	
	return nil, fmt.Errorf("failed to connect with any available keys: %v", lastErr)
}

// connectWithKey attempts to connect using a specific SSH key
func (sm *SSHManager) connectWithKey(ctx context.Context, host, username, keyPath string) (*ssh.Client, error) {
	// Read private key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH key %s: %v", keyPath, err)
	}
	
	// Parse private key
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH key %s: %v", keyPath, err)
	}
	
	// Configure SSH client
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, implement proper host key verification
		Timeout:         sm.timeout,
	}
	
	// Add port if not specified
	if !strings.Contains(host, ":") {
		host = host + ": 22"
	}
	
	// Create connection with context
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %v", host, err)
	}
	
	// Perform SSH handshake
	c, chans, reqs, err := ssh.NewClientConn(conn, host, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SSH handshake failed: %v", err)
	}
	
	return ssh.NewClient(c, chans, reqs), nil
}

// ExecuteCommand runs a command on the remote host
func (sm *SSHManager) ExecuteCommand(ctx context.Context, host, username, command string) (string, error) {
	conn, err := sm.Connect(ctx, host, username)
	if err != nil {
		return "", err
	}
	
	session, err := conn.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()
	
	// Set up context cancellation
	done := make(chan error, 1)
	var output []byte
	
	go func() {
		output, err = session.CombinedOutput(command)
		done <- err
	}()
	
	select {
	case err := <-done:
		if err != nil {
			return string(output), fmt.Errorf("command execution failed: %v", err)
		}
		return string(output), nil
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return string(output), ctx.Err()
	}
}

// StartInteractiveSession starts an interactive SSH session
func (sm *SSHManager) StartInteractiveSession(ctx context.Context, host, username string) error {
	// For interactive sessions, we'll use the system SSH client
	// This provides better terminal handling and user experience
	conn, err := sm.Connect(ctx, host, username)
	if err != nil {
		return err
	}
	
	// Use system SSH client for interactive sessions
	sshCmd := exec.CommandContext(ctx, "ssh", 
		"-i", conn.KeyPath,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("%s@%s", username, conn.Host))
	
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	
	return sshCmd.Run()
}

// ConnectViaBas tion connects to a VM through Azure Bastion
func (sm *SSHManager) ConnectViaBastion(ctx context.Context, resourceGroupName, bastionName, vmName, username string) error {
	// Use Azure CLI for Bastion connections
	cmd := exec.CommandContext(ctx, "az", "network", "bastion", "ssh",
		"--resource-group", resourceGroupName,
		"--name", bastionName,
		"--target-resource-id", vmName,
		"--auth-type", "ssh-key",
		"--username", username)
	
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// isConnectionAlive checks if an SSH connection is still active
func (sm *SSHManager) isConnectionAlive(conn *SSHConnection) bool {
	if conn == nil || conn.Client == nil {
		return false
	}
	
	// Try to create a session to test the connection
	session, err := conn.Client.NewSession()
	if err != nil {
		return false
	}
	session.Close()
	
	return true
}

// GetActiveConnections returns all active SSH connections
func (sm *SSHManager) GetActiveConnections() map[string]*SSHConnection {
	activeConns := make(map[string]*SSHConnection)
	
	for key, conn := range sm.connections {
		if sm.isConnectionAlive(conn) {
			activeConns[key] = conn
		} else {
			// Clean up dead connections
			delete(sm.connections, key)
		}
	}
	
	return activeConns
}

// CloseConnection closes a specific SSH connection
func (sm *SSHManager) CloseConnection(host, username string) error {
	connectionKey := fmt.Sprintf("%s@%s", username, host)
	
	if conn, exists := sm.connections[connectionKey]; exists {
		if conn.Client != nil {
			err := conn.Client.Close()
			delete(sm.connections, connectionKey)
			return err
		}
		delete(sm.connections, connectionKey)
	}
	
	return nil
}

// CloseAllConnections closes all active SSH connections
func (sm *SSHManager) CloseAllConnections() error {
	var lastErr error
	
	for key, conn := range sm.connections {
		if conn.Client != nil {
			if err := conn.Client.Close(); err != nil {
				lastErr = err
			}
		}
		delete(sm.connections, key)
	}
	
	return lastErr
}

// GetAvailableKeys returns the list of discovered SSH keys
func (sm *SSHManager) GetAvailableKeys() []string {
	return append([]string{}, sm.keyPaths...)
}

// ConnectToAzureVM connects to an Azure VM using its resource information
func (sm *SSHManager) ConnectToAzureVM(ctx context.Context, resourceGroupName, vmName string, keyPath ...string) (*SSHConnection, error) {
	if sm.vmManager == nil {
		return nil, fmt.Errorf("VM manager not initialized - use NewSSHManagerWithVM()")
	}

	// Get VM connection information
	host, username, err := sm.vmManager.GetVMConnectionInfo(ctx, resourceGroupName, vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM connection info: %v", err)
	}

	// Connect using standard SSH method
	return sm.Connect(ctx, host, username, keyPath...)
}

// StartInteractiveVMSession starts an interactive SSH session with an Azure VM
func (sm *SSHManager) StartInteractiveVMSession(ctx context.Context, resourceGroupName, vmName string) error {
	if sm.vmManager == nil {
		return fmt.Errorf("VM manager not initialized - use NewSSHManagerWithVM()")
	}

	// Get VM connection information
	host, username, err := sm.vmManager.GetVMConnectionInfo(ctx, resourceGroupName, vmName)
	if err != nil {
		return fmt.Errorf("failed to get VM connection info: %v", err)
	}

	// Start interactive session
	return sm.StartInteractiveSession(ctx, host, username)
}

// ExecuteVMCommand executes a command on an Azure VM
func (sm *SSHManager) ExecuteVMCommand(ctx context.Context, resourceGroupName, vmName, command string) (string, error) {
	if sm.vmManager == nil {
		return "", fmt.Errorf("VM manager not initialized - use NewSSHManagerWithVM()")
	}

	// Get VM connection information
	host, username, err := sm.vmManager.GetVMConnectionInfo(ctx, resourceGroupName, vmName)
	if err != nil {
		return "", fmt.Errorf("failed to get VM connection info: %v", err)
	}

	// Execute command
	return sm.ExecuteCommand(ctx, host, username, command)
}

// GetVMInfo retrieves VM information including connection details
func (sm *SSHManager) GetVMInfo(ctx context.Context, resourceGroupName, vmName string) (*vm.VMInfo, error) {
	if sm.vmManager == nil {
		return nil, fmt.Errorf("VM manager not initialized - use NewSSHManagerWithVM()")
	}

	return sm.vmManager.GetVMInfo(ctx, resourceGroupName, vmName)
}

// SetTimeout sets the SSH connection timeout
func (sm *SSHManager) SetTimeout(timeout time.Duration) {
	sm.timeout = timeout
}
