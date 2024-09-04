package logger_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/bartmika/ipfs-cli-wrapper/internal/logger"
)

func TestNewProviderInitialization(t *testing.T) {
	// Redirect stdout to capture log output.
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer func() {
		os.Stdout = stdout
	}()
	os.Stdout = w

	// Initialize the logger using the NewProvider function.
	log := logger.NewProvider()

	// Log a test message to verify the logger output.
	log.Debug("Testing debug message")

	// Close the writer to flush the buffer.
	w.Close()

	// Read the output from the pipe.
	outBytes := make([]byte, 1024)
	n, _ := r.Read(outBytes)
	buf.Write(outBytes[:n])

	// Check that the logger is not nil.
	if log == nil {
		t.Fatalf("Expected logger instance, got nil")
	}

	// Check that the output contains the debug message.
	logOutput := buf.String()
	expected := "Testing debug message"
	if !strings.Contains(logOutput, expected) {
		t.Errorf("Expected log output to contain %q, but got %q", expected, logOutput)
	}
}

// func TestNewProviderLoggingLevel(t *testing.T) {
// 	// Initialize the logger.
// 	log := logger.NewProvider()
//
// 	// Check that the default logging level is set to debug.
// 	var loggingLevel slog.LevelVar
// 	if log.Handler().Options().Level == nil || log.Handler().Options().Level.Level() != slog.LevelDebug {
// 		t.Errorf("Expected logging level to be 'Debug', but got %v", log.Handler().Options().Level)
// 	}
// }
