package testequals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_toInt64(t *testing.T) {
	t.Run("int converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(int(5)))
	})

	t.Run("int8 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(int8(5)))
	})

	t.Run("int16 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(int16(5)))
	})

	t.Run("int32 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(int32(5)))
	})

	t.Run("int64 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(int64(5)))
	})

	t.Run("uint converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(uint(5)))
	})

	t.Run("uint8 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(uint8(5)))
	})

	t.Run("uint16 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(uint16(5)))
	})

	t.Run("uint32 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(uint32(5)))
	})

	t.Run("uint64 converts", func(t *testing.T) {
		assert.Equal(t, int64(5), toInt64(uint64(5)))
	})

	t.Run("non-numeric returns 0", func(t *testing.T) {
		assert.Equal(t, int64(0), toInt64("5"))
	})
}

func Test_toFloat64(t *testing.T) {
	t.Run("float32 converts", func(t *testing.T) {
		got, ok := toFloat64(float32(5.25))
		assert.Equal(t, 5.25, got)
		assert.True(t, ok)
	})

	t.Run("float64 converts", func(t *testing.T) {
		got, ok := toFloat64(float64(6.5))
		assert.Equal(t, 6.5, got)
		assert.True(t, ok)
	})

	t.Run("int converts", func(t *testing.T) {
		got, ok := toFloat64(int(7))
		assert.Equal(t, 7.0, got)
		assert.True(t, ok)
	})

	t.Run("int8 converts", func(t *testing.T) {
		got, ok := toFloat64(int8(8))
		assert.Equal(t, 8.0, got)
		assert.True(t, ok)
	})

	t.Run("int16 converts", func(t *testing.T) {
		got, ok := toFloat64(int16(9))
		assert.Equal(t, 9.0, got)
		assert.True(t, ok)
	})

	t.Run("int32 converts", func(t *testing.T) {
		got, ok := toFloat64(int32(10))
		assert.Equal(t, 10.0, got)
		assert.True(t, ok)
	})

	t.Run("int64 converts", func(t *testing.T) {
		got, ok := toFloat64(int64(11))
		assert.Equal(t, 11.0, got)
		assert.True(t, ok)
	})

	t.Run("uint converts", func(t *testing.T) {
		got, ok := toFloat64(uint(12))
		assert.Equal(t, 12.0, got)
		assert.True(t, ok)
	})

	t.Run("uint8 converts", func(t *testing.T) {
		got, ok := toFloat64(uint8(13))
		assert.Equal(t, 13.0, got)
		assert.True(t, ok)
	})

	t.Run("uint16 converts", func(t *testing.T) {
		got, ok := toFloat64(uint16(14))
		assert.Equal(t, 14.0, got)
		assert.True(t, ok)
	})

	t.Run("uint32 converts", func(t *testing.T) {
		got, ok := toFloat64(uint32(15))
		assert.Equal(t, 15.0, got)
		assert.True(t, ok)
	})

	t.Run("uint64 converts", func(t *testing.T) {
		got, ok := toFloat64(uint64(16))
		assert.Equal(t, 16.0, got)
		assert.True(t, ok)
	})

	t.Run("non-numeric returns false", func(t *testing.T) {
		got, ok := toFloat64("17")
		assert.Equal(t, 0.0, got)
		assert.False(t, ok)
	})
}

func Test_trimFloat(t *testing.T) {
	t.Run("int represented succeeds", func(t *testing.T) {
		assert.Equal(t, int64(5), trimFloat(5.0))
	})

	t.Run("fractional stays float succeeds", func(t *testing.T) {
		assert.Equal(t, 5.25, trimFloat(5.25))
	})
}
