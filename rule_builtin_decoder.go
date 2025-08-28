package testequals

import (
	"errors"
	"regexp"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// All builtin directive decoder/unmarshal functions moved from builtin.go

func unmarshalEqual(dec *jsontext.Decoder) (*Equal, error) {
	var expected any
	if err := json.UnmarshalDecode(dec, &expected); err != nil {
		return nil, err
	}
	return &Equal{expected}, nil
}

func unmarshalNil(expected bool) func(dec *jsontext.Decoder) (*Nil, error) {
	return func(dec *jsontext.Decoder) (*Nil, error) {
		var want bool
		if err := json.UnmarshalDecode(dec, &want); err != nil {
			return nil, err
		}
		return &Nil{expected: expected, wanted: want}, nil
	}
}

func unmarshalRequired(expected bool) func(dec *jsontext.Decoder) (*Required, error) {
	return func(dec *jsontext.Decoder) (*Required, error) {
		var want bool
		if err := json.UnmarshalDecode(dec, &want); err != nil {
			return nil, err
		}
		if !want {
			return nil, errors.New("required directive must be true")
		}
		return &Required{expected && want}, nil
	}
}

func unmarshalAny(dec *jsontext.Decoder) (*Any, error) {
	var discard any
	if err := json.UnmarshalDecode(dec, &discard); err != nil {
		return nil, err
	}
	return &Any{}, nil
}

func unmarshalMatchString(dec *jsontext.Decoder) (*MatchString, error) {
	var expr string
	if err := json.UnmarshalDecode(dec, &expr); err != nil {
		return nil, err
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &MatchString{re}, nil
}

func unmarshalElementsMatch(dec *jsontext.Decoder) (*ElementsMatch, error) {
	var expected []any
	if err := json.UnmarshalDecode(dec, &expected); err != nil {
		return nil, err
	}
	return &ElementsMatch{expected}, nil
}

func unmarshalLength(dec *jsontext.Decoder) (*Length, error) {
	type aux struct {
		Eq  *int `json:"eq,omitempty"`
		Lt  *int `json:"lt,omitempty"`
		Lte *int `json:"lte,omitempty"`
		Gt  *int `json:"gt,omitempty"`
		Gte *int `json:"gte,omitempty"`
	}

	c := &Length{}

	setIntPtr := func(p **int, v int) {
		*p = new(int)
		**p = v
	}

	validateInt := func(n int) error {
		if n < 0 {
			return errors.New("length must be non-negative")
		}
		return nil
	}

	if dec.PeekKind() == '{' {
		var a aux
		if err := json.UnmarshalDecode(dec, &a); err != nil {
			return nil, err
		}

		if a.Eq != nil {
			if err := validateInt(*a.Eq); err != nil {
				return nil, err
			}
			setIntPtr(&c.eq, *a.Eq)
		}
		if a.Lt != nil {
			if err := validateInt(*a.Lt); err != nil {
				return nil, err
			}
			setIntPtr(&c.lt, *a.Lt)
		}
		if a.Lte != nil {
			if err := validateInt(*a.Lte); err != nil {
				return nil, err
			}
			setIntPtr(&c.lte, *a.Lte)
		}
		if a.Gt != nil {
			if err := validateInt(*a.Gt); err != nil {
				return nil, err
			}
			setIntPtr(&c.gt, *a.Gt)
		}
		if a.Gte != nil {
			if err := validateInt(*a.Gte); err != nil {
				return nil, err
			}
			setIntPtr(&c.gte, *a.Gte)
		}
	} else {
		var num float64
		if err := json.UnmarshalDecode(dec, &num); err != nil {
			return nil, errors.New("invalid length directive payload")
		}
		if num != float64(int(num)) {
			return nil, errors.New("length value must be integer")
		}
		if num < 0 {
			return nil, errors.New("length must be non-negative")
		}
		i := int(num)
		c.eq = &i
	}

	if c.eq == nil && c.lt == nil && c.lte == nil && c.gt == nil && c.gte == nil {
		return nil, errors.New("no length comparator provided")
	}
	if c.eq != nil && (c.lt != nil || c.lte != nil || c.gt != nil || c.gte != nil) {
		return nil, errors.New("eq cannot be combined with other length comparators")
	}

	return c, nil
}

func unmarshalEmpty(dec *jsontext.Decoder) (*Empty, error) {
	var want bool
	if err := json.UnmarshalDecode(dec, &want); err != nil {
		return nil, err
	}
	return &Empty{want}, nil
}

func unmarshalNotEqual(dec *jsontext.Decoder) (*NotEqual, error) {
	var exp any
	if err := json.UnmarshalDecode(dec, &exp); err != nil {
		return nil, err
	}
	return &NotEqual{exp}, nil
}

func unmarshalLT(incl bool) func(*jsontext.Decoder) (*numericCompare, error) {
	return func(dec *jsontext.Decoder) (*numericCompare, error) {
		ref, err := decodeNumber(dec)
		if err != nil {
			return nil, err
		}
		return &numericCompare{op: "lt", ref: ref, incl: incl}, nil
	}
}

func unmarshalGT(incl bool) func(*jsontext.Decoder) (*numericCompare, error) {
	return func(dec *jsontext.Decoder) (*numericCompare, error) {
		ref, err := decodeNumber(dec)
		if err != nil {
			return nil, err
		}
		return &numericCompare{op: "gt", ref: ref, incl: incl}, nil
	}
}

func unmarshalIn(dec *jsontext.Decoder) (*InSet, error) {
	var arr []any
	if err := json.UnmarshalDecode(dec, &arr); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, errors.New("in directive requires non-empty array")
	}
	return &InSet{arr}, nil
}

func unmarshalAnd(dec *jsontext.Decoder) (*And, error) {
	arr, err := decodeArray(dec)
	if err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, errors.New("and directive requires non-empty array")
	}
	return &And{arr}, nil
}

func unmarshalOr(dec *jsontext.Decoder) (*Or, error) {
	arr, err := decodeArray(dec)
	if err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, errors.New("or directive requires non-empty array")
	}
	return &Or{arr}, nil
}

func unmarshalNor(dec *jsontext.Decoder) (*Nor, error) {
	arr, err := decodeArray(dec)
	if err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, errors.New("nor directive requires non-empty array")
	}
	return &Nor{arr}, nil
}

func unmarshalNot(dec *jsontext.Decoder) (*Not, error) {
	var v any
	if err := json.UnmarshalDecode(dec, &v); err != nil {
		return nil, err
	}
	return &Not{v}, nil
}

func decodeNumber(dec *jsontext.Decoder) (float64, error) {
	var n any
	if err := json.UnmarshalDecode(dec, &n); err != nil {
		return 0, err
	}
	if f, ok := toFloat64(n); ok {
		return f, nil
	}
	return 0, errors.New("expected numeric value")
}

func decodeArray(dec *jsontext.Decoder) ([]any, error) {
	var arr []any
	if err := json.UnmarshalDecode(dec, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}
