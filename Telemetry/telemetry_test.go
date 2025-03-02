package Telemetry

import (
	"testing"
	"time"
)

func TestGetInstance(t *testing.T) {
	instance1 := GetInstance()
	instance2 := GetInstance()

	if instance1 != instance2 {
		t.Errorf("Expected singleton instance, got different instances")
	}
}

func TestSetLevel(t *testing.T) {
	telemetry := GetInstance()
	telemetry.SetLevel(LevelDebug)

	if telemetry.level != LevelDebug {
		t.Errorf("Expected level to be LevelDebug, got %d", telemetry.level)
	}
}

func TestEnableConsole(t *testing.T) {
	telemetry := GetInstance()
	telemetry.EnableConsole(false)

	if telemetry.enableConsole != false {
		t.Errorf("Expected enableConsole to be false, got %v", telemetry.enableConsole)
	}
}

func TestLogError(t *testing.T) {
	telemetry := GetInstance()
	telemetry.SetLevel(LevelError)
	telemetry.LogError("TestCategory", "Test error message", nil)

	// Check if the log file contains the error message
	// Implementation details would go here
}

func TestLogWarning(t *testing.T) {
	telemetry := GetInstance()
	telemetry.SetLevel(LevelWarning)
	telemetry.LogWarning("TestCategory", "Test warning message", nil)

	// Check if the log file contains the warning message
	// Implementation details would go here
}

func TestLogInfo(t *testing.T) {
	telemetry := GetInstance()
	telemetry.SetLevel(LevelInfo)
	telemetry.LogInfo("TestCategory", "Test info message", nil)

	// Check if the log file contains the info message
	// Implementation details would go here
}

func TestLogDebug(t *testing.T) {
	telemetry := GetInstance()
	telemetry.SetLevel(LevelDebug)
	telemetry.LogDebug("TestCategory", "Test debug message", nil)

	// Check if the log file contains the debug message
	// Implementation details would go here
}

func TestClearLogs(t *testing.T) {
	telemetry := GetInstance()
	err := telemetry.ClearLogs()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check if the logs are cleared
	// Implementation details would go here
}

func TestFilterLogs(t *testing.T) {
	telemetry := GetInstance()
	telemetry.LogInfo("TestCategory", "Test info message", nil)
	time.Sleep(1 * time.Second) // Ensure the log entry is written

	entries := telemetry.FilterLogs("INFO", "TestCategory", "Test info message", 10)
	if len(entries) == 0 {
		t.Errorf("Expected to find log entries, got 0")
	}
}
