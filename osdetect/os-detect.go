// File: os-detect.go
// Directory: EagleDeploy_CLI/osdetect
// Purpose: Detects OS type of hosts using SSH and TCP fingerprinting

package osdetect

import (
	"EagleDeploy_CLI/sshutils"
	"fmt"
	"net"
	"strings"
	"time"
)

// Function: DetectOS
// Purpose: Primary OS detection function using multiple methods
// Parameters:
//   - host: string - Target host address
//   - user: string - SSH username
//   - password: string - SSH password
//   - port: int - SSH port number
//
// Returns:
//   - string - Detected OS type
//   - error - Any detection errors
//
// Called By:
//   - [`inventory.AddHost`](../inventory/inventory.go)
//
// Dependencies:
//   - sshutils.ConnectSSH
//   - detectOSFromTCP
func DetectOS(host, user, password string, port int) (string, error) {
	// First try TCP fingerprinting
	osType, err := detectOSFromTCP(host)
	if err == nil && osType != "Unknown" {
		return osType, nil
	}

	// Fall back to SSH detection if TCP fingerprinting fails
	client, err := sshutils.ConnectSSH(host, user, password, port)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %v", err)
	}
	defer sshutils.CloseSSHConnection(client)

	// Enhanced detection methods in order of reliability
	detectionMethods := []struct {
		name    string
		command string
		parser  func(string) string
	}{
		// Windows detection
		{
			name:    "Windows PowerShell",
			command: "powershell.exe -Command \"(Get-CimInstance Win32_OperatingSystem).Caption\"",
			parser:  parseWindowsOutput,
		},
		// Linux detection methods
		{
			name:    "SystemD OS-Release",
			command: "cat /etc/os-release",
			parser:  parseOSRelease,
		},
		{
			name:    "LSB Release",
			command: "lsb_release -a",
			parser:  parseLSBRelease,
		},
		{
			name:    "Fedora/RHEL Release",
			command: "cat /etc/redhat-release",
			parser:  parseRedHatRelease,
		},
		{
			name:    "System Information",
			command: "hostnamectl",
			parser:  parseHostnamectl,
		},
		{
			name:    "Uname Full",
			command: "uname -a",
			parser:  parseUname,
		},
	}

	for _, method := range detectionMethods {
		output, err := sshutils.RunSSHCommand(client, method.command)
		if err == nil {
			if osType := method.parser(output); osType != "" {
				return osType, nil
			}
		}
	}

	return "Unknown", fmt.Errorf("failed to detect OS using all methods")
}

// Function: detectOSFromTCP
// Purpose: OS detection using TCP fingerprinting
// Parameters:
//   - host: string - Target host address
//
// Returns:
//   - string - Detected OS type
//   - error - Any detection errors
//
// Called By:
//   - DetectOS as first detection attempt
//
// Dependencies:
//   - net.DialTimeout
//   - TCP ports: 22, 445, 139, 135
func detectOSFromTCP(host string) (string, error) {
	// Try connecting to common ports
	ports := []int{22, 445, 139, 135} // SSH, SMB, NetBIOS
	timeout := time.Second * 2

	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), timeout)
		if err != nil {
			continue
		}
		defer conn.Close()

		// Get TCP connection details
		_, ok := conn.(*net.TCPConn)
		if !ok {
			continue
		}

		// Check TCP window size and other characteristics
		if port == 445 || port == 139 || port == 135 {
			return "Windows", nil // Strong indication of Windows
		}

		// Analyze SSH banner if port 22
		if port == 22 {
			buffer := make([]byte, 64)
			conn.SetReadDeadline(time.Now().Add(timeout))
			n, _ := conn.Read(buffer)
			banner := string(buffer[:n])

			switch {
			case strings.Contains(strings.ToLower(banner), "ubuntu"):
				return "Linux - Ubuntu", nil
			case strings.Contains(strings.ToLower(banner), "fedora"):
				return "Linux - Fedora", nil
			case strings.Contains(strings.ToLower(banner), "openssh"):
				if strings.Contains(strings.ToLower(banner), "windows") {
					return "Windows", nil
				}
				return "Linux", nil
			}
		}
	}

	return "Unknown", fmt.Errorf("couldn't determine OS from TCP fingerprinting")
}

// Function: parseHostnamectl
// Purpose: Parses output from hostnamectl command
// Parameters:
//   - output: string - Command output to parse
//
// Returns:
//   - string - Formatted OS string or empty if not found
//
// Called By:
//   - DetectOS via detection methods
func parseHostnamectl(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Operating System:") {
			os := strings.TrimSpace(strings.TrimPrefix(line, "Operating System:"))
			return fmt.Sprintf("Linux - %s", os)
		}
	}
	return ""
}

// Function: parseWindowsOutput
// Purpose: Parses Windows system information output
// Parameters:
//   - output: string - PowerShell command output
//
// Returns:
//   - string - Formatted Windows version string
//
// Called By:
//   - DetectOS via detection methods
func parseWindowsOutput(output string) string {
	output = strings.TrimSpace(output)
	if strings.Contains(output, "Microsoft") {
		return strings.Replace(output, "Microsoft ", "", 1)
	}
	return output
}

// Function: parseOSRelease
// Purpose: Parses /etc/os-release file content
// Parameters:
//   - output: string - File content
//
// Returns:
//   - string - Formatted Linux distribution string
//
// Called By:
//   - DetectOS via detection methods
func parseOSRelease(output string) string {
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
			return fmt.Sprintf("Linux - %s %s", name, version)
		}
		return fmt.Sprintf("Linux - %s", name)
	}
	return ""
}

// Function: parseLSBRelease
// Purpose: Parses lsb_release command output
// Parameters:
//   - output: string - Command output
//
// Returns:
//   - string - Formatted Linux distribution string
//
// Called By:
//   - DetectOS via detection methods
func parseLSBRelease(output string) string {
	lines := strings.Split(output, "\n")
	var name, version string
	for _, line := range lines {
		if strings.HasPrefix(line, "DISTRIB_ID=") {
			name = strings.Trim(strings.TrimPrefix(line, "DISTRIB_ID="), "\"")
		}
		if strings.HasPrefix(line, "DISTRIB_RELEASE=") {
			version = strings.Trim(strings.TrimPrefix(line, "DISTRIB_RELEASE="), "\"")
		}
	}
	if name != "" {
		if version != "" {
			return fmt.Sprintf("Linux - %s %s", name, version)
		}
		return fmt.Sprintf("Linux - %s", name)
	}
	return ""
}

// Function: parseRedHatRelease
// Purpose: Parses /etc/redhat-release file content
// Parameters:
//   - output: string - File content
//
// Returns:
//   - string - Formatted Red Hat/Fedora string
//
// Called By:
//   - DetectOS via detection methods
func parseRedHatRelease(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "release") {
			return fmt.Sprintf("Linux - %s", strings.TrimSpace(line))
		}
	}
	return ""
}

// Function: parseUname
// Purpose: Parses uname -a command output
// Parameters:
//   - output: string - Command output
//
// Returns:
//   - string - Basic OS type string
//
// Called By:
//   - DetectOS via detection methods
//
// Notes:
//   - Currently unimplemented (placeholder)
func parseUname(output string) string {
	panic("unimplemented")
}
