package testequals

import (
	"fmt"
	"reflect"
)

var nilableKinds = map[reflect.Kind]struct{}{
	reflect.Chan:      {},
	reflect.Func:      {},
	reflect.Interface: {},
	reflect.Map:       {},
	reflect.Pointer:   {},
	reflect.Slice:     {},
}

func isNil(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	if _, ok := nilableKinds[v.Kind()]; ok && v.IsNil() {
		return true
	}
	return false
}

func isZero(v reflect.Value) bool {
	return !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func isList(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func isEmpty(v reflect.Value) bool {
	if isNil(v) {
		return true
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return v.Len() == 0
	default:
		return isZero(v)
	}
}

// fastPrimitiveEqual attempts a high-performance comparison for primitive
// scalar types. It returns (handled, err). When handled is true, either the
// values were equal (err == nil) or a mismatch is described by err.
func fastPrimitiveEqual(a, b any) (handled bool, err error) {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		if !ok {
			return true, fmt.Errorf("expected string %q, got %T", av, b)
		}
		if av != bv {
			return true, fmt.Errorf("expected string %q, got %q", av, bv)
		}
		return true, nil
	case bool:
		bv, ok := b.(bool)
		if !ok {
			return true, fmt.Errorf("expected bool %v, got %T", av, b)
		}
		if av != bv {
			return true, fmt.Errorf("expected bool %v, got %v", av, bv)
		}
		return true, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		ai := toInt64(av)
		switch bv := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			bi := toInt64(bv)
			if ai != bi {
				return true, fmt.Errorf("expected int %d, got %d", ai, bi)
			}
			return true, nil
		case float32, float64:
			bf, _ := toFloat64(bv)
			if float64(ai) != bf {
				return true, fmt.Errorf("expected int %d, got float %v", ai, bf)
			}
			return true, nil
		default:
			return true, fmt.Errorf("expected integer (%d), got %T", ai, b)
		}
	case float32, float64:
		af, _ := toFloat64(av)
		switch bv := b.(type) {
		case float32, float64:
			bf, _ := toFloat64(bv)
			if af != bf {
				return true, fmt.Errorf("expected float %v, got %v", af, bf)
			}
			return true, nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			bi := toInt64(bv)
			if af != float64(bi) {
				return true, fmt.Errorf("expected float %v, got int %d", af, bi)
			}
			return true, nil
		default:
			return true, fmt.Errorf("expected float (%v), got %T", af, b)
		}
	}
	return false, nil
}
