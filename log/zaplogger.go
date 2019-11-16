package log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a structure that provide a fast, leveled, structured logging using Uber zap.
// All methods are safe for concurrent use.
type ZapLogger struct {
	*zap.Logger
}

// Field is a marshaling operation used to add a key-value pair to a logger's context.
type Field = zap.Field

// New creates a new logger.
func New() (*ZapLogger, error) {
	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig.LevelKey = "severity"
	zapCfg.EncoderConfig.MessageKey = "message"
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(t.Unix())
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger}, nil
}

// NewNop returns a no-op Logger. It never writes out logs or internal errors.
func NewNop() *ZapLogger {
	return &ZapLogger{zap.NewNop()}
}

// Close closes the logger.
func (l *ZapLogger) Close() {
	// we dont care about the error.
	_ = l.Sync()
}

// String constructs a field with the given key and value.
func String(key string, val string) Field {
	return zap.String(key, val)
}

// ByteString constructs a field with the given key and value.
func ByteString(key string, val []byte) Field {
	return zap.ByteString(key, val)
}

// Stringer constructs a field with the given key and value.
func Stringer(key string, val fmt.Stringer) Field {
	return zap.Stringer(key, val)
}

// Bool constructs a field with the given key and value.
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Int constructs a field with the given key and value.
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Int32 constructs a field with the given key and value.
func Int32(key string, val int32) Field {
	return zap.Int32(key, val)
}

// Int64 constructs a field with the given key and value.
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Float32 constructs a field with the given key and value.
func Float32(key string, val float32) Field {
	return zap.Float32(key, val)
}

// Float64 constructs a field with the given key and value.
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Error constructs a field with the given key and value.
func Error(err error) Field {
	return zap.Error(err)
}

// Duration field.
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Time field.
func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}

// Any field.
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}
