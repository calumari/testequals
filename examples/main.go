package main

import (
	"fmt"
	"strings"

	"github.com/calumari/jwalk"

	"github.com/calumari/testequals"
)

// expected showcases every built-in directive exposed via JSON operators.
// Conventions:
//
//	$eq               strict equality for subtree (reject extra keys)
//	$ne               value must differ
//	$nil              value must be nil (true asserts, false => implicit)
//	$required         field must exist and not be zero / nil
//	$any              wildcard (always passes)
//	$regex            string must match pattern
//	$elementsMatch    order-insensitive exact multiset match
//	$length length    constraints (eq / lt / lte / gt / gte)
//	$empty            true => empty, false => non-empty
//	$lt/$lte/$gt/$gte numeric comparisons
//	$in               membership
//	$and/$or/$nor     logical combinators over inline expectations
//	$not              negation
const a = `{
  "user": {
    "$eq": {
      "id": { "$and": [{ "$required": true },  { "$not": 1 }] }, 
      "name": { "$regex": "^Al.*" }, 
      "age": { "$and": [{ "$gte": 30 }, { "$lt": 40 }] }, 
      "nicknames": { "$length": { "gte": 1, "lte": 3 } }, 
      "tags": { "$elementsMatch": ["blue", "green"] }, 
      "meta": { "$any": true }, 
      "optional": { "$nil": false }, 
      "maybeEmpty": { "$empty": false }, 
      "status": { "$in": ["active", "pending"] }, 
      "score": { "$and": [{ "$gt": 10 }, { "$lt": 100 }] }, 
      "disallowed": { "$nor": [{ "$eq": "banned" }, { "$eq": "disabled" }] }, 
      "note": { "$not": { "$regex": "ERROR" } }, 
      "different": { "$or": [{"$eq":123}] }
    }
  }
}`

// actual attempts to satisfy above; tweak to see mismatches.
const b = `{
  "user": {
    "id": 42,
    "name": "Alice",
    "age": 31,
    "nicknames": ["Al"],
    "tags": ["green","blue"],
    "meta": {"extra": true},
    "optional": "present",
    "maybeEmpty": [1],
    "status": "active",
    "score": 55,
    "disallowed": "ok",
    "note": "all good",
    "different": 456,
    "extra": true
  }
}`

func main() {
	reg := jwalk.DefaultRegistry()
	x := []*jwalk.Directive{
		testequals.TestEqualDirective,
		testequals.TestNotEqualDirective,
		testequals.TestNilDirective,
		testequals.TestRequiredDirective,
		testequals.TestAnyDirective,
		testequals.TestMatchStringDirective,
		testequals.TestElementsMatchDirective,
		testequals.TestLengthDirective,
		testequals.TestEmptyDirective,
		testequals.TestLessThanDirective,
		testequals.TestLessThanOrEqualDirective,
		testequals.TestGreaterThanDirective,
		testequals.TestGreaterThanOrEqualDirective,
		testequals.TestInDirective,
		testequals.TestAndDirective,
		testequals.TestOrDirective,
		testequals.TestNorDirective,
		testequals.TestNotDirective,
	}
	for _, v := range x {
		reg.Register(v)
	}

	var expected any
	if err := reg.Unmarshal([]byte(a), &expected); err != nil {
		panic(err)
	}
	var actual any
	if err := reg.Unmarshal([]byte(b), &actual); err != nil {
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
		// Single mismatch
		fmt.Println("Mismatch:", err)
		return
	}
	fmt.Println("All constraints satisfied")
	// Show subset nature: modify actual with extra field under strict $eq subtree -> mismatch
	mutated := strings.Replace(b, "\n  }\n}", ",\n    \"extraInsideEq\": true\n  }\n}", 1)
	var actual2 any
	if err := reg.Unmarshal([]byte(mutated), &actual2); err == nil {
		if err2 := tester.Test(expected, actual2); err2 != nil {
			fmt.Println("Adding extraInsideEq under $eq subtree causes mismatch:")
			fmt.Println(" -", err2)
		}
	}
}
