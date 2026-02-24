package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFieldConstructors(t *testing.T) {
	now := time.Unix(1700000000, 0)
	tests := []fieldConstructorCase{
		{name: "rune", field: Rune("r", 'x'), key: "r"},
		{name: "byte", field: Byte("b", byte('a')), key: "b"},
		{name: "int", field: Int("i", 1), key: "i"},
		{name: "int8", field: Int8("i8", 8), key: "i8"},
		{name: "int16", field: Int16("i16", 16), key: "i16"},
		{name: "int32", field: Int32("i32", 32), key: "i32"},
		{name: "int64", field: Int64("i64", 64), key: "i64"},
		{name: "uint8", field: Uint8("u8", 8), key: "u8"},
		{name: "uint16", field: Uint16("u16", 16), key: "u16"},
		{name: "uint32", field: Uint32("u32", 32), key: "u32"},
		{name: "uint64", field: Uint64("u64", 64), key: "u64"},
		{name: "float32", field: Float32("f32", 3.2), key: "f32"},
		{name: "float64", field: Float64("f64", 6.4), key: "f64"},
		{name: "bool", field: Bool("ok", true), key: "ok"},
		{name: "string", field: String("s", "v"), key: "s"},
		{name: "duration", field: Duration("d", time.Second), key: "d"},
		{name: "time", field: Time("t", now), key: "t"},
		{name: "any", field: Any("a", map[string]any{"k": 1}), key: "a"},
		{name: "bytes", field: Bytes("bs", []byte("x")), key: "bs"},
		{name: "strings", field: Strings("ss", []string{"a", "b"}), key: "ss"},
		{name: "runes", field: Runes("rs", []rune{'x'}), key: "rs"},
		{name: "bools", field: Bools("bo", []bool{true}), key: "bo"},
		{name: "ints", field: Ints("is", []int{1}), key: "is"},
		{name: "int8s", field: Int8s("i8s", []int8{8}), key: "i8s"},
		{name: "int16s", field: Int16s("i16s", []int16{16}), key: "i16s"},
		{name: "int32s", field: Int32s("i32s", []int32{32}), key: "i32s"},
		{name: "int64s", field: Int64s("i64s", []int64{64}), key: "i64s"},
		{name: "uint8s", field: Uint8s("u8s", []uint8{8}), key: "u8s"},
		{name: "uint16s", field: Uint16s("u16s", []uint16{16}), key: "u16s"},
		{name: "uint32s", field: Uint32s("u32s", []uint32{32}), key: "u32s"},
		{name: "uint64s", field: Uint64s("u64s", []uint64{64}), key: "u64s"},
		{name: "float32s", field: Float32s("f32s", []float32{3.2}), key: "f32s"},
		{name: "float64s", field: Float64s("f64s", []float64{6.4}), key: "f64s"},
		{name: "group", field: Group("g", String("k", "v")), key: "g"},
	}

	for i, tc := range tests {
		t.Run(testCaseName(testPrefixFields, tc.name, i), func(t *testing.T) {
			assert.Equal(t, tc.key, tc.field.Key)
		})
	}
}
