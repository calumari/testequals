package main

import (
	"fmt"

	"github.com/calumari/jwalk"
	"github.com/calumari/testequals"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// expected document (left) using $eq rule expressing required subset structure.
const a = `{
	"user": {
		"$eq": {
			"name": "Alice",
			"age": 30,
			"address": { "city": "Paris" }
		}
	}
}`

// actual document (right) differing in several fields and containing extras.
const b = `{
	"user": {
		"name": "Alicia",
		"age": 31,
		"address": { "country": "FR" },
		"extra": true
	}
}`

func init() {
	// Register the $eq rule for use in JSON documents.
	jwalk.MustRegister(jwalk.DefaultRegistry, "eq", func(dec *jsontext.Decoder) (*testequals.Equal, error) {
		var vv any
		if err := json.UnmarshalDecode(dec, &vv); err != nil {
			return nil, err
		}
		return testequals.NewEqual(vv), nil
	})
}

func main() {
	var expected any
	if err := json.Unmarshal([]byte(a), &expected, json.WithUnmarshalers(jwalk.Unmarshalers(jwalk.DefaultRegistry))); err != nil {
		panic(err)
	}
	var actual any
	if err := json.Unmarshal([]byte(b), &actual, json.WithUnmarshalers(jwalk.Unmarshalers(jwalk.DefaultRegistry))); err != nil {
		panic(err)
	}

	tester := testequals.New(testequals.WithCollectAll())

	if err := tester.Test(expected, actual); err != nil {
		if me, ok := err.(*testequals.MultiError); ok {
			fmt.Printf("%d mismatches:\n", len(me.Mismatches))
			for _, m := range me.Mismatches {
				fmt.Println(" -", m)
			}
			return
		}
		panic(err)
	}
	fmt.Println("Match: actual satisfies expected subset")
}
