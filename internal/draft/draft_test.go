package draft

import (
	"strings"
	"testing"

	"Heddle/internal/shafts"
)

// twill returns the 2/2 twill the rest of this file works on.
func twill(t *testing.T) Draft {
	t.Helper()
	built, err := Lookup("twill-2-2")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	return built
}

// mustSet builds a shaft set or fails the test.
func mustSet(t *testing.T, list ...int) shafts.Set {
	t.Helper()
	out, err := shafts.New(list...)
	if err != nil {
		t.Fatalf("shafts.New(%v): %v", list, err)
	}
	return out
}

// same fails the test unless two sequences of numbers agree.
func same[T ~int](t *testing.T, got []T, want []int, label string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for index := range want {
		if int(got[index]) != want[index] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}

func TestStraight(t *testing.T) {
	got, err := Straight(4, 10)
	if err != nil {
		t.Fatalf("Straight: %v", err)
	}
	same(t, got, []int{1, 2, 3, 4, 1, 2, 3, 4, 1, 2}, "Straight(4, 10)")
	if got.Ends() != 10 {
		t.Fatalf("Ends = %d, want 10", got.Ends())
	}
	if got.Repeat() != 4 {
		t.Fatalf("Repeat = %d, want 4", got.Repeat())
	}
	for label, item := range map[string]struct{ loom, ends int }{
		"a loom of one":    {1, 8},
		"a loom too wide":  {33, 8},
		"no ends":          {4, 0},
		"too many ends":    {4, 513},
		"a negative count": {4, -2},
	} {
		if _, err := Straight(item.loom, item.ends); err == nil {
			t.Fatalf("%s: Straight(%d, %d) = nil error, want a failure", label, item.loom, item.ends)
		}
	}
}

func TestPoint(t *testing.T) {
	// A point draw over four shafts turns at the fourth shaft and comes back down without
	// repeating it, so it repeats over six ends.
	got, err := Point(4, 12)
	if err != nil {
		t.Fatalf("Point: %v", err)
	}
	same(t, got, []int{1, 2, 3, 4, 3, 2, 1, 2, 3, 4, 3, 2}, "Point(4, 12)")
	if got.Repeat() != 6 {
		t.Fatalf("Repeat = %d, want 6", got.Repeat())
	}
	if err := got.Validate(4); err != nil {
		t.Fatalf("a point draw is a threading: %v", err)
	}
	// Over two shafts a point draw is the same as a straight draw.
	two, err := Point(2, 6)
	if err != nil {
		t.Fatalf("Point: %v", err)
	}
	same(t, two, []int{1, 2, 1, 2, 1, 2}, "Point(2, 6)")
	if _, err := Point(1, 8); err == nil {
		t.Fatalf("a loom of one must be reported")
	}
	if _, err := Point(4, 0); err == nil {
		t.Fatalf("no ends must be reported")
	}
}

func TestThreadingValidate(t *testing.T) {
	if err := (Threading{1, 2, 3, 4}).Validate(4); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	for label, item := range map[string]struct {
		threading Threading
		loom      int
	}{
		"a shaft off the loom": {Threading{1, 2, 3, 5}, 4},
		"a shaft of zero":      {Threading{1, 0, 3}, 4},
		"a negative shaft":     {Threading{1, -2}, 4},
		"no ends":              {Threading{}, 4},
		"a loom of one":        {Threading{1}, 1},
		"a loom too wide":      {Threading{1}, 33},
	} {
		if err := item.threading.Validate(item.loom); err == nil {
			t.Fatalf("%s: Validate = nil error, want a failure", label)
		}
	}
}

func TestThreadingUsage(t *testing.T) {
	threading := Threading{1, 2, 3, 4, 1, 2, 1, 1}
	same(t, threading.Usage(4), []int{4, 2, 1, 1}, "Usage")
	same(t, threading.EndsOnShaft(1), []int{1, 5, 7, 8}, "EndsOnShaft(1)")
	same(t, threading.EndsOnShaft(3), []int{3}, "EndsOnShaft(3)")
	if got := threading.EndsOnShaft(9); len(got) != 0 {
		t.Fatalf("EndsOnShaft(9) = %v", got)
	}
	// A shaft with no end on it does nothing, and the threading says which those are.
	if got := threading.EmptyShafts(6); len(got) != 2 || got[0] != 5 || got[1] != 6 {
		t.Fatalf("EmptyShafts = %v, want 5 and 6", got)
	}
	if got := threading.EmptyShafts(4); len(got) != 0 {
		t.Fatalf("EmptyShafts = %v, want none", got)
	}
	// The usage adds up to the ends.
	total := 0
	for _, count := range threading.Usage(4) {
		total += count
	}
	if total != threading.Ends() {
		t.Fatalf("the usage adds up to %d and there are %d end(s)", total, threading.Ends())
	}
}

func TestThreadingCloneAndReverse(t *testing.T) {
	threading := Threading{1, 2, 3, 4, 1, 2, 3, 4}
	copied := threading.Clone()
	copied[0] = 4
	same(t, threading, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the original after writing to the clone")

	reversed := threading.Reverse()
	same(t, reversed, []int{4, 3, 2, 1, 4, 3, 2, 1}, "Reverse")
	// The threading it was read from is left alone.
	same(t, threading, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the original after Reverse")
	// Reversing twice gives the threading back.
	same(t, reversed.Reverse(), []int{1, 2, 3, 4, 1, 2, 3, 4}, "Reverse twice")
	// A threading that reads the same either way is its own reverse.
	symmetric := Threading{1, 2, 3, 3, 2, 1}
	same(t, symmetric.Reverse(), []int{1, 2, 3, 3, 2, 1}, "Reverse of a symmetric threading")
}

func TestRepeat(t *testing.T) {
	for _, item := range []struct {
		threading Threading
		want      int
	}{
		{Threading{1, 2, 1, 2, 1, 2}, 2},
		{Threading{1, 2, 3, 4, 1, 2, 3, 4}, 4},
		{Threading{1, 1, 2, 2, 1, 1, 2, 2}, 4},
		{Threading{1, 2, 3, 4, 3, 2, 1, 2, 3, 4, 3, 2}, 6},
		{Threading{1, 2, 3}, 3},
		{Threading{1, 1, 1, 1}, 1},
		// A threading that does not repeat is its own repeat.
		{Threading{1, 2, 4, 3}, 4},
	} {
		if got := item.threading.Repeat(); got != item.want {
			t.Fatalf("the repeat of %v = %d, want %d", item.threading, got, item.want)
		}
	}
	// A treadling is the same kind of sequence read down the cloth.
	if got := (Treadling{1, 2, 1, 2}).Repeat(); got != 2 {
		t.Fatalf("the repeat of a treadling = %d, want 2", got)
	}
}

func TestTreadling(t *testing.T) {
	treadling := Treadling{1, 2, 3, 4, 1, 2}
	if treadling.Picks() != 6 {
		t.Fatalf("Picks = %d, want 6", treadling.Picks())
	}
	same(t, treadling.Usage(4), []int{2, 2, 1, 1}, "Usage")
	same(t, treadling.Reverse(), []int{2, 1, 4, 3, 2, 1}, "Reverse")
	same(t, treadling, []int{1, 2, 3, 4, 1, 2}, "the original after Reverse")
	copied := treadling.Clone()
	copied[0] = 3
	same(t, treadling, []int{1, 2, 3, 4, 1, 2}, "the original after writing to the clone")
	if got := treadling.String(); got != "1 2 3 4 1 2" {
		t.Fatalf("String = %q", got)
	}
	for label, item := range map[string]struct {
		treadling Treadling
		treadles  int
	}{
		"a treadle that is not there": {Treadling{1, 5}, 4},
		"a treadle of zero":           {Treadling{0}, 4},
		"no picks":                    {Treadling{}, 4},
		"no treadles":                 {Treadling{1}, 0},
		"too many treadles":           {Treadling{1}, 33},
	} {
		if err := item.treadling.Validate(item.treadles); err == nil {
			t.Fatalf("%s: Validate = nil error, want a failure", label)
		}
	}
}

func TestTieUp(t *testing.T) {
	tieup := TieUp{mustSet(t, 1, 2), mustSet(t, 2, 3), mustSet(t, 3, 4), mustSet(t, 1, 4)}
	if err := tieup.Validate(4, 4); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if tieup.Treadles() != 4 {
		t.Fatalf("Treadles = %d, want 4", tieup.Treadles())
	}
	if got := tieup.String(); got != "12 23 34 14" {
		t.Fatalf("String = %q", got)
	}
	// Read the other way round, the tie-up says which treadles lift each shaft.
	transposed := tieup.Transpose(4)
	if len(transposed) != 4 {
		t.Fatalf("Transpose gave %d shaft(s)", len(transposed))
	}
	if got := transposed[0].Compact(); got != "14" {
		t.Fatalf("shaft 1 is lifted by %q, want 14", got)
	}
	if got := transposed[2].Compact(); got != "23" {
		t.Fatalf("shaft 3 is lifted by %q, want 23", got)
	}
	// Every shaft of this tie-up can be lifted by two treadles, and the counts add up.
	total := 0
	for _, set := range transposed {
		total += set.Count()
	}
	lifted := 0
	for _, set := range tieup {
		lifted += set.Count()
	}
	if total != lifted {
		t.Fatalf("the tie-up lifts %d shaft(s) and the transpose holds %d", lifted, total)
	}
	copied := tieup.Clone()
	copied[0] = mustSet(t, 1)
	if got := tieup[0].Compact(); got != "12" {
		t.Fatalf("writing to the clone changed the tie-up to %q", got)
	}
}

func TestTieUpValidateRejectsATreadleThatOpensNoShed(t *testing.T) {
	// A treadle that lifts every shaft leaves no gap for the shuttle.
	all := TieUp{mustSet(t, 1, 2, 3, 4), mustSet(t, 1, 2)}
	if err := all.Validate(4, 2); err == nil {
		t.Fatalf("a treadle that lifts every shaft must be reported")
	}
	// So does one that lifts none.
	none := TieUp{shafts.Set(0), mustSet(t, 1, 2)}
	if err := none.Validate(4, 2); err == nil {
		t.Fatalf("a treadle that lifts nothing must be reported")
	}
	// A shaft off the loom cannot be tied to.
	off := TieUp{mustSet(t, 1, 5)}
	if err := off.Validate(4, 1); err == nil {
		t.Fatalf("a shaft off the loom must be reported")
	}
	// The tie-up has to cover exactly the treadles the draft has.
	short := TieUp{mustSet(t, 1)}
	if err := short.Validate(4, 2); err == nil {
		t.Fatalf("a tie-up that is short of a treadle must be reported")
	}
	for _, loom := range []int{1, 33} {
		if err := (TieUp{mustSet(t, 1)}).Validate(loom, 1); err == nil {
			t.Fatalf("Validate on a loom of %d = nil error, want a failure", loom)
		}
	}
}

func TestDraftValidate(t *testing.T) {
	built := twill(t)
	if err := built.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if built.Ends() != 8 || built.Picks() != 8 {
		t.Fatalf("the draft is %d end(s) by %d pick(s)", built.Ends(), built.Picks())
	}
	unnamed := built.Clone()
	unnamed.Name = "  "
	if err := unnamed.Validate(); err == nil {
		t.Fatalf("a draft with no name must be reported")
	}
	badThreading := built.Clone()
	badThreading.Threading[0] = 9
	if err := badThreading.Validate(); err == nil {
		t.Fatalf("a threading off the loom must be reported")
	}
	badTreadling := built.Clone()
	badTreadling.Treadling[0] = 9
	if err := badTreadling.Validate(); err == nil {
		t.Fatalf("a treadling off the treadles must be reported")
	}
	badTieUp := built.Clone()
	badTieUp.TieUp = badTieUp.TieUp[:2]
	if err := badTieUp.Validate(); err == nil {
		t.Fatalf("a short tie-up must be reported")
	}
}

func TestCloneIsIndependent(t *testing.T) {
	built := twill(t)
	copied := built.Clone()
	copied.Threading[0] = 4
	copied.Treadling[0] = 4
	copied.TieUp[0] = mustSet(t, 1)
	if built.Threading[0] != 1 || built.Treadling[0] != 1 {
		t.Fatalf("writing to the clone changed the draft: %v %v", built.Threading, built.Treadling)
	}
	if got := built.TieUp[0].Compact(); got != "12" {
		t.Fatalf("writing to the clone changed the tie-up to %q", got)
	}
}

func TestLiftedAndWarpUp(t *testing.T) {
	built := twill(t)
	lifted, err := built.Lifted(1)
	if err != nil {
		t.Fatalf("Lifted: %v", err)
	}
	if got := lifted.Compact(); got != "12" {
		t.Fatalf("the first pick lifts %q, want 12", got)
	}
	// End 1 is on shaft 1 and the first pick lifts shaft 1, so it is raised.
	up, err := built.WarpUp(1, 1)
	if err != nil {
		t.Fatalf("WarpUp: %v", err)
	}
	if !up {
		t.Fatalf("end 1 is raised over pick 1")
	}
	// End 3 is on shaft 3, which the first pick leaves down.
	down, err := built.WarpUp(3, 1)
	if err != nil {
		t.Fatalf("WarpUp: %v", err)
	}
	if down {
		t.Fatalf("end 3 is not raised over pick 1")
	}
	for label, item := range map[string]struct{ end, pick int }{
		"an end that is not there": {9, 1},
		"an end of zero":           {0, 1},
		"a pick that is not there": {1, 9},
		"a pick of zero":           {1, 0},
	} {
		if _, err := built.WarpUp(item.end, item.pick); err == nil {
			t.Fatalf("%s: WarpUp(%d, %d) = nil error, want a failure", label, item.end, item.pick)
		}
	}
	if _, err := built.Lifted(0); err == nil {
		t.Fatalf("a pick of zero must be reported")
	}
}

func TestTrompAsWrit(t *testing.T) {
	built := twill(t)
	tromped, err := built.TrompAsWrit()
	if err != nil {
		t.Fatalf("TrompAsWrit: %v", err)
	}
	// The treadling is read straight off the threading.
	same(t, tromped.Treadling, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the tromped treadling")
	if tromped.Picks() != built.Ends() {
		t.Fatalf("the tromped draft is %d pick(s) for %d end(s)", tromped.Picks(), built.Ends())
	}
	if !strings.Contains(tromped.Name, "tromped") {
		t.Fatalf("the tromped draft is called %q", tromped.Name)
	}
	// The draft it came from is untouched.
	same(t, built.Treadling, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the original treadling")
	// A draft with fewer treadles than shafts cannot be treadled as it is threaded.
	plain, err := Lookup("plain-weave-4")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if _, err := plain.TrompAsWrit(); err == nil {
		t.Fatalf("a draft with four shafts and two treadles cannot be tromped as writ")
	}
}

func TestMirror(t *testing.T) {
	built := twill(t)
	mirrored, err := built.Mirror()
	if err != nil {
		t.Fatalf("Mirror: %v", err)
	}
	same(t, mirrored.Threading, []int{4, 3, 2, 1, 4, 3, 2, 1}, "the mirrored threading")
	// The draft it came from is untouched.
	same(t, built.Threading, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the original threading")
	if !strings.Contains(mirrored.Name, "mirrored") {
		t.Fatalf("the mirrored draft is called %q", mirrored.Name)
	}
	// Mirroring twice gives the threading back.
	back, err := mirrored.Mirror()
	if err != nil {
		t.Fatalf("Mirror: %v", err)
	}
	same(t, back.Threading, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the threading mirrored twice")
}

func TestCatalogue(t *testing.T) {
	names, err := Names()
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if len(names) < 8 {
		t.Fatalf("the catalogue holds %d draft(s)", len(names))
	}
	for index := 1; index < len(names); index++ {
		if names[index] <= names[index-1] {
			t.Fatalf("Names came back out of order: %v", names)
		}
	}
	for _, key := range names {
		built, err := Lookup(key)
		if err != nil {
			t.Fatalf("Lookup(%s): %v", key, err)
		}
		if err := built.Validate(); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		// Every draft in the catalogue can be woven right through.
		for pick := 1; pick <= built.Picks(); pick++ {
			if _, err := built.Lifted(pick); err != nil {
				t.Fatalf("%s pick %d: %v", key, pick, err)
			}
		}
		if len(built.Threading.EmptyShafts(built.Shafts)) > 0 {
			t.Fatalf("%s leaves shaft(s) %v with no end on them",
				key, built.Threading.EmptyShafts(built.Shafts))
		}
	}
	if _, err := Lookup("crackle"); err == nil {
		t.Fatalf("a draft that is not in the catalogue must be reported")
	}
	if _, err := Lookup("  TWILL-2-2 "); err != nil {
		t.Fatalf("Lookup with space and capitals: %v", err)
	}
}

func TestDescribe(t *testing.T) {
	got := twill(t).Describe()
	for _, fragment := range []string{"2/2 Twill", "4 shaft(s)", "4 treadle(s)", "8 end(s)", "12 23 34 14"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("Describe = %q, which is missing %q", got, fragment)
		}
	}
}
