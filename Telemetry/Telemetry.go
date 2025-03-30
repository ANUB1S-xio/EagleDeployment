package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// TelemetryLevel controls what gets logged
type TelemetryLevel int

const (
	LevelError TelemetryLevel = iota
	LevelWarning
	LevelInfo
	LevelDebug
)

// Telemetry represents the telemetry logger
type Telemetry struct {
	file          *os.File
	mu            sync.Mutex
	closed        bool
	level         TelemetryLevel
	enableConsole bool
	buffer        []Event
	maxBufferSize int
}

// Event represents a telemetry event
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// Singleton instance of Telemetry
var instance *Telemetry
var once sync.Once

// GetInstance returns the singleton instance of Telemetry
func GetInstance() *Telemetry {
	once.Do(func() {
		instance = &Telemetry{
			level:         LevelInfo,
			enableConsole: false,
			buffer:        make([]Event, 0),
			maxBufferSize: 10,
		}
	})
	return instance
}

// New creates a new telemetry logger
func New() (*Telemetry, error) {
	// Define the log file name
	logFileName := "eagledeployment.log"

	// Open or create the log file in the project's root directory
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Verify that the file was created successfully
	if _, err := os.Stat(logFileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to create log file: %s", logFileName)
	}

	// Initialize the Telemetry instance
	t := GetInstance()
	t.file = file
	return t, nil
}

// SetLevel sets the logging level
func (t *Telemetry) SetLevel(level TelemetryLevel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.level = level
}

// EnableConsole enables or disables console logging
func (t *Telemetry) EnableConsole(enable bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enableConsole = enable
}

// LogError logs an error message
func (t *Telemetry) LogError(category, message string, payload map[string]interface{}) {
	t.log("ERROR", category, message, payload)
}

// LogWarning logs a warning message
func (t *Telemetry) LogWarning(category, message string, payload map[string]interface{}) {
	t.log("WARNING", category, message, payload)
}

// LogInfo logs an info message
func (t *Telemetry) LogInfo(category, message string, payload map[string]interface{}) {
	t.log("INFO", category, message, payload)
}

// LogDebug logs a debug message
func (t *Telemetry) LogDebug(category, message string, payload map[string]interface{}) {
	t.log("DEBUG", category, message, payload)
}

// log is a helper method to log events
func (t *Telemetry) log(level, category, message string, payload map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if the log level is enabled
	if !t.isLevelEnabled(level) {
		return
	}

	event := Event{
		Type:      level,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"category": category,
			"message":  message,
			"data":     payload,
		},
	}

	// Add to buffer
	t.buffer = append(t.buffer, event)

	// Print to console if enabled
	if t.enableConsole {
		t.printToConsole(event)
	}

	// Flush buffer if full
	if len(t.buffer) >= t.maxBufferSize {
		t.flushBuffer()
	}
}

// isLevelEnabled checks if the given log level is enabled
func (t *Telemetry) isLevelEnabled(level string) bool {
	levelMap := map[string]TelemetryLevel{
		"ERROR":   LevelError,
		"WARNING": LevelWarning,
		"INFO":    LevelInfo,
		"DEBUG":   LevelDebug,
	}
	return levelMap[level] <= t.level
}

// printToConsole prints an event to the console
func (t *Telemetry) printToConsole(event Event) {
	data, _ := json.Marshal(event)
	println(string(data))
}

// flushBuffer writes the buffer to the log file
func (t *Telemetry) flushBuffer() {
	if t.file == nil {
		return
	}

	for _, event := range t.buffer {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		t.file.WriteString(string(data) + "\n")
	}

	// Clear the buffer
	t.buffer = make([]Event, 0)
}

// FilterLogs filters logs based on criteria
func (t *Telemetry) FilterLogs(level, category, message string, limit int) []Event {
	t.mu.Lock()
	defer t.mu.Unlock()

	var filtered []Event
	for _, event := range t.buffer {
		if level != "" && event.Type != level {
			continue
		}
		if category != "" && event.Payload["category"] != category {
			continue
		}
		if message != "" && event.Payload["message"] != message {
			continue
		}
		filtered = append(filtered, event)
		if len(filtered) >= limit {
			break
		}
	}
	return filtered
}

// ClearLogs clears all logs
func (t *Telemetry) ClearLogs() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.buffer = make([]Event, 0)
	if t.file != nil {
		err := t.file.Truncate(0)
		if err != nil {
			return err
		}
		_, err = t.file.Seek(0, 0)
		return err
	}
	return nil
}

// Close closes the telemetry logger
func (t *Telemetry) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return os.ErrClosed
	}

	t.closed = true
	return t.file.Close()
}
