package testequals

import (
	"fmt"
)

func fastPrimitiveEqual(a, b any) (equal bool, handled bool, msg string) {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		if !ok {
			return false, true, fmt.Sprintf("expected string %q, got %T", av, b)
		}
		if av != bv {
			return false, true, fmt.Sprintf("expected string %q, got %q", av, bv)
		}
		return true, true, ""
	case bool:
		bv, ok := b.(bool)
		if !ok {
			return false, true, fmt.Sprintf("expected bool %v, got %T", av, b)
		}
		if av != bv {
			return false, true, fmt.Sprintf("expected bool %v, got %v", av, bv)
		}
		return true, true, ""
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		ai := toInt64(av)
		switch bv := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			bi := toInt64(bv)
			if ai != bi {
				return false, true, fmt.Sprintf("expected int %d, got %d", ai, bi)
			}
			return true, true, ""
		case float32, float64:
			bf := toFloat64(bv)
			if float64(ai) != bf {
				return false, true, fmt.Sprintf("expected int %d, got float %v", ai, bf)
			}
			return true, true, ""
		default:
			return false, true, fmt.Sprintf("expected integer (%d), got %T", ai, b)
		}
	case float32, float64:
		af := toFloat64(av)
		switch bv := b.(type) {
		case float32, float64:
			bf := toFloat64(bv)
			if af != bf {
				return false, true, fmt.Sprintf("expected float %v, got %v", af, bf)
			}
			return true, true, ""
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			bi := toInt64(bv)
			if af != float64(bi) {
				return false, true, fmt.Sprintf("expected float %v, got int %d", af, bi)
			}
			return true, true, ""
		default:
			return false, true, fmt.Sprintf("expected float (%v), got %T", af, b)
		}
	}
	return false, false, ""
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case uint:
		return int64(n)
	case uint8:
		return int64(n)
	case uint16:
		return int64(n)
	case uint32:
		return int64(n)
	case uint64:
		return int64(n)
	default:
		return 0
	}
}

func toFloat64(v any) float64 {
	switch n := v.(type) {
	case float32:
		return float64(n)
	case float64:
		return n
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	default:
		return 0
	}
}
