// Package logger provides a logging utility for the IPFS CLI Wrapper project.
// It offers a pre-configured logger instance with customizable log levels and
// additional options such as source file information to aid in debugging and monitoring.
package logger

import (
	"log/slog"
	"os"
)

// NewProvider initializes and returns a new instance of slog.Logger with default settings
// suitable for application-wide logging. The logger is configured to output log messages
// to the standard output (stdout) in a text format and includes additional information
// such as source file names and line numbers to aid in debugging.
//
// The logger uses a logging level variable (`loggingLevel`) which is set to `LevelDebug`
// by default, allowing all log messages at the debug level and above (info, warning, error, etc.)
// to be recorded. The logging level can be adjusted dynamically during runtime, making the logger
// versatile for different stages of development and production environments.
//
// Usage:
//
//	logger := logger.NewProvider()
//	logger.Debug("This is a debug message.")
//
// Returns:
//   - *slog.Logger: A pointer to the configured logger instance that can be used for structured logging
//     within the application.
//
// Key Features:
//   - **Logging Level:** The logger is initialized with a logging level of `LevelDebug`, allowing for
//     detailed log messages suitable for development and debugging. The level can be adjusted dynamically
//     at runtime using the `loggingLevel` variable.
//   - **Source Information:** The logger is configured with the `AddSource` option enabled, which includes
//     source file names and line numbers in the log output. This feature is particularly useful for tracing
//     the origin of log messages within the codebase.
//   - **Text Output Format:** The logger uses a text handler (`slog.NewTextHandler`) to format the log
//     messages as plain text, which is directed to the standard output (`os.Stdout`). This format is
//     human-readable and suitable for console output or basic logging needs.
//
// Example:
//
//	// Initialize the logger
//	logger := logger.NewProvider()
//
//	// Log a debug message
//	logger.Debug("Debugging the application startup.")
//
//	// Change the logging level dynamically
//	logger.SetLevel(slog.LevelInfo)
//
// Notes:
//   - The logger is currently set to output all debug-level messages and above by default.
//     This configuration can be modified to suit the needs of production environments where
//     more restrictive logging levels may be desired.
//   - The logger does not set itself as the default logger for the `slog` package (`slog.SetDefault`
//     is commented out). This allows flexibility in choosing whether to use this logger as the
//     primary logger for the application or alongside other logging configurations.
func NewProvider() *slog.Logger {
	// Create a logging level variable with Info as the default level
	var loggingLevel = new(slog.LevelVar)

	// Initialize the logger with a text handler, adding source file information and the ability to change levels dynamically
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: loggingLevel}))

	// Set the default logging level to debug
	loggingLevel.Set(slog.LevelDebug)

	// Return the configured logger instance
	return logger
}
