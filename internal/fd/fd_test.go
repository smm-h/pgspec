package fd

import (
	"reflect"
	"sort"
	"testing"
)

func TestClosure_SingleChain(t *testing.T) {
	// A→B, B→C, C→D: closure of {A} should be {A,B,C,D}
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B"}},
		{Determinant: []string{"B"}, Dependent: []string{"C"}},
		{Determinant: []string{"C"}, Dependent: []string{"D"}},
	}
	got := Closure([]string{"A"}, fds)
	want := []string{"A", "B", "C", "D"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Closure({A}) = %v, want %v", got, want)
	}
}

func TestClosure_CompositeKey(t *testing.T) {
	// A→C, BC→D: closure of {A,B} should be {A,B,C,D}
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"C"}},
		{Determinant: []string{"B", "C"}, Dependent: []string{"D"}},
	}
	got := Closure([]string{"A", "B"}, fds)
	want := []string{"A", "B", "C", "D"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Closure({A,B}) = %v, want %v", got, want)
	}
}

func TestClosure_NoFDs(t *testing.T) {
	got := Closure([]string{"A", "B"}, nil)
	want := []string{"A", "B"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Closure({A,B}, nil) = %v, want %v", got, want)
	}
}

func TestClosure_MultipleDependents(t *testing.T) {
	// A→{B,C}: closure of {A} should be {A,B,C}
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B", "C"}},
	}
	got := Closure([]string{"A"}, fds)
	want := []string{"A", "B", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Closure({A}) = %v, want %v", got, want)
	}
}

func TestMinimalCover_RemovesRedundant(t *testing.T) {
	// A→B, B→C, A→C: the FD A→C is redundant (derivable via A→B, B→C)
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B"}},
		{Determinant: []string{"B"}, Dependent: []string{"C"}},
		{Determinant: []string{"A"}, Dependent: []string{"C"}},
	}
	got := MinimalCover(fds)
	want := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B"}},
		{Determinant: []string{"B"}, Dependent: []string{"C"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MinimalCover = %v, want %v", got, want)
	}
}

func TestMinimalCover_RemovesExtraneousLHS(t *testing.T) {
	// A→B, AB→C should simplify to A→B, A→C (since A→B makes B extraneous in AB→C)
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B"}},
		{Determinant: []string{"A", "B"}, Dependent: []string{"C"}},
	}
	got := MinimalCover(fds)
	want := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B"}},
		{Determinant: []string{"A"}, Dependent: []string{"C"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MinimalCover = %v, want %v", got, want)
	}
}

func TestMinimalCover_DecomposesRHS(t *testing.T) {
	// A→{B,C} should decompose into A→B, A→C
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B", "C"}},
	}
	got := MinimalCover(fds)
	want := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B"}},
		{Determinant: []string{"A"}, Dependent: []string{"C"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MinimalCover = %v, want %v", got, want)
	}
}

func TestCandidateKeys_CompositeAndAlternate(t *testing.T) {
	// R(A,B,C,D), FDs: AB→CD, C→A
	// Keys should be {A,B} and {B,C}
	allAttrs := []string{"A", "B", "C", "D"}
	fds := []FuncDep{
		{Determinant: []string{"A", "B"}, Dependent: []string{"C", "D"}},
		{Determinant: []string{"C"}, Dependent: []string{"A"}},
	}
	got := CandidateKeys(allAttrs, fds)
	want := [][]string{{"A", "B"}, {"B", "C"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CandidateKeys = %v, want %v", got, want)
	}
}

func TestCandidateKeys_SingleKey(t *testing.T) {
	// R(A,B,C), FDs: A→B, A→C — only key is {A}
	allAttrs := []string{"A", "B", "C"}
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"B"}},
		{Determinant: []string{"A"}, Dependent: []string{"C"}},
	}
	got := CandidateKeys(allAttrs, fds)
	want := [][]string{{"A"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CandidateKeys = %v, want %v", got, want)
	}
}

func TestCandidateKeys_AllAttrsAreKey(t *testing.T) {
	// R(A,B,C) with no FDs — the only candidate key is {A,B,C}
	allAttrs := []string{"A", "B", "C"}
	got := CandidateKeys(allAttrs, nil)
	want := [][]string{{"A", "B", "C"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CandidateKeys = %v, want %v", got, want)
	}
}

func TestIsSuperkey_True(t *testing.T) {
	allAttrs := []string{"A", "B", "C", "D"}
	fds := []FuncDep{
		{Determinant: []string{"A", "B"}, Dependent: []string{"C", "D"}},
	}
	if !IsSuperkey([]string{"A", "B"}, allAttrs, fds) {
		t.Error("expected {A,B} to be a superkey")
	}
}

func TestIsSuperkey_False(t *testing.T) {
	allAttrs := []string{"A", "B", "C", "D"}
	fds := []FuncDep{
		{Determinant: []string{"A", "B"}, Dependent: []string{"C", "D"}},
	}
	if IsSuperkey([]string{"A"}, allAttrs, fds) {
		t.Error("expected {A} to not be a superkey")
	}
}

func TestIsSuperkey_SupersetOfKey(t *testing.T) {
	allAttrs := []string{"A", "B", "C", "D"}
	fds := []FuncDep{
		{Determinant: []string{"A", "B"}, Dependent: []string{"C", "D"}},
	}
	if !IsSuperkey([]string{"A", "B", "C"}, allAttrs, fds) {
		t.Error("expected {A,B,C} to be a superkey (superset of key {A,B})")
	}
}

func TestIsPrime_True(t *testing.T) {
	keys := [][]string{{"A", "B"}, {"B", "C"}}
	if !IsPrime("A", keys) {
		t.Error("expected A to be prime (in key {A,B})")
	}
	if !IsPrime("B", keys) {
		t.Error("expected B to be prime (in both keys)")
	}
	if !IsPrime("C", keys) {
		t.Error("expected C to be prime (in key {B,C})")
	}
}

func TestIsPrime_False(t *testing.T) {
	keys := [][]string{{"A", "B"}, {"B", "C"}}
	if IsPrime("D", keys) {
		t.Error("expected D to not be prime")
	}
}

func TestIsSubset(t *testing.T) {
	if !isSubset([]string{"A", "B"}, []string{"A", "B", "C"}) {
		t.Error("{A,B} should be subset of {A,B,C}")
	}
	if isSubset([]string{"A", "D"}, []string{"A", "B", "C"}) {
		t.Error("{A,D} should not be subset of {A,B,C}")
	}
	if !isSubset(nil, []string{"A"}) {
		t.Error("empty set should be subset of anything")
	}
}

func TestSetUnion(t *testing.T) {
	got := setUnion([]string{"A", "C"}, []string{"B", "C"})
	want := []string{"A", "B", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("setUnion = %v, want %v", got, want)
	}
}

func TestSetEquals(t *testing.T) {
	if !setEquals([]string{"A", "B"}, []string{"A", "B"}) {
		t.Error("equal sets should be equal")
	}
	if setEquals([]string{"A", "B"}, []string{"A", "C"}) {
		t.Error("different sets should not be equal")
	}
	if setEquals([]string{"A"}, []string{"A", "B"}) {
		t.Error("different-length sets should not be equal")
	}
}

func TestClosure_IsDeterministic(t *testing.T) {
	fds := []FuncDep{
		{Determinant: []string{"A"}, Dependent: []string{"C", "B"}},
		{Determinant: []string{"B"}, Dependent: []string{"D"}},
	}
	for i := 0; i < 10; i++ {
		got := Closure([]string{"A"}, fds)
		if !sort.StringsAreSorted(got) {
			t.Fatalf("closure result not sorted: %v", got)
		}
	}
}
