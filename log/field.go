package log

import (
	"log/slog"
	"time"
)

// Field is the public structured logging attribute abstraction
type Field struct {
	attr slog.Attr
}

// toAttr converts the Field to a slog.Attr
func (f Field) toAttr() slog.Attr {
	return f.attr
}

// makeField creates a Field from a key and value
func makeField(key string, value any) Field {
	return Field{attr: slog.Any(key, value)}
}

// Rune creates a rune field
func Rune(key string, value rune) Field {
	return makeField(key, value)
}

// Byte creates a byte field
func Byte(key string, value byte) Field {
	return makeField(key, value)
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{attr: slog.Int(key, value)}
}

// Int8 creates an int8 field
func Int8(key string, value int8) Field {
	return makeField(key, value)
}

// Int16 creates an int16 field
func Int16(key string, value int16) Field {
	return makeField(key, value)
}

// Int32 creates an int32 field
func Int32(key string, value int32) Field {
	return makeField(key, value)
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{attr: slog.Int64(key, value)}
}

// Uint8 creates an uint8 field
func Uint8(key string, value uint8) Field {
	return makeField(key, value)
}

// Uint16 creates an uint16 field
func Uint16(key string, value uint16) Field {
	return makeField(key, value)
}

// Uint32 creates an uint32 field
func Uint32(key string, value uint32) Field {
	return makeField(key, value)
}

// Uint64 creates an uint64 field
func Uint64(key string, value uint64) Field {
	return makeField(key, value)
}

// Float32 creates a float32 field
func Float32(key string, value float32) Field {
	return makeField(key, value)
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{attr: slog.Float64(key, value)}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{attr: slog.Bool(key, value)}
}

// String creates a string field
func String(key, value string) Field {
	return Field{attr: slog.String(key, value)}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{attr: slog.Duration(key, value)}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{attr: slog.Time(key, value)}
}

// Any creates a field from an arbitrary value
func Any(key string, value any) Field {
	return makeField(key, value)
}

// Error creates an error field
func Error(key string, err error) Field {
	return makeField(key, err)
}

// Bytes creates a []byte field
func Bytes(key string, value []byte) Field {
	return makeField(key, value)
}

// Strings creates a []string field
func Strings(key string, value []string) Field {
	return makeField(key, value)
}

// Runes creates a []rune field
func Runes(key string, value []rune) Field {
	return makeField(key, value)
}

// Bools creates a []bool field
func Bools(key string, value []bool) Field {
	return makeField(key, value)
}

// Ints creates a []int field
func Ints(key string, value []int) Field {
	return makeField(key, value)
}

// Int8s creates a []int8 field
func Int8s(key string, value []int8) Field {
	return makeField(key, value)
}

// Int16s creates a []int16 field
func Int16s(key string, value []int16) Field {
	return makeField(key, value)
}

// Int32s creates a []int32 field
func Int32s(key string, value []int32) Field {
	return makeField(key, value)
}

// Int64s creates a []int64 field
func Int64s(key string, value []int64) Field {
	return makeField(key, value)
}

// Uint8s creates a []uint8 field
func Uint8s(key string, value []uint8) Field {
	return makeField(key, value)
}

// Uint16s creates a []uint16 field
func Uint16s(key string, value []uint16) Field {
	return makeField(key, value)
}

// Uint32s creates a []uint32 field
func Uint32s(key string, value []uint32) Field {
	return makeField(key, value)
}

// Uint64s creates a []uint64 field
func Uint64s(key string, value []uint64) Field {
	return makeField(key, value)
}

// Float32s creates a []float32 field
func Float32s(key string, value []float32) Field {
	return makeField(key, value)
}

// Float64s creates a []float64 field
func Float64s(key string, value []float64) Field {
	return makeField(key, value)
}

// Errors creates a []error field
func Errors(key string, value []error) Field {
	return makeField(key, value)
}

// Group creates a grouped field that will be flattened with dot ыnotation
func Group(name string, fields ...Field) Field {
	attrs := make([]any, 0, len(fields))

	for _, f := range fields {
		attrs = append(attrs, f.toAttr())
	}

	return Field{attr: slog.Group(name, attrs...)}
}
