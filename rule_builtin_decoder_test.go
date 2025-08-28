package testequals

import (
	"strings"
	"testing"

	"github.com/calumari/jwalk"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_unmarshalEqual(t *testing.T) {
	t.Run("basic value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader("42"))
		got, err := unmarshalEqual(dec)
		require.NoError(t, err)
		assert.Equal(t, &Equal{expected: float64(42)}, got)
	})

	t.Run("nil value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader("null"))
		got, err := unmarshalEqual(dec)
		require.NoError(t, err)
		assert.Equal(t, &Equal{expected: nil}, got)
	})
}

func Test_unmarshalNil(t *testing.T) {
	t.Run("true expected succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader("true"))
		got, err := unmarshalNil(true)(dec)
		require.NoError(t, err)
		assert.Equal(t, &Nil{expected: true, wanted: true}, got)
	})

	t.Run("false expected succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader("false"))
		got, err := unmarshalNil(false)(dec)
		require.NoError(t, err)
		assert.Equal(t, &Nil{expected: false, wanted: false}, got)
	})
}

func Test_unmarshalRequired(t *testing.T) {
	t.Run("true expected succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader("true"))
		got, err := unmarshalRequired(true)(dec)
		require.NoError(t, err)
		assert.Equal(t, &Required{true}, got)
	})

	t.Run("false expected returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader("false"))
		got, err := unmarshalRequired(true)(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func Test_unmarshalAny(t *testing.T) {
	t.Run("any value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`true`))
		got, err := unmarshalAny(dec)
		require.NoError(t, err)
		assert.IsType(t, &Any{}, got)
	})
}

func Test_unmarshalMatchString(t *testing.T) {
	t.Run("valid regex succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`"^abc$"`))
		got, err := unmarshalMatchString(dec)
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "^abc$", got.re.String())
	})

	t.Run("invalid regex returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`"*invalid"`))
		got, err := unmarshalMatchString(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func Test_unmarshalElementsMatch(t *testing.T) {
	t.Run("valid array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[1, 2, 3]`))
		got, err := unmarshalElementsMatch(dec)
		require.NoError(t, err)
		assert.Equal(t, &ElementsMatch{expected: []any{float64(1), float64(2), float64(3)}}, got)
	})

	t.Run("empty array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[]`))
		got, err := unmarshalElementsMatch(dec)
		require.NoError(t, err)
		assert.Equal(t, &ElementsMatch{expected: []any{}}, got)
	})
}

func Test_unmarshalLength(t *testing.T) {
	t.Run("eq field succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"eq": 5}`))
		got, err := unmarshalLength(dec)
		require.NoError(t, err)
		assert.Equal(t, 5, *got.eq)
	})

	t.Run("negative eq returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"eq": -1}`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("lt field succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"lt": 2}`))
		got, err := unmarshalLength(dec)
		require.NoError(t, err)
		assert.Equal(t, 2, *got.lt)
	})

	t.Run("negative lt value returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"lt": -1}`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("lte field succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"lte": 10}`))
		got, err := unmarshalLength(dec)
		require.NoError(t, err)
		assert.Equal(t, 10, *got.lte)
	})

	t.Run("negative lte value returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"lte": -5}`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("gt field succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"gt": 1}`))
		got, err := unmarshalLength(dec)
		require.NoError(t, err)
		assert.Equal(t, 1, *got.gt)
	})

	t.Run("negative gt value returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"gt": -2}`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("gte field succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"gte": 0}`))
		got, err := unmarshalLength(dec)
		require.NoError(t, err)
		assert.Equal(t, 0, *got.gte)
	})

	t.Run("negative gte value returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"gte": -10}`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("float value returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`3.14`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("simple int value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`3`))
		got, err := unmarshalLength(dec)
		require.NoError(t, err)
		assert.Equal(t, 3, *got.eq)
	})

	t.Run("negative int value returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`-1`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("no comparator returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{}`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("multiple fields returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`{"eq": 1, "lt": 2}`))
		got, err := unmarshalLength(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func Test_unmarshalEmpty(t *testing.T) {
	t.Run("true value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`true`))
		got, err := unmarshalEmpty(dec)
		require.NoError(t, err)
		assert.Equal(t, &Empty{true}, got)
	})

	t.Run("false value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`false`))
		got, err := unmarshalEmpty(dec)
		require.NoError(t, err)
		assert.Equal(t, &Empty{false}, got)
	})
}

func Test_unmarshalNotEqual(t *testing.T) {
	t.Run("basic value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`42`))
		got, err := unmarshalNotEqual(dec)
		require.NoError(t, err)
		assert.Equal(t, &NotEqual{expected: float64(42)}, got)
	})

	t.Run("nil value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`null`))
		got, err := unmarshalNotEqual(dec)
		require.NoError(t, err)
		assert.Equal(t, &NotEqual{expected: nil}, got)
	})
}

func Test_unmarshalLT(t *testing.T) {
	t.Run("valid number succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`5`))
		got, err := unmarshalLT(false)(dec)
		require.NoError(t, err)
		assert.Equal(t, &numericCompare{op: "lt", ref: 5, incl: false}, got)
	})
}

func Test_unmarshalGT(t *testing.T) {
	t.Run("valid number succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`5`))
		got, err := unmarshalGT(true)(dec)
		require.NoError(t, err)
		assert.Equal(t, &numericCompare{op: "gt", ref: 5, incl: true}, got)
	})
}

func Test_unmarshalIn(t *testing.T) {
	t.Run("valid array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[1, 2]`))
		got, err := unmarshalIn(dec)
		require.NoError(t, err)
		assert.Equal(t, &InSet{[]any{float64(1), float64(2)}}, got)
	})

	t.Run("empty array returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[]`))
		got, err := unmarshalIn(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("nested directive succeeds", func(t *testing.T) {
		reg, err := jwalk.NewRegistry(jwalk.WithDirective(TestInDirective), jwalk.WithDirective(TestEqualDirective))
		require.NoError(t, err)
		var got any
		err = json.UnmarshalRead(strings.NewReader(`{"$in": [{"$eq": 5}]}`), &got, json.WithUnmarshalers(jwalk.Unmarshalers(reg)))
		require.NoError(t, err)
		assert.Equal(t, &InSet{elems: []any{&Equal{expected: float64(5)}}}, got)
	})
}

func Test_unmarshalAnd(t *testing.T) {
	t.Run("valid array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[1, 2]`))
		got, err := unmarshalAnd(dec)
		require.NoError(t, err)
		assert.Equal(t, &And{[]any{float64(1), float64(2)}}, got)
	})

	t.Run("empty array returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[]`))
		got, err := unmarshalAnd(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("nested directive succeeds", func(t *testing.T) {
		reg, err := jwalk.NewRegistry(jwalk.WithDirective(TestAndDirective), jwalk.WithDirective(TestEqualDirective))
		require.NoError(t, err)
		var got any
		err = json.UnmarshalRead(strings.NewReader(`{"$and": [{"$eq": 5}]}`), &got, json.WithUnmarshalers(jwalk.Unmarshalers(reg)))
		require.NoError(t, err)
		assert.Equal(t, &And{rules: []any{&Equal{expected: float64(5)}}}, got)
	})
}

func Test_unmarshalOr(t *testing.T) {
	t.Run("valid array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[1, 2]`))
		got, err := unmarshalOr(dec)
		require.NoError(t, err)
		assert.Equal(t, &Or{[]any{float64(1), float64(2)}}, got)
	})

	t.Run("empty array returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[]`))
		got, err := unmarshalOr(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("nested directive succeeds", func(t *testing.T) {
		reg, err := jwalk.NewRegistry(jwalk.WithDirective(TestOrDirective), jwalk.WithDirective(TestEqualDirective))
		require.NoError(t, err)
		var got any
		err = json.UnmarshalRead(strings.NewReader(`{"$or": [{"$eq": 5}]}`), &got, json.WithUnmarshalers(jwalk.Unmarshalers(reg)))
		require.NoError(t, err)
		assert.Equal(t, &Or{rules: []any{&Equal{expected: float64(5)}}}, got)
	})
}

func Test_unmarshalNor(t *testing.T) {
	t.Run("valid array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[1, 2]`))
		got, err := unmarshalNor(dec)
		require.NoError(t, err)
		assert.Equal(t, &Nor{[]any{float64(1), float64(2)}}, got)
	})

	t.Run("empty array returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[]`))
		got, err := unmarshalNor(dec)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("nested directive succeeds", func(t *testing.T) {
		reg, err := jwalk.NewRegistry(jwalk.WithDirective(TestNorDirective), jwalk.WithDirective(TestEqualDirective))
		require.NoError(t, err)
		var got any
		err = json.UnmarshalRead(strings.NewReader(`{"$nor": [{"$eq": 5}]}`), &got, json.WithUnmarshalers(jwalk.Unmarshalers(reg)))
		require.NoError(t, err)
		assert.Equal(t, &Nor{rules: []any{&Equal{expected: float64(5)}}}, got)
	})
}

func Test_unmarshalNot(t *testing.T) {
	t.Run("basic value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`42`))
		got, err := unmarshalNot(dec)
		require.NoError(t, err)
		assert.Equal(t, &Not{rule: float64(42)}, got)
	})

	t.Run("nil value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`null`))
		got, err := unmarshalNot(dec)
		require.NoError(t, err)
		assert.Equal(t, &Not{rule: nil}, got)
	})

	t.Run("nested directive succeeds", func(t *testing.T) {
		reg, err := jwalk.NewRegistry(jwalk.WithDirective(TestNotDirective), jwalk.WithDirective(TestEqualDirective))
		require.NoError(t, err)
		var got any
		err = json.UnmarshalRead(strings.NewReader(`{"$not": {"$eq": 5}}`), &got, json.WithUnmarshalers(jwalk.Unmarshalers(reg)))
		require.NoError(t, err)
		assert.Equal(t, &Not{rule: &Equal{expected: float64(5)}}, got)
	})
}

func Test_decodeNumber(t *testing.T) {
	t.Run("int value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`5`))
		got, err := decodeNumber(dec)
		require.NoError(t, err)
		assert.Equal(t, float64(5), got)
	})

	t.Run("float value succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`3.14`))
		got, err := decodeNumber(dec)
		require.NoError(t, err)
		assert.Equal(t, 3.14, got)
	})

	t.Run("non-numeric returns error", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`"not a number"`))
		got, err := decodeNumber(dec)
		assert.Error(t, err)
		assert.Equal(t, float64(0), got)
	})
}

func Test_decodeArray(t *testing.T) {
	t.Run("valid array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[1, 2]`))
		got, err := decodeArray(dec)
		require.NoError(t, err)
		assert.Equal(t, []any{float64(1), float64(2)}, got)
	})

	t.Run("empty array succeeds", func(t *testing.T) {
		dec := jsontext.NewDecoder(strings.NewReader(`[]`))
		got, err := decodeArray(dec)
		require.NoError(t, err)
		assert.Equal(t, []any{}, got)
	})
}
