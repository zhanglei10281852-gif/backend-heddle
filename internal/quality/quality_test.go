package quality

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

// mustAssess reads a cloth or fails the test.
func mustAssess(t *testing.T, fabric cloth.Cloth, limit int) Report {
	t.Helper()
	out, err := Assess(fabric, limit)
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	return out
}

func TestAssessPlainWeave(t *testing.T) {
	assessed := mustAssess(t, mustWeave(t, "plain-weave"), DefaultLimit)
	if assessed.Ends != 8 || assessed.Picks != 8 {
		t.Fatalf("Assess = %+v", assessed)
	}
	if assessed.Limit != DefaultLimit {
		t.Fatalf("Limit = %d, want %d", assessed.Limit, DefaultLimit)
	}
	// Plain weave interlaces at every cell, so its floats are all one long and it is as firm
	// as a cloth can be.
	if assessed.LongestWarp != 1 || assessed.LongestWeft != 1 {
		t.Fatalf("the longest floats are %d and %d", assessed.LongestWarp, assessed.LongestWeft)
	}
	if len(assessed.Over) != 0 {
		t.Fatalf("plain weave has %d float(s) over the limit", len(assessed.Over))
	}
	if assessed.Firmness != 1 {
		t.Fatalf("Firmness = %g, want 1", assessed.Firmness)
	}
	if assessed.Sleazy {
		t.Fatalf("plain weave is not sleazy")
	}
	if len(assessed.UnboundEnds) != 0 || len(assessed.UnboundPicks) != 0 {
		t.Fatalf("plain weave leaves %v and %v unbound",
			assessed.UnboundEnds, assessed.UnboundPicks)
	}
	if !assessed.Sound() {
		t.Fatalf("plain weave is sound, got %v", assessed.Problems())
	}
	if got := assessed.Problems(); len(got) != 0 {
		t.Fatalf("Problems = %v", got)
	}
	if assessed.Balance != 0.5 || assessed.Face != "balanced" {
		t.Fatalf("Assess = %+v", assessed)
	}
	if assessed.WarpUp != 32 {
		t.Fatalf("WarpUp = %d, want 32", assessed.WarpUp)
	}
}

func TestAssessATwill(t *testing.T) {
	assessed := mustAssess(t, mustWeave(t, "twill-2-2"), DefaultLimit)
	// A 2/2 twill floats over two in both directions, which the default limit lets pass.
	if assessed.LongestWarp != 2 || assessed.LongestWeft != 2 {
		t.Fatalf("the longest floats are %d and %d", assessed.LongestWarp, assessed.LongestWeft)
	}
	if !assessed.Sound() {
		t.Fatalf("a 2/2 twill is sound at a limit of %d, got %v", DefaultLimit, assessed.Problems())
	}
	// Asked to allow only floats of one, the same cloth has plenty to report.
	strict := mustAssess(t, mustWeave(t, "twill-2-2"), 1)
	if len(strict.Over) == 0 {
		t.Fatalf("a 2/2 twill has floats longer than 1")
	}
	if strict.Sound() {
		t.Fatalf("a 2/2 twill is not sound at a limit of 1")
	}
	if strict.Over[0].Length != 2 {
		t.Fatalf("the worst float is %d long, want 2", strict.Over[0].Length)
	}
	// The floats come back worst first.
	for index := 1; index < len(strict.Over); index++ {
		if strict.Over[index].Length > strict.Over[index-1].Length {
			t.Fatalf("the floats came back out of order at %d", index)
		}
	}
}

func TestAssessFindsAFloatingEnd(t *testing.T) {
	// The tie-up of this draft ties shaft one to every treadle, so the ends on shaft one are
	// raised over every pick and are never woven in.
	assessed := mustAssess(t, mustWeave(t, "floating-end"), DefaultLimit)
	if len(assessed.UnboundEnds) != 2 || assessed.UnboundEnds[0] != 1 || assessed.UnboundEnds[1] != 5 {
		t.Fatalf("UnboundEnds = %v, want 1 and 5", assessed.UnboundEnds)
	}
	if len(assessed.UnboundPicks) != 0 {
		t.Fatalf("UnboundPicks = %v, want none", assessed.UnboundPicks)
	}
	if assessed.Sound() {
		t.Fatalf("a cloth with an end that never interlaces is not sound")
	}
	// The end that never interlaces floats over the whole repeat.
	if assessed.LongestWarp != 8 {
		t.Fatalf("the longest warp float is %d, want 8", assessed.LongestWarp)
	}
	problems := assessed.Problems()
	if len(problems) < 2 {
		t.Fatalf("Problems = %v", problems)
	}
	// The end that will pull out is named first, before the floats.
	if !strings.Contains(problems[0], "never interlace") {
		t.Fatalf("the first problem is %q", problems[0])
	}
	if !strings.Contains(problems[0], "1 5") {
		t.Fatalf("the first problem does not name the ends: %q", problems[0])
	}
}

func TestAssessFindsAnUnboundPick(t *testing.T) {
	// A pick over which every end is raised is not woven in either, and the first and last
	// ends of this cloth never interlace at all.
	assessed := mustAssess(t, fromRows(t, "xxxx", "x..x"), DefaultLimit)
	if len(assessed.UnboundPicks) != 1 || assessed.UnboundPicks[0] != 1 {
		t.Fatalf("UnboundPicks = %v, want 1", assessed.UnboundPicks)
	}
	if len(assessed.UnboundEnds) != 2 || assessed.UnboundEnds[0] != 1 || assessed.UnboundEnds[1] != 4 {
		t.Fatalf("UnboundEnds = %v, want 1 and 4", assessed.UnboundEnds)
	}
	if assessed.Sound() {
		t.Fatalf("the cloth is not sound")
	}
	problems := assessed.Problems()
	found := false
	for _, problem := range problems {
		if strings.Contains(problem, "pick(s) never interlace") {
			found = true
		}
	}
	if !found {
		t.Fatalf("Problems = %v, and none of them names the pick", problems)
	}
}

func TestAssessFindsLongFloatsInASatin(t *testing.T) {
	assessed := mustAssess(t, mustWeave(t, "satin-8"), DefaultLimit)
	// An eight shaft satin floats over seven in both directions, which is what a satin is for
	// and also what makes it a cloth to handle carefully.
	if assessed.LongestWarp != 7 || assessed.LongestWeft != 7 {
		t.Fatalf("the longest floats are %d and %d", assessed.LongestWarp, assessed.LongestWeft)
	}
	if len(assessed.Over) == 0 {
		t.Fatalf("a satin has floats over the default limit")
	}
	if assessed.Over[0].Length != 7 {
		t.Fatalf("the worst float is %d long, want 7", assessed.Over[0].Length)
	}
	// It interlaces once per repeat in each direction, so it is sleazy by this measure.
	if !assessed.Sleazy {
		t.Fatalf("an eight shaft satin is sleazy at a firmness of %g", assessed.Firmness)
	}
	if assessed.Firmness >= SleazyBelow {
		t.Fatalf("Firmness = %g, and the threshold is %g", assessed.Firmness, SleazyBelow)
	}
	if assessed.Sound() {
		t.Fatalf("a satin is not sound by these measures")
	}
	// Given a limit that allows its floats, the only complaint left is the firmness.
	loose := mustAssess(t, mustWeave(t, "satin-8"), 7)
	if len(loose.Over) != 0 {
		t.Fatalf("a limit of 7 lets a float of 7 pass, got %d", len(loose.Over))
	}
	if !loose.Sleazy {
		t.Fatalf("the firmness does not depend on the limit")
	}
}

func TestAssessCountsInterlacements(t *testing.T) {
	plain := mustAssess(t, mustWeave(t, "plain-weave"), DefaultLimit)
	twill := mustAssess(t, mustWeave(t, "twill-2-2"), DefaultLimit)
	satin := mustAssess(t, mustWeave(t, "satin-8"), DefaultLimit)
	// The firmer the cloth, the more it interlaces, and plain weave interlaces twice per cell.
	if plain.Interlacements != 2*plain.Ends*plain.Picks {
		t.Fatalf("plain weave interlaces %d time(s)", plain.Interlacements)
	}
	if !(plain.Firmness > twill.Firmness && twill.Firmness > satin.Firmness) {
		t.Fatalf("the firmnesses went %g, %g, %g", plain.Firmness, twill.Firmness, satin.Firmness)
	}
}

func TestAssessRejectsBadInput(t *testing.T) {
	if _, err := Assess(cloth.Cloth{}, DefaultLimit); err == nil {
		t.Fatalf("an empty cloth must be reported")
	}
	for _, limit := range []int{0, -1, -5} {
		if _, err := Assess(mustWeave(t, "plain-weave"), limit); err == nil {
			t.Fatalf("a limit of %d = nil error, want a failure", limit)
		}
	}
}

func TestDescribe(t *testing.T) {
	got := mustAssess(t, mustWeave(t, "plain-weave"), DefaultLimit).Describe()
	for _, fragment := range []string{"8 end(s) by 8 pick(s)", "balanced", "longest float 1", "no fault"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("Describe = %q, which is missing %q", got, fragment)
		}
	}
	faulty := mustAssess(t, mustWeave(t, "floating-end"), DefaultLimit).Describe()
	if !strings.Contains(faulty, "fault(s)") {
		t.Fatalf("Describe = %q", faulty)
	}
}
