package log

// Logger defines a leveled logging structure.
type Logger interface {
	// Debug logs a message at DebugLevel.
	Debug(msg string, fields ...Field)
	// Info logs a message at InfoLevel.
	Info(msg string, fields ...Field)
	// Warn logs a message at WarnLevel.
	Warn(msg string, fields ...Field)
	// Error logs a message at ErrorLevel.
	Error(msg string, fields ...Field)
	// Fatal logs a message at FatalLevel.
	Fatal(msg string, fields ...Field)
	// Panic logs a message at PanicLevel.
	Panic(msg string, fields ...Field)
	// Close closes the logger.
	Close()
}
