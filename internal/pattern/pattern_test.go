package pattern

import (
	"strings"
	"testing"

	"Heddle/internal/cloth"
	"Heddle/internal/draft"
)

// mustWeave weaves a catalogue draft or fails the test.
func mustWeave(t *testing.T, key string) cloth.Cloth {
	t.Helper()
	built, err := draft.Lookup(key)
	if err != nil {
		t.Fatalf("draft.Lookup(%s): %v", key, err)
	}
	out, err := cloth.Weave(built)
	if err != nil {
		t.Fatalf("cloth.Weave(%s): %v", key, err)
	}
	return out
}

// fromRows builds a cloth from drawn rows, x for the warp raised.
func fromRows(t *testing.T, rows ...string) cloth.Cloth {
	t.Helper()
	cells := []bool{}
	for _, row := range rows {
		for index := 0; index < len(row); index++ {
			cells = append(cells, row[index] == 'x')
		}
	}
	out, err := cloth.New(len(rows[0]), len(rows), cells)
	if err != nil {
		t.Fatalf("cloth.New: %v", err)
	}
	return out
}

func TestIsPlainWeave(t *testing.T) {
	for _, key := range []string{"plain-weave", "plain-weave-4"} {
		got, err := IsPlainWeave(mustWeave(t, key))
		if err != nil {
			t.Fatalf("%s: IsPlainWeave: %v", key, err)
		}
		if !got {
			t.Fatalf("%s is plain weave", key)
		}
	}
	for _, key := range []string{"twill-2-2", "twill-1-3", "basket-2-2", "satin-5", "floating-end"} {
		got, err := IsPlainWeave(mustWeave(t, key))
		if err != nil {
			t.Fatalf("%s: IsPlainWeave: %v", key, err)
		}
		if got {
			t.Fatalf("%s is not plain weave", key)
		}
	}
	// A cloth with only one end or one pick is not plain weave, because it never changes in
	// that direction.
	single, err := IsPlainWeave(fromRows(t, "x."))
	if err != nil {
		t.Fatalf("IsPlainWeave: %v", err)
	}
	if single {
		t.Fatalf("one pick is not plain weave")
	}
	if _, err := IsPlainWeave(cloth.Cloth{}); err == nil {
		t.Fatalf("an empty cloth must be reported")
	}
}

func TestAsTwill(t *testing.T) {
	for _, item := range []struct {
		key       string
		up, down  int
		step      int
		repeat    int
		direction string
	}{
		{"twill-1-3", 1, 3, 1, 4, "right"},
		{"twill-2-2", 2, 2, 1, 4, "right"},
		{"twill-3-1", 3, 1, 1, 4, "right"},
		{"twill-3-1-repeat", 3, 1, 1, 4, "right"},
		// Plain weave is a 1/1 twill by this reading, which is why the classifier looks for
		// plain weave first.
		{"plain-weave", 1, 1, 1, 2, "right"},
	} {
		got, ok, err := AsTwill(mustWeave(t, item.key))
		if err != nil {
			t.Fatalf("%s: AsTwill: %v", item.key, err)
		}
		if !ok {
			t.Fatalf("%s is a twill", item.key)
		}
		if got.Up != item.up || got.Down != item.down {
			t.Fatalf("%s reads as %d/%d, want %d/%d", item.key, got.Up, got.Down, item.up, item.down)
		}
		if got.Step != item.step || got.Repeat != item.repeat {
			t.Fatalf("%s steps %d over %d, want %d over %d",
				item.key, got.Step, got.Repeat, item.step, item.repeat)
		}
		if got.Direction != item.direction {
			t.Fatalf("%s climbs to the %s", item.key, got.Direction)
		}
		// The counts are over the repeat, so they add up to it.
		if got.Up+got.Down != got.Repeat {
			t.Fatalf("%s reads as %d/%d over a repeat of %d", item.key, got.Up, got.Down, got.Repeat)
		}
	}
	// A basket is not a twill: the pick that follows is the same pick, not a shifted one.
	if _, ok, err := AsTwill(mustWeave(t, "basket-2-2")); err != nil || ok {
		t.Fatalf("AsTwill on a basket = %v, %v", ok, err)
	}
	// A cloth of two ends is too narrow to carry a diagonal.
	if _, ok, err := AsTwill(fromRows(t, "x.", ".x")); err != nil || ok {
		t.Fatalf("AsTwill on a cloth of two ends = %v, %v", ok, err)
	}
	// A cloth in which every cell is the same is not a twill.
	if _, ok, err := AsTwill(fromRows(t, "xxxx", "xxxx")); err != nil || ok {
		t.Fatalf("AsTwill on a solid cloth = %v, %v", ok, err)
	}
	if _, _, err := AsTwill(cloth.Cloth{}); err == nil {
		t.Fatalf("an empty cloth must be reported")
	}
}

func TestATwillClimbingTheOtherWay(t *testing.T) {
	// The same 2/2 twill with the picks read the other way round climbs to the left.
	fabric := fromRows(t, "xx..", "x..x", "..xx", ".xx.")
	got, ok, err := AsTwill(fabric)
	if err != nil {
		t.Fatalf("AsTwill: %v", err)
	}
	if !ok {
		t.Fatalf("the cloth is a twill")
	}
	if got.Direction != "left" {
		t.Fatalf("the twill climbs to the %s", got.Direction)
	}
	if got.Shift != 1 {
		t.Fatalf("the twill slopes %d, want 1", got.Shift)
	}
	if got.Step != 3 {
		t.Fatalf("the twill steps %d to the right, want 3", got.Step)
	}
}

func TestSatinCountersOnAPrimeNumberOfShafts(t *testing.T) {
	// On a prime number of shafts no step can share a factor with the shafts, so every step
	// except the two at the ends can be used.
	for loom, want := range map[int][]int{
		5:  {2, 3},
		7:  {2, 3, 4, 5},
		11: {2, 3, 4, 5, 6, 7, 8, 9},
		13: {2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	} {
		got, err := SatinCounters(loom)
		if err != nil {
			t.Fatalf("SatinCounters(%d): %v", loom, err)
		}
		if len(got) != len(want) {
			t.Fatalf("SatinCounters(%d) = %v, want %v", loom, got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("SatinCounters(%d) = %v, want %v", loom, got, want)
			}
		}
		// Every step it offers is one the rule accepts, and it comes back in order.
		for index, counter := range got {
			if err := IsValidSatinCounter(loom, counter); err != nil {
				t.Fatalf("SatinCounters(%d) offered %d: %v", loom, counter, err)
			}
			if index > 0 && counter <= got[index-1] {
				t.Fatalf("SatinCounters(%d) came back out of order: %v", loom, got)
			}
		}
	}
	for _, loom := range []int{2, 3, 4, 33, 0, -5} {
		if _, err := SatinCounters(loom); err == nil {
			t.Fatalf("SatinCounters(%d) = nil error, want a failure", loom)
		}
	}
}

func TestIsValidSatinCounterRejectsTheStepsAtTheEnds(t *testing.T) {
	// A step of one, and a step of one less than the shafts, put every interlacement beside
	// the one before it, which is a twill.
	for _, loom := range []int{5, 7, 11, 13} {
		if err := IsValidSatinCounter(loom, 1); err == nil {
			t.Fatalf("a step of 1 on %d shafts must be reported", loom)
		}
		if err := IsValidSatinCounter(loom, loom-1); err == nil {
			t.Fatalf("a step of %d on %d shafts must be reported", loom-1, loom)
		}
		// A step outside the repeat is not a step at all.
		for _, counter := range []int{0, -1, loom, loom + 1} {
			if err := IsValidSatinCounter(loom, counter); err == nil {
				t.Fatalf("a step of %d on %d shafts must be reported", counter, loom)
			}
		}
		// The steps in between are all usable on a prime number of shafts.
		for counter := 2; counter < loom-1; counter++ {
			if err := IsValidSatinCounter(loom, counter); err != nil {
				t.Fatalf("a step of %d on %d shafts: %v", counter, loom, err)
			}
		}
	}
	for _, loom := range []int{2, 3, 4, 33} {
		if err := IsValidSatinCounter(loom, 2); err == nil {
			t.Fatalf("a satin on %d shafts must be reported", loom)
		}
	}
}

func TestSatinTieUpLiftsEveryShaftOnce(t *testing.T) {
	// A satin tie-up steps one interlacement round the shafts, so over the whole tie-up each
	// shaft is lifted exactly once.
	for _, loom := range []int{5, 7, 11} {
		counters, err := SatinCounters(loom)
		if err != nil {
			t.Fatalf("SatinCounters: %v", err)
		}
		for _, counter := range counters {
			tieup, err := SatinTieUp(loom, counter, false)
			if err != nil {
				t.Fatalf("SatinTieUp(%d, %d): %v", loom, counter, err)
			}
			if tieup.Treadles() != loom {
				t.Fatalf("the tie-up covers %d treadle(s) for %d shaft(s)", tieup.Treadles(), loom)
			}
			counts := make([]int, loom+1)
			for _, set := range tieup {
				if set.Count() != 1 {
					t.Fatalf("a treadle of a %d shaft sateen lifts %d shaft(s)", loom, set.Count())
				}
				counts[set.Highest()]++
			}
			for shaft := 1; shaft <= loom; shaft++ {
				if counts[shaft] != 1 {
					t.Fatalf("shaft %d is lifted %d time(s) by a %d shaft sateen stepping %d",
						shaft, counts[shaft], loom, counter)
				}
			}
			// The warp faced tie-up is the same tie-up turned over.
			warp, err := SatinTieUp(loom, counter, true)
			if err != nil {
				t.Fatalf("SatinTieUp: %v", err)
			}
			for index := range warp {
				if warp[index].Count() != loom-1 {
					t.Fatalf("a treadle of a %d shaft satin lifts %d shaft(s)",
						loom, warp[index].Count())
				}
				if !warp[index].Intersect(tieup[index]).IsEmpty() {
					t.Fatalf("the two faces of the tie-up share a shaft on treadle %d", index+1)
				}
			}
		}
	}
	// A step the rule refuses gives no tie-up.
	for _, counter := range []int{1, 4} {
		if _, err := SatinTieUp(5, counter, false); err == nil {
			t.Fatalf("a step of %d on 5 shafts must be reported", counter)
		}
	}
}

func TestSatinDraft(t *testing.T) {
	built, err := SatinDraft(5, 2, false, 2)
	if err != nil {
		t.Fatalf("SatinDraft: %v", err)
	}
	if built.Shafts != 5 || built.Treadles != 5 {
		t.Fatalf("SatinDraft = %+v", built)
	}
	if built.Ends() != 10 || built.Picks() != 10 {
		t.Fatalf("the draft is %d end(s) by %d pick(s)", built.Ends(), built.Picks())
	}
	if err := built.Validate(); err != nil {
		t.Fatalf("the satin draft is not a draft: %v", err)
	}
	// Every shaft carries ends, so no part of the warp is left out.
	if empty := built.Threading.EmptyShafts(built.Shafts); len(empty) != 0 {
		t.Fatalf("the satin leaves shaft(s) %v with no end on them", empty)
	}
	fabric, err := cloth.Weave(built)
	if err != nil {
		t.Fatalf("Weave: %v", err)
	}
	// A five shaft sateen raises one end in five, so it is weft faced.
	if fabric.WarpUp() != fabric.Cells()/5 {
		t.Fatalf("the sateen raises %d of %d cell(s)", fabric.WarpUp(), fabric.Cells())
	}
	if got := fabric.Face(); got != "weft faced" {
		t.Fatalf("the sateen is %q", got)
	}
	// The warp faced satin is the other way round.
	warp, err := SatinDraft(5, 2, true, 2)
	if err != nil {
		t.Fatalf("SatinDraft: %v", err)
	}
	warpCloth, err := cloth.Weave(warp)
	if err != nil {
		t.Fatalf("Weave: %v", err)
	}
	if warpCloth.WarpUp() != warpCloth.Cells()-fabric.WarpUp() {
		t.Fatalf("the satin raises %d of %d cell(s)", warpCloth.WarpUp(), warpCloth.Cells())
	}
	for label, item := range map[string]struct {
		loom, counter, repeats int
	}{
		"no repeats":     {5, 2, 0},
		"too many":       {5, 2, 33},
		"a bad step":     {5, 1, 2},
		"too few shafts": {4, 2, 2},
	} {
		if _, err := SatinDraft(item.loom, item.counter, false, item.repeats); err == nil {
			t.Fatalf("%s: SatinDraft = nil error, want a failure", label)
		}
	}
}

func TestAsSatin(t *testing.T) {
	satin, ok, err := AsSatin(mustWeave(t, "satin-5"))
	if err != nil {
		t.Fatalf("AsSatin: %v", err)
	}
	if !ok {
		t.Fatalf("the satin reads as a satin")
	}
	if satin.Shafts != 5 || satin.Counter != 2 || satin.Up != 4 {
		t.Fatalf("AsSatin = %+v", satin)
	}
	sateen, ok, err := AsSatin(mustWeave(t, "sateen-5"))
	if err != nil {
		t.Fatalf("AsSatin: %v", err)
	}
	if !ok {
		t.Fatalf("the sateen reads as a satin")
	}
	if sateen.Shafts != 5 || sateen.Counter != 2 || sateen.Up != 1 {
		t.Fatalf("AsSatin = %+v", sateen)
	}
	// A twill that steps by one is not a satin, however many shafts it is on.
	for _, key := range []string{"plain-weave", "twill-2-2", "twill-1-3", "twill-3-1", "basket-2-2"} {
		if _, ok, err := AsSatin(mustWeave(t, key)); err != nil || ok {
			t.Fatalf("%s reads as a satin", key)
		}
	}
	if got := satin.Describe(); !strings.Contains(got, "warp faced") ||
		!strings.Contains(got, "stepping 2") {
		t.Fatalf("Describe = %q", got)
	}
	if got := sateen.Describe(); !strings.Contains(got, "weft faced") {
		t.Fatalf("Describe = %q", got)
	}
}

func TestRepeat(t *testing.T) {
	// A cloth of one repeat repeats over itself.
	ends, picks, err := Repeat(fromRows(t, "x.", ".x"))
	if err != nil {
		t.Fatalf("Repeat: %v", err)
	}
	if ends != 2 || picks != 2 {
		t.Fatalf("Repeat = %d by %d, want 2 by 2", ends, picks)
	}
	// The same cloth woven twice over repeats over the same block.
	ends, picks, err = Repeat(fromRows(t, "x.x.", ".x.x", "x.x.", ".x.x"))
	if err != nil {
		t.Fatalf("Repeat: %v", err)
	}
	if ends != 2 || picks != 2 {
		t.Fatalf("Repeat = %d by %d, want 2 by 2", ends, picks)
	}
	// A cloth that does not repeat within itself repeats over the whole of itself.
	ends, picks, err = Repeat(fromRows(t, "xx.", "x..", "..x"))
	if err != nil {
		t.Fatalf("Repeat: %v", err)
	}
	if ends != 3 || picks != 3 {
		t.Fatalf("Repeat = %d by %d, want 3 by 3", ends, picks)
	}
	// The repeat of every draft in the catalogue divides its size, which is the check that the
	// repeat was found rather than guessed.
	names, err := draft.Names()
	if err != nil {
		t.Fatalf("draft.Names: %v", err)
	}
	for _, key := range names {
		fabric := mustWeave(t, key)
		ends, picks, err := Repeat(fabric)
		if err != nil {
			t.Fatalf("%s: Repeat: %v", key, err)
		}
		if fabric.Ends%ends != 0 || fabric.Picks%picks != 0 {
			t.Fatalf("%s is %d by %d and repeats over %d by %d",
				key, fabric.Ends, fabric.Picks, ends, picks)
		}
	}
	if _, _, err := Repeat(cloth.Cloth{}); err == nil {
		t.Fatalf("an empty cloth must be reported")
	}
}

func TestThreadingSymmetric(t *testing.T) {
	if !ThreadingSymmetric(draft.Threading{1, 2, 3, 3, 2, 1}) {
		t.Fatalf("a threading that reads the same either way is symmetric")
	}
	if !ThreadingSymmetric(draft.Threading{2}) {
		t.Fatalf("one end is symmetric")
	}
	if !ThreadingSymmetric(draft.Threading{1, 2, 1}) {
		t.Fatalf("a threading that reads the same either way is symmetric")
	}
	if ThreadingSymmetric(draft.Threading{1, 2, 3, 4}) {
		t.Fatalf("a straight draw is not symmetric")
	}
	if ThreadingSymmetric(draft.Threading{1, 2, 3, 4, 1, 2, 3, 4}) {
		t.Fatalf("a straight draw is not symmetric")
	}
	if ThreadingSymmetric(draft.Threading{1, 2, 3, 4, 3, 2}) {
		t.Fatalf("a point draw over four shafts is not symmetric")
	}
}

func TestClassify(t *testing.T) {
	for key, want := range map[string]string{
		"plain-weave":   "plain weave",
		"plain-weave-4": "plain weave",
		"twill-2-2":     "2/2 twill stepping 1 to the right",
		"twill-1-3":     "1/3 twill stepping 1 to the right",
		"twill-3-1":     "3/1 twill stepping 1 to the right",
		"satin-5":       "5 shaft warp faced satin stepping 2",
		"sateen-5":      "5 shaft weft faced satin stepping 2",
		"basket-2-2":    "no plain weave, twill or satin",
	} {
		got, err := Classify(mustWeave(t, key))
		if err != nil {
			t.Fatalf("%s: Classify: %v", key, err)
		}
		if got != want {
			t.Fatalf("%s reads as %q, want %q", key, got, want)
		}
	}
	if _, err := Classify(cloth.Cloth{}); err == nil {
		t.Fatalf("an empty cloth must be reported")
	}
}

func TestAnalyse(t *testing.T) {
	analysis, err := Analyse(mustWeave(t, "twill-2-2"))
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if analysis.Ends != 8 || analysis.Picks != 8 {
		t.Fatalf("Analyse = %+v", analysis)
	}
	if analysis.RepeatEnds != 4 || analysis.RepeatPicks != 4 {
		t.Fatalf("the repeat is %d by %d, want 4 by 4", analysis.RepeatEnds, analysis.RepeatPicks)
	}
	if analysis.PlainWeave {
		t.Fatalf("a twill is not plain weave")
	}
	if !analysis.IsTwill || analysis.IsSatin {
		t.Fatalf("Analyse = %+v", analysis)
	}
	if analysis.Twill.Up != 2 || analysis.Twill.Down != 2 {
		t.Fatalf("the twill reads as %d/%d", analysis.Twill.Up, analysis.Twill.Down)
	}
	if analysis.Balance != 0.5 || analysis.Face != "balanced" {
		t.Fatalf("Analyse = %+v", analysis)
	}
	if analysis.Structure != "2/2 twill stepping 1 to the right" {
		t.Fatalf("Structure = %q", analysis.Structure)
	}
	got := analysis.Describe()
	for _, fragment := range []string{"8 end(s) by 8 pick(s)", "4 by 4", "2/2 twill", "balanced"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("Describe = %q, which is missing %q", got, fragment)
		}
	}
	if _, err := Analyse(cloth.Cloth{}); err == nil {
		t.Fatalf("an empty cloth must be reported")
	}
}

func TestCounterList(t *testing.T) {
	if got := CounterList(nil); got != "none" {
		t.Fatalf("CounterList of nothing = %q, want none", got)
	}
	if got := CounterList([]int{3, 2, 5}); got != "2 3 5" {
		t.Fatalf("CounterList = %q, want 2 3 5", got)
	}
	// The list it is given is left alone.
	counters := []int{5, 2}
	CounterList(counters)
	if counters[0] != 5 {
		t.Fatalf("CounterList sorted the list it was given: %v", counters)
	}
}
