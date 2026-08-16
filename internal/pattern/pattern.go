// Package pattern names the structure a cloth has.
//
// Three structures account for most woven cloth, and each is a statement about the drawdown
// rather than about the draft that made it. Plain weave changes the thread on top at every
// cell in both directions. A twill repeats one pick shifted sideways by a fixed step, which is
// what makes the diagonal. A satin is a twill whose step is chosen so that the single
// interlacement of each pick never lands beside the one before it, and that is the one place
// in weaving where an arithmetic condition decides whether a cloth exists at all: the step has
// to share no factor with the number of shafts, and it cannot be one or one less than the
// number of shafts, because those two give a twill instead. Six shafts have no such step,
// which is why there is no six shaft satin.
package pattern

import (
	"fmt"
	"sort"
	"strings"

	"Heddle/internal/cloth"
	"Heddle/internal/draft"
	"Heddle/internal/shafts"
)

// IsPlainWeave reports whether the cloth changes the thread on top at every cell in both
// directions, which is what plain weave is.
func IsPlainWeave(c cloth.Cloth) (bool, error) {
	if err := c.Validate(); err != nil {
		return false, err
	}
	if c.Ends < 2 || c.Picks < 2 {
		return false, nil
	}
	for pick := 1; pick <= c.Picks; pick++ {
		for end := 1; end <= c.Ends; end++ {
			here, err := c.At(end, pick)
			if err != nil {
				return false, err
			}
			across, err := c.At(end%c.Ends+1, pick)
			if err != nil {
				return false, err
			}
			down, err := c.At(end, pick%c.Picks+1)
			if err != nil {
				return false, err
			}
			if here == across || here == down {
				return false, nil
			}
		}
	}
	return true, nil
}

// Twill is a cloth made of one pick shifted sideways for each following pick.
//
// Every count here is over the repeat rather than over the width of the piece: a 2/2 twill is
// a 2/2 twill whether it is woven over four ends or over four hundred.
type Twill struct {
	// Step is how far the pick moves for the next pick, counted to the right within the
	// repeat.
	Step int
	// Repeat is how many ends the twill repeats over.
	Repeat int
	// Shift is the same step counted the shorter way round, which is how a weaver reads the
	// slope of the diagonal.
	Shift int
	// Direction is right for a diagonal that climbs to the right and left for one that climbs
	// to the left.
	Direction string
	// Up and Down are how many ends the warp is over and under in each pick.
	Up   int
	Down int
}

// Describe renders the twill for a report.
func (t Twill) Describe() string {
	return fmt.Sprintf("%d/%d twill stepping %d to the %s", t.Up, t.Down, t.Shift, t.Direction)
}

// AsTwill reads the cloth as a twill, reporting whether it is one.
func AsTwill(c cloth.Cloth) (Twill, bool, error) {
	if err := c.Validate(); err != nil {
		return Twill{}, false, err
	}
	if c.Ends < 3 || c.Picks < 2 {
		return Twill{}, false, nil
	}
	width, _, err := Repeat(c)
	if err != nil {
		return Twill{}, false, err
	}
	first, err := c.Row(1)
	if err != nil {
		return Twill{}, false, err
	}
	up := 0
	for end := 0; end < width; end++ {
		if first[end] {
			up++
		}
	}
	if up == 0 || up == width {
		return Twill{}, false, nil
	}
	for step := 1; step < c.Ends; step++ {
		if !shiftsBy(c, step) {
			continue
		}
		reduced := step % width
		if reduced == 0 {
			continue
		}
		out := Twill{Step: reduced, Repeat: width, Up: up, Down: width - up}
		if reduced*2 <= width {
			out.Shift, out.Direction = reduced, "right"
		} else {
			out.Shift, out.Direction = width-reduced, "left"
		}
		return out, true, nil
	}
	return Twill{}, false, nil
}

// shiftsBy reports whether every pick is the pick before it moved by the step.
func shiftsBy(c cloth.Cloth, step int) bool {
	for pick := 1; pick <= c.Picks; pick++ {
		previous := (pick-2+c.Picks)%c.Picks + 1
		for end := 1; end <= c.Ends; end++ {
			source := (end-1-step+c.Ends*2)%c.Ends + 1
			here, err := c.At(end, pick)
			if err != nil {
				return false
			}
			there, err := c.At(source, previous)
			if err != nil {
				return false
			}
			if here != there {
				return false
			}
		}
	}
	return true
}

// gcd returns the greatest common divisor.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

// IsValidSatinCounter reports whether a step can be used for a satin on that many shafts.
//
// Two things have to hold. The step has to share no factor with the number of shafts, or the
// interlacements land on only some of the shafts and the repeat is smaller than the loom. And
// the step cannot be one, or one less than the number of shafts, because either of those puts
// each interlacement next to the one before it, which is a twill and not a satin.
func IsValidSatinCounter(loom, counter int) error {
	if loom < 5 || loom > shafts.MaxShafts {
		return fmt.Errorf("a satin needs at least 5 shafts, and %d is outside the range 5 to %d",
			loom, shafts.MaxShafts)
	}
	if counter < 1 || counter >= loom {
		return fmt.Errorf("a step of %d is not one of the steps of a %d shaft repeat", counter, loom)
	}
	if counter == 1 || counter == loom-1 {
		return fmt.Errorf("a step of %d on %d shafts puts every interlacement beside the one before it, which is a twill and not a satin",
			counter, loom)
	}
	if factor := gcd(counter, loom); factor != 1 {
		return fmt.Errorf("a step of %d and %d shafts share the factor %d, so the interlacements fall on only %d of the shafts",
			counter, loom, factor, loom/factor)
	}
	return nil
}

// SatinCounters returns the steps a satin on that many shafts can use, in order. There are
// none for four or six shafts, and that is a fact about the arithmetic rather than a gap in
// this list.
func SatinCounters(loom int) ([]int, error) {
	if loom < 5 || loom > shafts.MaxShafts {
		return nil, fmt.Errorf("a satin needs at least 5 shafts, and %d is outside the range 5 to %d",
			loom, shafts.MaxShafts)
	}
	out := []int{}
	for counter := 1; counter < loom; counter++ {
		if IsValidSatinCounter(loom, counter) != nil {
			continue
		}
		out = append(out, counter)
	}
	return out, nil
}

// SatinTieUp returns the tie-up of a satin: one shaft lifted per treadle for a weft faced
// satin, or all but one for a warp faced one, stepped by the counter.
func SatinTieUp(loom, counter int, warpFaced bool) (draft.TieUp, error) {
	if err := IsValidSatinCounter(loom, counter); err != nil {
		return nil, err
	}
	out := make(draft.TieUp, loom)
	for treadle := 1; treadle <= loom; treadle++ {
		shaft := ((treadle-1)*counter)%loom + 1
		set, err := shafts.New(shaft)
		if err != nil {
			return nil, err
		}
		if warpFaced {
			set, err = set.Complement(loom)
			if err != nil {
				return nil, err
			}
		}
		out[treadle-1] = set
	}
	return out, nil
}

// SatinDraft returns a whole satin draft over the given number of repeats.
func SatinDraft(loom, counter int, warpFaced bool, repeats int) (draft.Draft, error) {
	if repeats < 1 || repeats > 32 {
		return draft.Draft{}, fmt.Errorf("%d repeat(s) is outside the range 1 to 32", repeats)
	}
	tieup, err := SatinTieUp(loom, counter, warpFaced)
	if err != nil {
		return draft.Draft{}, err
	}
	threading, err := draft.Straight(loom, loom*repeats)
	if err != nil {
		return draft.Draft{}, err
	}
	treadling := make(draft.Treadling, loom*repeats)
	for index := range treadling {
		treadling[index] = index%loom + 1
	}
	face := "weft faced"
	if warpFaced {
		face = "warp faced"
	}
	out := draft.Draft{
		Name:      fmt.Sprintf("%d shaft %s satin stepping %d", loom, face, counter),
		Shafts:    loom,
		Treadles:  loom,
		Threading: threading,
		TieUp:     tieup,
		Treadling: treadling,
	}
	if err := out.Validate(); err != nil {
		return draft.Draft{}, err
	}
	return out, nil
}

// Satin is a cloth read as a satin.
type Satin struct {
	Shafts  int
	Counter int
	// Up is how many ends the warp is over in each pick, which is one for a weft faced satin.
	Up int
}

// Describe renders the satin for a report.
func (s Satin) Describe() string {
	face := "weft faced"
	if s.Up > 1 {
		face = "warp faced"
	}
	return fmt.Sprintf("%d shaft %s satin stepping %d", s.Shafts, face, s.Counter)
}

// AsSatin reads the cloth as a satin, reporting whether it is one. A satin is a twill whose
// step is a legal satin step and which interlaces once per pick.
func AsSatin(c cloth.Cloth) (Satin, bool, error) {
	twill, isTwill, err := AsTwill(c)
	if err != nil {
		return Satin{}, false, err
	}
	if !isTwill {
		return Satin{}, false, nil
	}
	if twill.Up != 1 && twill.Down != 1 {
		return Satin{}, false, nil
	}
	if err := IsValidSatinCounter(twill.Repeat, twill.Step); err != nil {
		return Satin{}, false, nil
	}
	return Satin{Shafts: twill.Repeat, Counter: twill.Step, Up: twill.Up}, true, nil
}

// Repeat returns the smallest block the cloth repeats over, in ends and in picks.
func Repeat(c cloth.Cloth) (int, int, error) {
	if err := c.Validate(); err != nil {
		return 0, 0, err
	}
	ends, err := repeatEnds(c)
	if err != nil {
		return 0, 0, err
	}
	picks, err := repeatPicks(c)
	if err != nil {
		return 0, 0, err
	}
	return ends, picks, nil
}

// repeatEnds returns the shortest run of ends the cloth repeats over.
func repeatEnds(c cloth.Cloth) (int, error) {
	for width := 1; width < c.Ends; width++ {
		if c.Ends%width != 0 {
			continue
		}
		same := true
		for pick := 1; pick <= c.Picks && same; pick++ {
			for end := width + 1; end <= c.Ends; end++ {
				here, err := c.At(end, pick)
				if err != nil {
					return 0, err
				}
				there, err := c.At((end-1)%width+1, pick)
				if err != nil {
					return 0, err
				}
				if here != there {
					same = false
					break
				}
			}
		}
		if same {
			return width, nil
		}
	}
	return c.Ends, nil
}

// repeatPicks returns the shortest run of picks the cloth repeats over.
func repeatPicks(c cloth.Cloth) (int, error) {
	for height := 1; height < c.Picks; height++ {
		if c.Picks%height != 0 {
			continue
		}
		same := true
		for pick := height + 1; pick <= c.Picks && same; pick++ {
			for end := 1; end <= c.Ends; end++ {
				here, err := c.At(end, pick)
				if err != nil {
					return 0, err
				}
				there, err := c.At(end, (pick-1)%height+1)
				if err != nil {
					return 0, err
				}
				if here != there {
					same = false
					break
				}
			}
		}
		if same {
			return height, nil
		}
	}
	return c.Picks, nil
}

// ThreadingSymmetric reports whether a threading reads the same from either selvedge, which is
// what makes a border turn cleanly at both edges.
func ThreadingSymmetric(t draft.Threading) bool {
	reversed := t.Reverse()
	if len(reversed) != len(t) {
		return false
	}
	for index := range t {
		if t[index] != reversed[index] {
			return false
		}
	}
	return true
}

// Classify names the structure of a cloth in a few words.
func Classify(c cloth.Cloth) (string, error) {
	plain, err := IsPlainWeave(c)
	if err != nil {
		return "", err
	}
	if plain {
		return "plain weave", nil
	}
	if satin, ok, err := AsSatin(c); err != nil {
		return "", err
	} else if ok {
		return satin.Describe(), nil
	}
	if twill, ok, err := AsTwill(c); err != nil {
		return "", err
	} else if ok {
		return twill.Describe(), nil
	}
	return "no plain weave, twill or satin", nil
}

// Analysis is what this package says about a cloth.
type Analysis struct {
	Ends        int
	Picks       int
	RepeatEnds  int
	RepeatPicks int
	PlainWeave  bool
	Twill       Twill
	IsTwill     bool
	Satin       Satin
	IsSatin     bool
	Structure   string
	// Balance and Face come from the cloth itself, and are here so that a reader of the
	// classification does not have to fetch them separately.
	Balance float64
	Face    string
}

// Analyse works out what this package says about a cloth.
func Analyse(c cloth.Cloth) (Analysis, error) {
	if err := c.Validate(); err != nil {
		return Analysis{}, err
	}
	out := Analysis{Ends: c.Ends, Picks: c.Picks, Balance: c.Balance(), Face: c.Face()}
	repeatEnds, repeatPicks, err := Repeat(c)
	if err != nil {
		return Analysis{}, err
	}
	out.RepeatEnds, out.RepeatPicks = repeatEnds, repeatPicks
	plain, err := IsPlainWeave(c)
	if err != nil {
		return Analysis{}, err
	}
	out.PlainWeave = plain
	twill, isTwill, err := AsTwill(c)
	if err != nil {
		return Analysis{}, err
	}
	out.Twill, out.IsTwill = twill, isTwill
	satin, isSatin, err := AsSatin(c)
	if err != nil {
		return Analysis{}, err
	}
	out.Satin, out.IsSatin = satin, isSatin
	structure, err := Classify(c)
	if err != nil {
		return Analysis{}, err
	}
	out.Structure = structure
	return out, nil
}

// Describe renders the analysis for a report.
func (a Analysis) Describe() string {
	return fmt.Sprintf("%d end(s) by %d pick(s) repeating over %d by %d, %s, %s",
		a.Ends, a.Picks, a.RepeatEnds, a.RepeatPicks, a.Structure, a.Face)
}

// CounterList renders a list of satin steps.
func CounterList(counters []int) string {
	if len(counters) == 0 {
		return "none"
	}
	sorted := append([]int(nil), counters...)
	sort.Ints(sorted)
	parts := make([]string, 0, len(sorted))
	for _, counter := range sorted {
		parts = append(parts, fmt.Sprintf("%d", counter))
	}
	return strings.Join(parts, " ")
}
