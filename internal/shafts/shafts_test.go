package shafts

import (
	"strings"
	"testing"
)

// must builds a set or fails the test.
func must(t *testing.T, list ...int) Set {
	t.Helper()
	out, err := New(list...)
	if err != nil {
		t.Fatalf("New(%v): %v", list, err)
	}
	return out
}

func TestNew(t *testing.T) {
	set := must(t, 1, 3, 5)
	if set.Count() != 3 {
		t.Fatalf("Count = %d, want 3", set.Count())
	}
	for _, shaft := range []int{1, 3, 5} {
		if !set.Has(shaft) {
			t.Fatalf("the set does not hold shaft %d", shaft)
		}
	}
	for _, shaft := range []int{2, 4, 6, 0, 33} {
		if set.Has(shaft) {
			t.Fatalf("the set holds shaft %d", shaft)
		}
	}
	// The order the shafts are given in makes no difference.
	if !must(t, 5, 1, 3).Equal(set) {
		t.Fatalf("the same shafts given in another order gave another set")
	}
	empty, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if !empty.IsEmpty() || empty.Count() != 0 {
		t.Fatalf("New() = %s", empty)
	}
	for label, list := range map[string][]int{
		"a shaft of zero":           {0},
		"a negative shaft":          {-1},
		"a shaft off the loom":      {33},
		"a shaft named twice":       {1, 1},
		"a shaft named twice later": {1, 2, 3, 2},
	} {
		if _, err := New(list...); err == nil {
			t.Fatalf("%s: New(%v) = nil error, want a failure", label, list)
		}
	}
}

func TestAddAndRemove(t *testing.T) {
	set := Set(0).Add(2).Add(4)
	if !set.Equal(must(t, 2, 4)) {
		t.Fatalf("Add gave %s", set)
	}
	// Adding a shaft that is already there changes nothing.
	if !set.Add(2).Equal(set) {
		t.Fatalf("adding a shaft twice changed the set")
	}
	if got := set.Remove(2); !got.Equal(must(t, 4)) {
		t.Fatalf("Remove gave %s", got)
	}
	// Removing a shaft that is not there changes nothing.
	if !set.Remove(7).Equal(set) {
		t.Fatalf("removing a shaft that is not there changed the set")
	}
	// A shaft off the loom is neither added nor removed.
	if !set.Add(99).Equal(set) || !set.Remove(99).Equal(set) {
		t.Fatalf("a shaft off the loom changed the set")
	}
}

func TestShaftsAndHighest(t *testing.T) {
	set := must(t, 2, 3, 7)
	list := set.Shafts()
	if len(list) != 3 || list[0] != 2 || list[1] != 3 || list[2] != 7 {
		t.Fatalf("Shafts = %v", list)
	}
	if set.Highest() != 7 {
		t.Fatalf("Highest = %d, want 7", set.Highest())
	}
	// The list belongs to the caller.
	list[0] = 99
	if set.Shafts()[0] != 2 {
		t.Fatalf("writing to the list changed the set to %s", set)
	}
	empty, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if len(empty.Shafts()) != 0 || empty.Highest() != 0 {
		t.Fatalf("the empty set has shafts %v and highest %d", empty.Shafts(), empty.Highest())
	}
}

func TestSetArithmetic(t *testing.T) {
	first := must(t, 1, 2, 3)
	second := must(t, 3, 4)
	if got := first.Union(second); !got.Equal(must(t, 1, 2, 3, 4)) {
		t.Fatalf("Union = %s", got)
	}
	if got := first.Intersect(second); !got.Equal(must(t, 3)) {
		t.Fatalf("Intersect = %s", got)
	}
	if got := first.Difference(second); !got.Equal(must(t, 1, 2)) {
		t.Fatalf("Difference = %s", got)
	}
	complement, err := first.Complement(6)
	if err != nil {
		t.Fatalf("Complement: %v", err)
	}
	if !complement.Equal(must(t, 4, 5, 6)) {
		t.Fatalf("Complement = %s", complement)
	}
	// What is lifted and what stays down cover the loom between them and share nothing.
	if got := first.Union(complement); !got.Equal(must(t, 1, 2, 3, 4, 5, 6)) {
		t.Fatalf("the set and its complement cover %s", got)
	}
	if !first.Intersect(complement).IsEmpty() {
		t.Fatalf("the set and its complement share shafts")
	}
	// The complement of the complement is the set again.
	back, err := complement.Complement(6)
	if err != nil {
		t.Fatalf("Complement: %v", err)
	}
	if !back.Equal(first) {
		t.Fatalf("the complement of the complement = %s", back)
	}
	for _, loom := range []int{0, 1, 33} {
		if _, err := first.Complement(loom); err == nil {
			t.Fatalf("Complement on a loom of %d = nil error, want a failure", loom)
		}
	}
}

func TestCompare(t *testing.T) {
	if got := Compare(must(t, 1, 2), must(t, 1, 2)); got != 0 {
		t.Fatalf("Compare of a set with itself = %d", got)
	}
	// Fewer shafts come first.
	if got := Compare(must(t, 1), must(t, 1, 2)); got != -1 {
		t.Fatalf("Compare = %d, want -1", got)
	}
	if got := Compare(must(t, 1, 2), must(t, 1)); got != 1 {
		t.Fatalf("Compare = %d, want 1", got)
	}
	// Then by the shafts themselves.
	if got := Compare(must(t, 1, 2), must(t, 1, 3)); got != -1 {
		t.Fatalf("Compare = %d, want -1", got)
	}
	sets := []Set{must(t, 3, 4), must(t, 2), must(t, 1, 2)}
	Sort(sets)
	if !sets[0].Equal(must(t, 2)) {
		t.Fatalf("Sort put %s first", sets[0])
	}
	if !sets[1].Equal(must(t, 1, 2)) || !sets[2].Equal(must(t, 3, 4)) {
		t.Fatalf("Sort gave %s %s %s", sets[0], sets[1], sets[2])
	}
}

func TestPrinting(t *testing.T) {
	set := must(t, 1, 3, 5)
	if got := set.String(); got != "1.3.5" {
		t.Fatalf("String = %q, want 1.3.5", got)
	}
	if got := set.Compact(); got != "135" {
		t.Fatalf("Compact = %q, want 135", got)
	}
	if got := set.Grid(6); got != "x.x.x." {
		t.Fatalf("Grid = %q, want x.x.x.", got)
	}
	// A set that reaches past nine cannot be written as one digit per shaft.
	wide := must(t, 1, 10, 12)
	if got := wide.Compact(); got != "1.10.12" {
		t.Fatalf("Compact = %q, want 1.10.12", got)
	}
	if got := wide.Grid(12); got != "x........x.x" {
		t.Fatalf("Grid = %q", got)
	}
	empty, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if got := empty.String(); got != "-" {
		t.Fatalf("the empty set prints as %q", got)
	}
	if got := empty.Compact(); got != "-" {
		t.Fatalf("the empty set prints as %q", got)
	}
	if got := empty.Grid(4); got != "...." {
		t.Fatalf("the empty set draws as %q", got)
	}
}

func TestValidate(t *testing.T) {
	if err := must(t, 1, 4).Validate(4); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	// A set naming a shaft the loom does not have cannot be lifted.
	if err := must(t, 1, 5).Validate(4); err == nil {
		t.Fatalf("a shaft off the loom must be reported")
	}
	// A treadle that lifts nothing opens no shed.
	empty, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if err := empty.Validate(4); err == nil {
		t.Fatalf("the empty set must be reported")
	}
	for _, loom := range []int{0, 1, 33} {
		if err := must(t, 1).Validate(loom); err == nil {
			t.Fatalf("Validate on a loom of %d = nil error, want a failure", loom)
		}
	}
}

func TestParse(t *testing.T) {
	// A run of digits is one shaft each.
	set, err := Parse("135", 6)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !set.Equal(must(t, 1, 3, 5)) {
		t.Fatalf("Parse(135) = %s", set)
	}
	// Full stops separate the shafts when any of them needs two digits.
	wide, err := Parse("1.10.12", 12)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !wide.Equal(must(t, 1, 10, 12)) {
		t.Fatalf("Parse(1.10.12) = %s", wide)
	}
	// The order makes no difference and space around it is trimmed.
	spaced, err := Parse("  531 ", 6)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !spaced.Equal(set) {
		t.Fatalf("Parse of a spaced set = %s", spaced)
	}
	// One shaft is a set.
	single, err := Parse("4", 6)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !single.Equal(must(t, 4)) {
		t.Fatalf("Parse(4) = %s", single)
	}
	// Whatever comes back can be written out and read again.
	for _, text := range []string{"1", "12", "135", "1.10.12"} {
		loom := 12
		first, err := Parse(text, loom)
		if err != nil {
			t.Fatalf("Parse(%q): %v", text, err)
		}
		second, err := Parse(first.Compact(), loom)
		if err != nil {
			t.Fatalf("Parse(%q): %v", first.Compact(), err)
		}
		if !second.Equal(first) {
			t.Fatalf("%q read back as %s", text, second)
		}
	}
}

func TestParseRejectsWhatIsNotASet(t *testing.T) {
	for label, item := range map[string]struct {
		text string
		loom int
	}{
		"nothing":                       {"", 6},
		"only space":                    {"   ", 6},
		"a dash":                        {"-", 6},
		"a letter":                      {"1a", 6},
		"a shaft of zero":               {"10", 6},
		"an empty shaft":                {"1..3", 12},
		"a shaft named twice":           {"11", 6},
		"a shaft named twice by number": {"1.10.1", 12},
		"a shaft off the loom":          {"18", 6},
		"a shaft off a wide loom":       {"1.13", 12},
		"a loom of one":                 {"1", 1},
		"a loom too wide":               {"1", 33},
	} {
		if _, err := Parse(item.text, item.loom); err == nil {
			t.Fatalf("%s: Parse(%q, %d) = nil error, want a failure", label, item.text, item.loom)
		}
	}
}

func TestParseGrid(t *testing.T) {
	set, err := ParseGrid("x.x.", 4)
	if err != nil {
		t.Fatalf("ParseGrid: %v", err)
	}
	if !set.Equal(must(t, 1, 3)) {
		t.Fatalf("ParseGrid = %s", set)
	}
	// Anything other than a dot, a space, a zero or a dash lifts the shaft.
	mixed, err := ParseGrid("1001", 4)
	if err != nil {
		t.Fatalf("ParseGrid: %v", err)
	}
	if !mixed.Equal(must(t, 1, 4)) {
		t.Fatalf("ParseGrid = %s", mixed)
	}
	// A grid and a compact set describe the same thing.
	if got := set.Grid(4); got != "x.x." {
		t.Fatalf("Grid = %q", got)
	}
	for label, item := range map[string]struct {
		text string
		loom int
	}{
		"too short":       {"x.", 4},
		"too long":        {"x.x.x.", 4},
		"nothing lifted":  {"....", 4},
		"a loom of one":   {"x", 1},
		"a loom too wide": {strings.Repeat("x", 33), 33},
	} {
		if _, err := ParseGrid(item.text, item.loom); err == nil {
			t.Fatalf("%s: ParseGrid(%q, %d) = nil error, want a failure", label, item.text, item.loom)
		}
	}
}

func TestDescribe(t *testing.T) {
	got := must(t, 1, 3).Describe(4)
	for _, fragment := range []string{"13", "2 of 4 shaft(s) lifted", "x.x."} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("Describe = %q, which is missing %q", got, fragment)
		}
	}
}
