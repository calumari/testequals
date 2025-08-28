package testequals

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isNil(t *testing.T) {
	t.Run("untyped nil returns true", func(t *testing.T) {
		assert.True(t, isNil(reflect.ValueOf(interface{}(nil))))
	})

	t.Run("typed nil pointer returns true", func(t *testing.T) {
		assert.True(t, isNil(reflect.ValueOf((*int)(nil))))
	})

	t.Run("non nil slice returns false", func(t *testing.T) {
		assert.False(t, isNil(reflect.ValueOf([]int{})))
	})
}

func Test_isZero(t *testing.T) {
	t.Run("zero int returns true", func(t *testing.T) {
		assert.True(t, isZero(reflect.ValueOf(0)))
	})

	t.Run("non zero int returns false", func(t *testing.T) {
		assert.False(t, isZero(reflect.ValueOf(1)))
	})

	t.Run("nil pointer returns true", func(t *testing.T) {
		assert.True(t, isZero(reflect.ValueOf((*int)(nil))))
	})
}

func Test_isList(t *testing.T) {
	t.Run("slice returns true", func(t *testing.T) {
		assert.True(t, isList(reflect.ValueOf([]int{1})))
	})

	t.Run("array returns true", func(t *testing.T) {
		assert.True(t, isList(reflect.ValueOf([1]int{1})))
	})

	t.Run("map returns false", func(t *testing.T) {
		assert.False(t, isList(reflect.ValueOf(map[string]int{"a": 1})))
	})
}

func Test_isEmpty(t *testing.T) {
	t.Run("nil slice returns true", func(t *testing.T) {
		assert.True(t, isEmpty(reflect.ValueOf([]int(nil))))
	})

	t.Run("empty slice returns true", func(t *testing.T) {
		assert.True(t, isEmpty(reflect.ValueOf([]int{})))
	})

	t.Run("non-empty slice returns false", func(t *testing.T) {
		assert.False(t, isEmpty(reflect.ValueOf([]int{1})))
	})

	t.Run("nil map returns true", func(t *testing.T) {
		assert.True(t, isEmpty(reflect.ValueOf(map[string]int(nil))))
	})

	t.Run("empty map returns true", func(t *testing.T) {
		assert.True(t, isEmpty(reflect.ValueOf(map[string]int{})))
	})

	t.Run("non-empty map returns false", func(t *testing.T) {
		assert.False(t, isEmpty(reflect.ValueOf(map[string]int{"a": 1})))
	})

	t.Run("zero int returns true", func(t *testing.T) {
		assert.True(t, isEmpty(reflect.ValueOf(0)))
	})

	t.Run("non-zero int returns false", func(t *testing.T) {
		assert.False(t, isEmpty(reflect.ValueOf(1)))
	})
}

func Test_fastPrimitiveEqual(t *testing.T) {
	// string tests
	t.Run("string vs non-string returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual("a", 1)
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("string mismatch returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual("a", "b")
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("equal strings returns no error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual("a", "a")
		assert.True(t, handled)
		assert.NoError(t, err)
	})

	// bool tests
	t.Run("bool vs non-bool returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(true, 1)
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("bool mismatch returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(true, false)
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("equal bools returns no error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(true, true)
		assert.True(t, handled)
		assert.NoError(t, err)
	})

	// int tests
	t.Run("int vs non-number returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5, "a")
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("int vs float mismatch returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5, 5.1)
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("equal ints returns no error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5, 5)
		assert.True(t, handled)
		assert.NoError(t, err)
	})

	t.Run("int vs float equal returns no error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5, 5.0)
		assert.True(t, handled)
		assert.NoError(t, err)
	})

	// float tests
	t.Run("float vs non-number returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5.1, "a")
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("float vs int mismatch returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5.1, 5)
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("float vs float mismatch returns error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(float32(5.1), 5.2)
		assert.True(t, handled)
		assert.Error(t, err)
	})

	t.Run("float vs int equal returns no error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5.0, 5)
		assert.True(t, handled)
		assert.NoError(t, err)
	})

	t.Run("equal floats returns no error", func(t *testing.T) {
		handled, err := fastPrimitiveEqual(5.1, 5.1)
		assert.True(t, handled)
		assert.NoError(t, err)
	})

	// misc tests
	t.Run("non-primitive returns not handled", func(t *testing.T) {
		handled, err := fastPrimitiveEqual([]int{1}, []int{1})
		assert.False(t, handled)
		assert.NoError(t, err)
	})
}
