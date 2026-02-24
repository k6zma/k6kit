package main

import (
	"errors"
	"time"

	"github.com/k6zma/k6kit/log"
)

func main() {
	l, err := log.New(log.Config{Level: log.LevelDebug, Format: log.FormatText, Color: true})
	if err != nil {
		panic(err)
	}

	errA := errors.New("first error")
	errB := errors.New("second error")
	now := time.Date(2026, 2, 23, 12, 30, 0, 0, time.UTC)

	fields := []log.Field{
		log.Rune("rune", 'A'),
		log.Byte("byte", byte('B')),
		log.Int("int", 1),
		log.Int8("int8", int8(8)),
		log.Int16("int16", int16(16)),
		log.Int32("int32", int32(32)),
		log.Int64("int64", int64(64)),
		log.Uint8("uint8", uint8(8)),
		log.Uint16("uint16", uint16(16)),
		log.Uint32("uint32", uint32(32)),
		log.Uint64("uint64", uint64(64)),
		log.Float32("float32", float32(3.2)),
		log.Float64("float64", 6.4),
		log.Bool("bool", true),
		log.String("string", "value"),
		log.Duration("duration", 2*time.Second),
		log.Time("time", now),
		log.Any("any", map[string]int{"k": 1}),
		log.Any("error", errA),
		log.Bytes("bytes", []byte("bin")),
		log.Strings("strings", []string{"a", "b"}),
		log.Runes("runes", []rune{'x', 'y'}),
		log.Bools("bools", []bool{true, false}),
		log.Ints("ints", []int{1, 2}),
		log.Int8s("int8s", []int8{8, 9}),
		log.Int16s("int16s", []int16{16, 17}),
		log.Int32s("int32s", []int32{32, 33}),
		log.Int64s("int64s", []int64{64, 65}),
		log.Uint8s("uint8s", []uint8{8, 9}),
		log.Uint16s("uint16s", []uint16{16, 17}),
		log.Uint32s("uint32s", []uint32{32, 33}),
		log.Uint64s("uint64s", []uint64{64, 65}),
		log.Float32s("float32s", []float32{3.1, 3.2}),
		log.Float64s("float64s", []float64{6.1, 6.2}),
		log.Any("errors", []error{errA, errB}),
		log.Group("group", log.String("nested", "yes"), log.Int("count", 2)),
	}

	l.Info("all field constructors", fields...)
}
