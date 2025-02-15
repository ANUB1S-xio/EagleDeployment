// File: os-detect.go
// Directory: EagleDeploy_CLI/osdetect
// Purpose: Detects OS type of hosts using SSH

package osdetect

import (
	"EagleDeploy_CLI/sshutils"
	"fmt"
	"strings"
)

// DetectOS connects to a host via SSH and determines its OS.
func DetectOS(host, user, password string, port int) (string, error) {
	client, err := sshutils.ConnectSSH(host, user, password, port)
	if err != nil {
		return "", fmt.Errorf("failed to connect to host %s: %v", host, err)
	}
	defer sshutils.CloseSSHConnection(client)

	// Try Windows detection first using PowerShell
	output, err := sshutils.RunSSHCommand(client, "powershell.exe -Command \"(Get-CimInstance Win32_OperatingSystem).Caption\"")
	if err == nil && strings.Contains(strings.ToLower(output), "windows") {
		// Clean up the output
		output = strings.TrimSpace(output)
		output = strings.Replace(output, "Microsoft ", "", 1)
		return output, nil
	}

	// Try Linux detection using os-release
	output, err = sshutils.RunSSHCommand(client, "cat /etc/os-release")
	if err == nil {
		// Parse os-release file
		lines := strings.Split(output, "\n")
		var name, version string
		for _, line := range lines {
			if strings.HasPrefix(line, "NAME=") {
				name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
			}
			if strings.HasPrefix(line, "VERSION_ID=") {
				version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
			}
		}
		if name != "" {
			if version != "" {
				return fmt.Sprintf("Linux - %s %s", name, version), nil
			}
			return fmt.Sprintf("Linux - %s", name), nil
		}
	}

	// Fallback to basic Linux detection
	output, err = sshutils.RunSSHCommand(client, "uname -a")
	if err == nil {
		if strings.Contains(strings.ToLower(output), "ubuntu") {
			return "Linux - Ubuntu", nil
		} else if strings.Contains(strings.ToLower(output), "fedora") {
			return "Linux - Fedora", nil
		} else if strings.Contains(strings.ToLower(output), "centos") {
			return "Linux - CentOS", nil
		} else if strings.Contains(strings.ToLower(output), "debian") {
			return "Linux - Debian", nil
		}
		return "Linux - Generic", nil
	}

	return "Unknown", fmt.Errorf("unable to determine OS for host %s", host)
}
