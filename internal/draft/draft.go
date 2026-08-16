// Package draft holds the four parts a weaving draft is made of.
//
// Threading says which shaft each warp end passes through. The tie-up says which shafts each
// treadle lifts. Treadling says which treadle is pressed for each pick of weft. Those three
// together determine the fourth, the cloth, completely: an end is raised over a pick exactly
// when the treadle for that pick lifts the shaft that end is threaded on. Nothing about the
// cloth is stored or chosen, and that is the whole point of a draft.
//
// The parts also constrain each other, and the constraints are checked here rather than left
// to whoever reads the cloth. A threading that names a shaft the loom does not have cannot be
// woven. A treadle that lifts every shaft, or none, opens no shed and cannot be trodden. A
// shaft that no end is threaded on does nothing at all, which is not an error but is worth
// saying out loud. And nothing here hands out a slice the caller does not own, because a
// draft is read many times over and altered as often.
package draft

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"Heddle/internal/shafts"
)

// The sizes this package works over.
const (
	MaxEnds  = 512
	MaxPicks = 512
	// MaxTreadles is the widest treadling this package works over. A loom with more treadles
	// than shafts is unusual but not impossible, so this is its own limit.
	MaxTreadles = 32
)

// Threading is the shaft each warp end is threaded on, the first end first.
type Threading []int

// Treadling is the treadle pressed for each pick of weft, the first pick first.
type Treadling []int

// TieUp is the set of shafts each treadle lifts, the first treadle first.
type TieUp []shafts.Set

// Straight returns a straight draw: the ends go across the shafts in order and start again.
func Straight(loom, ends int) (Threading, error) {
	if loom < 2 || loom > shafts.MaxShafts {
		return nil, fmt.Errorf("a loom of %d shaft(s) is outside the range 2 to %d",
			loom, shafts.MaxShafts)
	}
	if ends < 1 || ends > MaxEnds {
		return nil, fmt.Errorf("%d end(s) is outside the range 1 to %d", ends, MaxEnds)
	}
	out := make(Threading, ends)
	for index := range out {
		out[index] = index%loom + 1
	}
	return out, nil
}

// Point returns a point draw: the ends go up across the shafts and back down again, without
// repeating the shaft they turned on.
func Point(loom, ends int) (Threading, error) {
	if loom < 2 || loom > shafts.MaxShafts {
		return nil, fmt.Errorf("a loom of %d shaft(s) is outside the range 2 to %d",
			loom, shafts.MaxShafts)
	}
	if ends < 1 || ends > MaxEnds {
		return nil, fmt.Errorf("%d end(s) is outside the range 1 to %d", ends, MaxEnds)
	}
	period := 2 * (loom - 1)
	out := make(Threading, ends)
	for index := range out {
		step := index % period
		if step < loom {
			out[index] = step + 1
			continue
		}
		out[index] = period - step + 1
	}
	return out, nil
}

// Ends is how many warp ends the threading holds.
func (t Threading) Ends() int { return len(t) }

// Clone returns an independent copy.
func (t Threading) Clone() Threading { return append(Threading(nil), t...) }

// Reverse returns the threading read from the other selvedge, which is what threading the
// same draft from the other side of the loom gives.
func (t Threading) Reverse() Threading {
	out := make(Threading, len(t))
	for index := range t {
		out[index] = t[len(t)-1-index]
	}
	return out
}

// Validate checks the threading against a loom.
func (t Threading) Validate(loom int) error {
	if loom < 2 || loom > shafts.MaxShafts {
		return fmt.Errorf("a loom of %d shaft(s) is outside the range 2 to %d",
			loom, shafts.MaxShafts)
	}
	if len(t) < 1 || len(t) > MaxEnds {
		return fmt.Errorf("a threading of %d end(s) is outside the range 1 to %d", len(t), MaxEnds)
	}
	for index, shaft := range t {
		if shaft < 1 || shaft > loom {
			return fmt.Errorf("end %d is threaded on shaft %d, which is not one of the %d shafts of the loom",
				index+1, shaft, loom)
		}
	}
	return nil
}

// Usage counts the ends threaded on each shaft, shaft one first.
func (t Threading) Usage(loom int) []int {
	out := make([]int, loom)
	for _, shaft := range t {
		if shaft >= 1 && shaft <= loom {
			out[shaft-1]++
		}
	}
	return out
}

// EmptyShafts returns the shafts no end is threaded on.
func (t Threading) EmptyShafts(loom int) []int {
	out := []int{}
	for shaft, count := range t.Usage(loom) {
		if count == 0 {
			out = append(out, shaft+1)
		}
	}
	return out
}

// EndsOnShaft returns the ends threaded on one shaft, counting ends from one.
func (t Threading) EndsOnShaft(shaft int) []int {
	out := []int{}
	for index, current := range t {
		if current == shaft {
			out = append(out, index+1)
		}
	}
	return out
}

// String returns the printed form of the threading.
func (t Threading) String() string { return numbers(t) }

// Repeat is the shortest length the threading repeats over, which is the size of the block a
// weaver actually threads.
func (t Threading) Repeat() int { return period(t) }

// Treadling reuses the same helpers, because a treadling is the same kind of sequence read
// down the cloth instead of across it.

// Picks is how many picks the treadling holds.
func (t Treadling) Picks() int { return len(t) }

// Clone returns an independent copy.
func (t Treadling) Clone() Treadling { return append(Treadling(nil), t...) }

// Reverse returns the treadling read from the other end.
func (t Treadling) Reverse() Treadling {
	out := make(Treadling, len(t))
	for index := range t {
		out[index] = t[len(t)-1-index]
	}
	return out
}

// Validate checks the treadling against a set of treadles.
func (t Treadling) Validate(treadles int) error {
	if treadles < 1 || treadles > MaxTreadles {
		return fmt.Errorf("%d treadle(s) is outside the range 1 to %d", treadles, MaxTreadles)
	}
	if len(t) < 1 || len(t) > MaxPicks {
		return fmt.Errorf("a treadling of %d pick(s) is outside the range 1 to %d", len(t), MaxPicks)
	}
	for index, treadle := range t {
		if treadle < 1 || treadle > treadles {
			return fmt.Errorf("pick %d treads treadle %d, and there are %d treadle(s)",
				index+1, treadle, treadles)
		}
	}
	return nil
}

// Usage counts the picks made on each treadle, treadle one first.
func (t Treadling) Usage(treadles int) []int {
	out := make([]int, treadles)
	for _, treadle := range t {
		if treadle >= 1 && treadle <= treadles {
			out[treadle-1]++
		}
	}
	return out
}

// String returns the printed form of the treadling.
func (t Treadling) String() string { return numbers(t) }

// Repeat is the shortest length the treadling repeats over.
func (t Treadling) Repeat() int { return period(t) }

// numbers renders a sequence of small numbers.
func numbers[T ~int](values []T) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(int(value)))
	}
	return strings.Join(parts, " ")
}

// period returns the shortest length a sequence repeats over, which is its own length when it
// does not repeat.
func period[T ~int](values []T) int {
	for length := 1; length < len(values); length++ {
		repeats := true
		for index := length; index < len(values); index++ {
			if values[index] != values[index-length] {
				repeats = false
				break
			}
		}
		if repeats {
			return length
		}
	}
	return len(values)
}

// Treadles is how many treadles the tie-up covers.
func (u TieUp) Treadles() int { return len(u) }

// Clone returns an independent copy.
func (u TieUp) Clone() TieUp { return append(TieUp(nil), u...) }

// Validate checks the tie-up against a loom.
//
// A treadle that lifts no shaft, and a treadle that lifts every shaft, both open no shed:
// there is no gap for the shuttle to pass through, so the pick cannot be woven at all.
func (u TieUp) Validate(loom, treadles int) error {
	if loom < 2 || loom > shafts.MaxShafts {
		return fmt.Errorf("a loom of %d shaft(s) is outside the range 2 to %d",
			loom, shafts.MaxShafts)
	}
	if len(u) != treadles {
		return fmt.Errorf("the tie-up covers %d treadle(s) and the draft has %d", len(u), treadles)
	}
	if treadles < 1 || treadles > MaxTreadles {
		return fmt.Errorf("%d treadle(s) is outside the range 1 to %d", treadles, MaxTreadles)
	}
	for index, set := range u {
		if err := set.Validate(loom); err != nil {
			return fmt.Errorf("treadle %d: %w", index+1, err)
		}
		if set.Count() == loom {
			return fmt.Errorf("treadle %d lifts all %d shafts, which opens no shed", index+1, loom)
		}
	}
	return nil
}

// Transpose returns, for each shaft, the treadles that lift it. It is the same tie-up read the
// other way round, which is how a weaver checks that every shaft can be lifted at all.
func (u TieUp) Transpose(loom int) []shafts.Set {
	out := make([]shafts.Set, loom)
	for treadle, set := range u {
		for shaft := 1; shaft <= loom; shaft++ {
			if set.Has(shaft) {
				out[shaft-1] = out[shaft-1].Add(treadle + 1)
			}
		}
	}
	return out
}

// String returns the printed form of the tie-up.
func (u TieUp) String() string {
	parts := make([]string, 0, len(u))
	for _, set := range u {
		parts = append(parts, set.Compact())
	}
	return strings.Join(parts, " ")
}

// Draft is a whole draft: a loom, a threading, a tie-up and a treadling.
type Draft struct {
	Name      string
	Shafts    int
	Treadles  int
	Threading Threading
	TieUp     TieUp
	Treadling Treadling
}

// Ends is how many warp ends the draft holds.
func (d Draft) Ends() int { return len(d.Threading) }

// Picks is how many picks of weft the draft holds.
func (d Draft) Picks() int { return len(d.Treadling) }

// Clone returns an independent copy, so that a caller can alter one part without altering the
// draft it came from.
func (d Draft) Clone() Draft {
	out := d
	out.Threading = d.Threading.Clone()
	out.TieUp = d.TieUp.Clone()
	out.Treadling = d.Treadling.Clone()
	return out
}

// Validate checks the draft over.
func (d Draft) Validate() error {
	if strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("a draft needs a name")
	}
	if err := d.Threading.Validate(d.Shafts); err != nil {
		return fmt.Errorf("draft %s: %w", d.Name, err)
	}
	if err := d.TieUp.Validate(d.Shafts, d.Treadles); err != nil {
		return fmt.Errorf("draft %s: %w", d.Name, err)
	}
	if err := d.Treadling.Validate(d.Treadles); err != nil {
		return fmt.Errorf("draft %s: %w", d.Name, err)
	}
	return nil
}

// Lifted returns the shafts lifted for one pick, counting picks from one.
func (d Draft) Lifted(pick int) (shafts.Set, error) {
	if pick < 1 || pick > d.Picks() {
		return 0, fmt.Errorf("pick %d is not one of the %d picks", pick, d.Picks())
	}
	treadle := d.Treadling[pick-1]
	if treadle < 1 || treadle > len(d.TieUp) {
		return 0, fmt.Errorf("pick %d treads treadle %d, and the tie-up covers %d treadle(s)",
			pick, treadle, len(d.TieUp))
	}
	return d.TieUp[treadle-1], nil
}

// WarpUp reports whether the warp end is raised over the pick, which is the one question the
// cloth is made of.
func (d Draft) WarpUp(end, pick int) (bool, error) {
	if end < 1 || end > d.Ends() {
		return false, fmt.Errorf("end %d is not one of the %d ends", end, d.Ends())
	}
	lifted, err := d.Lifted(pick)
	if err != nil {
		return false, err
	}
	return lifted.Has(d.Threading[end-1]), nil
}

// TrompAsWrit returns the draft treadled the way it is threaded, which needs as many treadles
// as shafts because the treadling is read off the threading.
func (d Draft) TrompAsWrit() (Draft, error) {
	if err := d.Validate(); err != nil {
		return Draft{}, err
	}
	if d.Treadles != d.Shafts {
		return Draft{}, fmt.Errorf("draft %s has %d shaft(s) and %d treadle(s), so it cannot be treadled as it is threaded",
			d.Name, d.Shafts, d.Treadles)
	}
	out := d.Clone()
	out.Name = d.Name + " tromped as writ"
	out.Treadling = make(Treadling, len(d.Threading))
	for index, shaft := range d.Threading {
		out.Treadling[index] = shaft
	}
	return out, nil
}

// Mirror returns the draft threaded from the other selvedge.
func (d Draft) Mirror() (Draft, error) {
	if err := d.Validate(); err != nil {
		return Draft{}, err
	}
	out := d.Clone()
	out.Name = d.Name + " mirrored"
	out.Threading = d.Threading.Reverse()
	return out, nil
}

// Describe renders the draft for a report.
func (d Draft) Describe() string {
	return fmt.Sprintf("%s: %d shaft(s), %d treadle(s), %d end(s), %d pick(s), tie-up %s",
		d.Name, d.Shafts, d.Treadles, d.Ends(), d.Picks(), d.TieUp)
}

// sortedNames returns map keys in order.
func sortedNames(keys map[string]Draft) []string {
	out := make([]string, 0, len(keys))
	for key := range keys {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
