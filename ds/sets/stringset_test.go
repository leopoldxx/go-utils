package sets_test

import (
	"testing"

	"github.com/leopoldxx/go-utils/ds/sets"
)

func TestStringSet(t *testing.T) {
	left := sets.NewStringSet("a", "b", "c", "d", "e")
	right := sets.NewStringSet("d", "e", "f", "g", "h")
	right2 := sets.NewStringSet("a", "b", "f", "n", "m")

	if !left.Diff(right.Union(right2)).Equal(sets.NewStringSet("c")) {
		t.Fatal("should equal")
	}

	if !left.Has("a") {
		t.Fatal("should have 'a'")
	}

	if left.Len() != right.Len() || left.Len() != 5 || len(left.List()) != len(left.SortedList()) {
		t.Fatal("not same length")
	}

	res := left.Diff(right)
	if !res.Equal(sets.NewStringSet("a", "b", "c")) {
		t.Fatal("invalid diff: %v", res)
	}

	res = right.Diff(left)
	if !res.Equal(sets.NewStringSet("f", "g", "h")) {
		t.Fatal("invalid diff: %v", res)
	}

	if !left.Intersection(right).Equal(sets.NewStringSet("d", "e")) {
		t.Fatal("invalid intersection: %v", res)
	}

	left.Delete("a")

	if left.Has("a") {
		t.Fatal("should not have 'a'")
	}

	if !left.HasAll("b", "c", "d", "e") {
		t.Fatal("should have all")
	}

	if !left.HasAny("b", "c", "m", "n") {
		t.Fatal("should have any")
	}

	if !left.Union(right).Equal(sets.NewStringSet("b", "c", "d", "e", "f", "g", "h")) {
		t.Fatal("invalid union")
	}

	if !left.IsSuperset(sets.NewStringSet("c", "d")) {
		t.Fatal("invalid superset")
	}

	left.PopAny()
	if left.Len() != 3 {
		t.Fatal("invalid popany")
	}

	_, ok := left.PopAny()
	for ok {
		_, ok = left.PopAny()
	}
	if left.Len() != 0 {
		t.Fatal("invalid popany2")
	}
	t.Logf("%v", left)
	left.Replace(right.List()...)
	t.Logf("%v", left)
	if !left.Equal(right) {
		t.Fatal("should equal")
	}
}
