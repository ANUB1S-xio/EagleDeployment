package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TelemetryLevel controls what gets logged
type TelemetryLevel int

const (
	// LevelError only logs errors
	LevelError TelemetryLevel = iota
	// LevelWarning logs errors and warnings
	LevelWarning
	// LevelInfo logs errors, warnings, and info
	LevelInfo
	// LevelDebug logs everything
	LevelDebug
)

// Event represents a telemetry event
type Event struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Category  string                 `json:"category"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Line      int                    `json:"line,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	HostInfo  HostInfo               `json:"host_info,omitempty"`
}

// HostInfo contains information about the host system
type HostInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
}

// LogEntry represents a single entry in the log file
type LogEntry struct {
	Event       Event     `json:"event"`
	TimeWritten time.Time `json:"time_written"`
}

// LogFile represents the structure of our single JSON log file
type LogFile struct {
	AppName  string     `json:"app_name"`
	Version  string     `json:"version"`
	Entries  []LogEntry `json:"entries"`
	LastSync time.Time  `json:"last_sync"`
}

// Telemetry is the main telemetry manager
type Telemetry struct {
	sessionID     string
	level         TelemetryLevel
	logFilePath   string
	hostInfo      HostInfo
	enableConsole bool
	buffer        []LogEntry
	maxBufferSize int
	mu            sync.Mutex
	logFile       LogFile
	version       string
}

var (
	instance *Telemetry
	once     sync.Once
)

// GetInstance returns the singleton telemetry instance
func GetInstance() *Telemetry {
	once.Do(func() {
		// Set up log directory
		if err := os.MkdirAll("logs", 0755); err != nil {
			fmt.Printf("Failed to create logs directory: %v\n", err)
		}

		logFilePath := filepath.Join("logs", "eagledeploy.json")

		instance = &Telemetry{
			sessionID:     uuid.New().String(),
			level:         LevelInfo, // Default level
			enableConsole: true,
			logFilePath:   logFilePath,
			buffer:        make([]LogEntry, 0),
			maxBufferSize: 50, // Flush to disk after 50 entries
			version:       "1.0.0",
		}

		// Get host information
		hostname, _ := os.Hostname()
		instance.hostInfo = HostInfo{
			Hostname: hostname,
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
		}

		// Load existing log file or create a new one
		instance.loadLogFile()

		// Log session start - prevent a recursive call
		instance.logEventDirect("INFO", "Telemetry", "Session started", map[string]interface{}{
			"version": instance.version,
		})
	})
	return instance
}

// loadLogFile loads the existing log file or creates a new one
func (t *Telemetry) loadLogFile() {
	// Check if file exists
	if _, err := os.Stat(t.logFilePath); os.IsNotExist(err) {
		// Create new log file
		t.logFile = LogFile{
			AppName:  "EagleDeploy",
			Version:  t.version,
			Entries:  []LogEntry{},
			LastSync: time.Now(),
		}
		return
	}

	// Read existing file
	data, err := os.ReadFile(t.logFilePath)
	if err != nil {
		fmt.Printf("Failed to read log file: %v\n", err)
		// Create new log file
		t.logFile = LogFile{
			AppName:  "EagleDeploy",
			Version:  t.version,
			Entries:  []LogEntry{},
			LastSync: time.Now(),
		}
		return
	}

	// Parse JSON
	if err := json.Unmarshal(data, &t.logFile); err != nil {
		fmt.Printf("Failed to parse log file: %v\n", err)
		// Create new log file
		t.logFile = LogFile{
			AppName:  "EagleDeploy",
			Version:  t.version,
			Entries:  []LogEntry{},
			LastSync: time.Now(),
		}
	}
}

// saveLogFile saves the log file to disk
func (t *Telemetry) saveLogFile() error {
	// Update timestamp
	t.logFile.LastSync = time.Now()

	// Convert to JSON
	data, err := json.MarshalIndent(t.logFile, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(t.logFilePath, data, 0644)
}

// flushBuffer writes buffered entries to the log file
func (t *Telemetry) flushBuffer() {
	if len(t.buffer) == 0 {
		return
	}

	// Add buffered entries to log file
	t.logFile.Entries = append(t.logFile.Entries, t.buffer...)

	// Clear buffer
	t.buffer = make([]LogEntry, 0)

	// Save to disk
	if err := t.saveLogFile(); err != nil {
		fmt.Printf("Failed to save log file: %v\n", err)
	}
}

// SetLevel changes the current telemetry level
func (t *Telemetry) SetLevel(level TelemetryLevel) {
	t.mu.Lock()
	oldLevel := t.level
	t.level = level
	t.mu.Unlock()

	// Log the level change without calling LogInfo to avoid deadlock
	if oldLevel != level {
		t.logEventDirect("INFO", "Telemetry", "Log level changed", map[string]interface{}{
			"oldLevel": oldLevel,
			"newLevel": level,
		})
	}
}

// GetLevel returns the current telemetry level
func (t *Telemetry) GetLevel() TelemetryLevel {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.level
}

// EnableConsole turns console logging on/off
func (t *Telemetry) EnableConsole(enable bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enableConsole = enable
}

// LogError logs an error event
func (t *Telemetry) LogError(category, message string, data map[string]interface{}) {
	if t.level >= LevelError {
		t.logEvent("ERROR", category, message, data)
	}
}

// LogWarning logs a warning event
func (t *Telemetry) LogWarning(category, message string, data map[string]interface{}) {
	if t.level >= LevelWarning {
		t.logEvent("WARNING", category, message, data)
	}
}

// LogInfo logs an info event
func (t *Telemetry) LogInfo(category, message string, data map[string]interface{}) {
	if t.level >= LevelInfo {
		t.logEvent("INFO", category, message, data)
	}
}

// LogDebug logs a debug event
func (t *Telemetry) LogDebug(category, message string, data map[string]interface{}) {
	if t.level >= LevelDebug {
		t.logEvent("DEBUG", category, message, data)
	}
}

// Close closes the telemetry system
func (t *Telemetry) Close() {
	// Use direct logging to avoid potential deadlocks
	t.logEventDirect("INFO", "Telemetry", "Session ended", nil)
	t.flushBuffer()
}

// FilterLogs returns filtered log entries
func (t *Telemetry) FilterLogs(level, category, message string, limit int) []LogEntry {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Make a copy of log entries to prevent modification during filtering
	entries := append([]LogEntry{}, t.logFile.Entries...)

	if limit <= 0 || limit > len(entries) {
		limit = len(entries)
	}

	// Apply filters
	filteredEntries := make([]LogEntry, 0)
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]

		// Filter by level
		if level != "" && entry.Event.Level != level {
			continue
		}

		// Filter by category
		if category != "" && entry.Event.Category != category {
			continue
		}

		// Filter by message
		if message != "" && !strings.Contains(
			strings.ToLower(entry.Event.Message),
			strings.ToLower(message)) {
			continue
		}

		filteredEntries = append(filteredEntries, entry)

		// Apply limit
		if len(filteredEntries) >= limit {
			break
		}
	}

	return filteredEntries
}

// ClearLogs clears all log entries
func (t *Telemetry) ClearLogs() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Create a new log file
	t.logFile.Entries = []LogEntry{}
	t.buffer = []LogEntry{}

	// Save to disk
	return t.saveLogFile()
}

// logEventDirect logs an event without checking log level or acquiring locks
// Used internally to avoid deadlocks in key places
func (t *Telemetry) logEventDirect(level, category, message string, data map[string]interface{}) {
	// Get caller information
	pc, file, line, ok := runtime.Caller(2)
	var funcName, source string
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
		source = filepath.Base(file)
	}

	// Create event
	event := Event{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		Data:      data,
		Source:    source,
		Function:  funcName,
		Line:      line,
		SessionID: t.sessionID,
		HostInfo:  t.hostInfo,
	}

	// Create log entry
	entry := LogEntry{
		Event:       event,
		TimeWritten: time.Now(),
	}

	// Add to buffer directly without locking
	t.buffer = append(t.buffer, entry)

	// Print to console if enabled
	if t.enableConsole {
		levelColor := "\033[0m" // Reset
		switch level {
		case "ERROR":
			levelColor = "\033[31m" // Red
		case "WARNING":
			levelColor = "\033[33m" // Yellow
		case "INFO":
			levelColor = "\033[36m" // Cyan
		case "DEBUG":
			levelColor = "\033[35m" // Magenta
		}

		fmt.Printf("%s[%s] %s: %s%s\n",
			levelColor,
			event.Timestamp.Format("15:04:05"),
			level,
			event.Message,
			"\033[0m", // Reset color
		)
	}

	// Flush buffer if it's full - don't lock here
	if len(t.buffer) >= t.maxBufferSize {
		go t.flushBuffer() // Run in a goroutine to avoid blocking
	}
}

func (t *Telemetry) logEvent(level, category, message string, data map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get caller information
	pc, file, line, ok := runtime.Caller(2)
	var funcName, source string
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
		source = filepath.Base(file)
	}

	// Create event
	event := Event{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		Data:      data,
		Source:    source,
		Function:  funcName,
		Line:      line,
		SessionID: t.sessionID,
		HostInfo:  t.hostInfo,
	}

	// Create log entry
	entry := LogEntry{
		Event:       event,
		TimeWritten: time.Now(),
	}

	// Add to buffer
	t.buffer = append(t.buffer, entry)

	// Print to console if enabled
	if t.enableConsole {
		levelColor := "\033[0m" // Reset
		switch level {
		case "ERROR":
			levelColor = "\033[31m" // Red
		case "WARNING":
			levelColor = "\033[33m" // Yellow
		case "INFO":
			levelColor = "\033[36m" // Cyan
		case "DEBUG":
			levelColor = "\033[35m" // Magenta
		}

		fmt.Printf("%s[%s] %s: %s%s\n",
			levelColor,
			event.Timestamp.Format("15:04:05"),
			level,
			event.Message,
			"\033[0m", // Reset color
		)
	}

	// Flush buffer if it's full
	if len(t.buffer) >= t.maxBufferSize {
		t.flushBuffer()
	}
}

// GetEntries is an alias for FilterLogs for compatibility
func (t *Telemetry) GetEntries(level, category, message string, limit int) []LogEntry {
	return t.FilterLogs(level, category, message, limit)
}
