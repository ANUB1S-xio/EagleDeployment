package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestTelemetry creates a test instance of Telemetry
func setupTestTelemetry(t *testing.T) (*Telemetry, func()) {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "telemetry_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a test instance
	testLogPath := filepath.Join(tempDir, "test_logs.json")
	telemetry := &Telemetry{
		sessionID:     "test-session-id",
		level:         LevelInfo,
		enableConsole: false, // Disable console output during tests
		logFilePath:   testLogPath,
		buffer:        make([]LogEntry, 0),
		maxBufferSize: 5, // Small buffer size for testing
		version:       "1.0.0-test",
		logFile: LogFile{
			AppName:  "TestApp",
			Version:  "1.0.0-test",
			Entries:  []LogEntry{},
			LastSync: time.Now(),
		},
	}

	// Return the instance and a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return telemetry, cleanup
}

// TestGetInstance tests that the singleton pattern works
func TestGetInstance(t *testing.T) {
	instance1 := GetInstance()
	instance2 := GetInstance()

	if instance1 != instance2 {
		t.Error("GetInstance should return the same instance")
	}
}

// TestLogLevels tests all log level functions
func TestLogLevels(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Test each log level
	telemetry.LogError("TestCategory", "Test error message", nil)
	telemetry.LogWarning("TestCategory", "Test warning message", nil)
	telemetry.LogInfo("TestCategory", "Test info message", nil)
	telemetry.LogDebug("TestCategory", "Test debug message", nil)

	// Force flush buffer to ensure logs are written
	telemetry.flushBuffer()

	// Check log file contains appropriate entries
	if len(telemetry.logFile.Entries) != 3 { // DEBUG shouldn't be logged at INFO level
		t.Errorf("Expected 3 log entries but got %d", len(telemetry.logFile.Entries))
	}
}

// TestSetLevel tests changing the log level
func TestSetLevel(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Set level to Error (only errors should be logged)
	telemetry.SetLevel(LevelError)

	telemetry.LogError("TestCategory", "Test error message", nil)
	telemetry.LogWarning("TestCategory", "Test warning message", nil)
	telemetry.flushBuffer()

	if len(telemetry.logFile.Entries) != 2 { // Error + "Log level changed" info
		t.Errorf("Expected 2 log entries at Error level but got %d", len(telemetry.logFile.Entries))
	}

	// Set level to Debug (everything should be logged)
	telemetry.SetLevel(LevelDebug)
	telemetry.LogDebug("TestCategory", "Test debug message", nil)
	telemetry.flushBuffer()

	if len(telemetry.logFile.Entries) != 4 { // Previous 2 + "Log level changed" + Debug
		t.Errorf("Expected 4 log entries at Debug level but got %d", len(telemetry.logFile.Entries))
	}
}

// TestFilterLogs tests the log filtering functionality
func TestFilterLogs(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Add some test logs with different levels and categories
	telemetry.LogError("Category1", "Error message", nil)
	telemetry.LogWarning("Category2", "Warning message", nil)
	telemetry.LogInfo("Category1", "Info message", nil)
	telemetry.LogInfo("Category3", "Another info message", nil)
	telemetry.flushBuffer()

	// Test filtering by level
	errorLogs := telemetry.FilterLogs("ERROR", "", "", 100)
	if len(errorLogs) != 1 {
		t.Errorf("Expected 1 error log but got %d", len(errorLogs))
	}

	// Test filtering by category
	category1Logs := telemetry.FilterLogs("", "Category1", "", 100)
	if len(category1Logs) != 2 {
		t.Errorf("Expected 2 Category1 logs but got %d", len(category1Logs))
	}

	// Test filtering by message content
	anotherLogs := telemetry.FilterLogs("", "", "Another", 100)
	if len(anotherLogs) != 1 {
		t.Errorf("Expected 1 log containing 'Another' but got %d", len(anotherLogs))
	}

	// Test limit
	limitedLogs := telemetry.FilterLogs("", "", "", 2)
	if len(limitedLogs) != 2 {
		t.Errorf("Expected 2 logs with limit but got %d", len(limitedLogs))
	}
}

// TestClearLogs tests clearing all logs
func TestClearLogs(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Add some test logs
	telemetry.LogError("TestCategory", "Test error message", nil)
	telemetry.LogInfo("TestCategory", "Test info message", nil)
	telemetry.flushBuffer()

	// Verify logs were added
	if len(telemetry.logFile.Entries) == 0 {
		t.Error("Expected log entries to be added but found none")
	}

	// Clear logs
	err := telemetry.ClearLogs()
	if err != nil {
		t.Errorf("Failed to clear logs: %v", err)
	}

	// Verify logs were cleared
	if len(telemetry.logFile.Entries) != 0 {
		t.Errorf("Expected 0 log entries after clear but got %d", len(telemetry.logFile.Entries))
	}
}

// TestSaveLoadLogFile tests saving and loading the log file
func TestSaveLoadLogFile(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Add a test log
	testMessage := "Test save/load message"
	telemetry.LogInfo("TestCategory", testMessage, nil)

	// Force save to disk
	telemetry.flushBuffer()

	// Create a new telemetry instance that will load the file
	newTelemetry := &Telemetry{
		logFilePath: telemetry.logFilePath,
	}
	newTelemetry.loadLogFile()

	// Check if the log message was loaded correctly
	var found bool
	for _, entry := range newTelemetry.logFile.Entries {
		if entry.Event.Message == testMessage {
			found = true
			break
		}
	}

	if !found {
		t.Error("Failed to load log message from file")
	}
}

// TestLogWithData tests logging with additional data
func TestLogWithData(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Create test data
	testData := map[string]interface{}{
		"user":    "testuser",
		"action":  "login",
		"success": true,
		"count":   42,
	}

	// Log with data
	telemetry.LogInfo("Auth", "User login", testData)
	telemetry.flushBuffer()

	// Check if data was saved correctly
	if len(telemetry.logFile.Entries) == 0 {
		t.Fatal("No log entries found")
	}

	entry := telemetry.logFile.Entries[0]
	if entry.Event.Category != "Auth" || entry.Event.Message != "User login" {
		t.Error("Log entry basic fields don't match expected values")
	}

	// Check data fields
	if entry.Event.Data["user"] != "testuser" ||
		entry.Event.Data["action"] != "login" ||
		entry.Event.Data["success"] != true ||
		entry.Event.Data["count"] != float64(42) { // JSON converts to float64
		t.Error("Log entry data doesn't match expected values")
	}
}

// TestMaxBufferSizeTriggersFlush tests that buffer is automatically flushed when full
func TestMaxBufferSizeTriggersFlush(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Buffer size is set to 5 in setupTestTelemetry
	// Add exactly maxBufferSize entries
	for i := 0; i < telemetry.maxBufferSize; i++ {
		telemetry.LogInfo("TestCategory", "Test message", nil)
	}

	// Give a moment for any goroutine that might be flushing to complete
	time.Sleep(100 * time.Millisecond)

	// Check if log file has entries (buffer should have been flushed)
	if len(telemetry.logFile.Entries) < telemetry.maxBufferSize {
		t.Errorf("Expected at least %d entries in log file after buffer flush, got %d",
			telemetry.maxBufferSize, len(telemetry.logFile.Entries))
	}
}

// TestLogFormatting tests the JSON structure of logged events
func TestLogFormatting(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// Add a log entry
	telemetry.LogInfo("TestCategory", "Test message", map[string]interface{}{
		"key": "value",
	})
	telemetry.flushBuffer()

	// Read the raw file to check JSON structure
	data, err := os.ReadFile(telemetry.logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Check that it's valid JSON
	var logFile LogFile
	err = json.Unmarshal(data, &logFile)
	if err != nil {
		t.Fatalf("Failed to parse log file JSON: %v", err)
	}

	// Check basic structure
	if logFile.AppName != "TestApp" || logFile.Version != "1.0.0-test" {
		t.Error("Log file header doesn't match expected values")
	}

	// Check entry structure
	if len(logFile.Entries) == 0 {
		t.Fatal("No entries in log file")
	}

	entry := logFile.Entries[0]
	if entry.Event.Level != "INFO" ||
		entry.Event.Category != "TestCategory" ||
		entry.Event.Message != "Test message" ||
		entry.Event.Data["key"] != "value" {
		t.Error("Log entry doesn't match expected structure")
	}
}

// TestEnabledDisableConsole tests enabling and disabling console output
func TestEnableDisableConsole(t *testing.T) {
	telemetry, cleanup := setupTestTelemetry(t)
	defer cleanup()

	// This is primarily a functional test since we can't easily check console output
	// Just make sure the method doesn't crash or cause errors
	telemetry.EnableConsole(true)
	telemetry.LogInfo("Console", "Should show in console if enabled", nil)

	telemetry.EnableConsole(false)
	telemetry.LogInfo("Console", "Should not show in console", nil)
}
