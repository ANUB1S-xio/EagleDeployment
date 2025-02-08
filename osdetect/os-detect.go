// File: os-detect.go
// Directory: EagleDeploy_CLI/osdetect
// Purpose: Detects OS type of hosts using SSH

package osdetect

import (
	"EagleDeploy_CLI/sshutils"
	"fmt"
)

// DetectOS connects to a host via SSH and determines its OS.
func DetectOS(host, user, password string, port int) (string, error) {
	// Connect to the remote host via SSH.
	client, err := sshutils.ConnectSSH(host, user, password, port)
	if err != nil {
		return "", fmt.Errorf("failed to connect to host %s: %v", host, err)
	}
	defer sshutils.CloseSSHConnection(client)

	// Try detecting OS using "uname -s" (common in Unix/Linux).
	output, err := sshutils.RunSSHCommand(client, "uname -s")
	if err == nil {
		return output, nil
	}

	// Fall back to "ver" for Windows.
	output, err = sshutils.RunSSHCommand(client, "ver")
	if err == nil {
		return output, nil
	}

	return "Unknown", fmt.Errorf("unable to determine OS for host %s", host)
}

// NOTE: The UpdateInventory function that depended on the inventory package has been removed
// to prevent import cycles. Inventory updates using OS detection should be handled in the inventory package.
