// Package shafts holds a set of shafts.
//
// A treadle lifts some shafts and leaves the rest down, and that is all a set here is. The
// reason it gets a package of its own is that the loom has a fixed number of shafts and a set
// has to be checked against it. A tie-up that names shaft 9 on an eight shaft loom is not a
// tie-up with a small mistake in it: it will still lift, it will still print, and the ends it
// was meant to raise will simply stay down, which shows up in the cloth as a pick with no
// warp in it rather than as an error. So a set is checked against the loom when it is read,
// and a set that names nothing at all is refused for the same reason.
package shafts

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// MaxShafts is the widest loom this package works over. Thirty-two fits a uint32 and is more
// shafts than any of the drafts this tool is for.
const MaxShafts = 32

// Set is a set of shafts, held as one bit per shaft with shaft 1 in the lowest bit.
type Set uint32

// New returns the set holding the given shafts.
func New(shafts ...int) (Set, error) {
	out := Set(0)
	for _, shaft := range shafts {
		if shaft < 1 || shaft > MaxShafts {
			return 0, fmt.Errorf("shaft %d is outside the range 1 to %d", shaft, MaxShafts)
		}
		if out.Has(shaft) {
			return 0, fmt.Errorf("shaft %d is named twice", shaft)
		}
		out = out.Add(shaft)
	}
	return out, nil
}

// Has reports whether the set holds a shaft.
func (s Set) Has(shaft int) bool {
	if shaft < 1 || shaft > MaxShafts {
		return false
	}
	return s&(1<<uint(shaft-1)) != 0
}

// Add returns the set with a shaft in it.
func (s Set) Add(shaft int) Set {
	if shaft < 1 || shaft > MaxShafts {
		return s
	}
	return s | 1<<uint(shaft-1)
}

// Remove returns the set without a shaft.
func (s Set) Remove(shaft int) Set {
	if shaft < 1 || shaft > MaxShafts {
		return s
	}
	return s &^ (1 << uint(shaft-1))
}

// Count is how many shafts the set holds.
func (s Set) Count() int {
	count := 0
	for shaft := 1; shaft <= MaxShafts; shaft++ {
		if s.Has(shaft) {
			count++
		}
	}
	return count
}

// IsEmpty reports whether the set holds no shaft.
func (s Set) IsEmpty() bool { return s == 0 }

// Shafts returns the shafts in the set, in order, as a slice the caller owns.
func (s Set) Shafts() []int {
	out := []int{}
	for shaft := 1; shaft <= MaxShafts; shaft++ {
		if s.Has(shaft) {
			out = append(out, shaft)
		}
	}
	return out
}

// Highest is the largest shaft in the set, or zero for an empty set.
func (s Set) Highest() int {
	highest := 0
	for shaft := 1; shaft <= MaxShafts; shaft++ {
		if s.Has(shaft) {
			highest = shaft
		}
	}
	return highest
}

// Union returns the shafts in either set.
func (s Set) Union(other Set) Set { return s | other }

// Intersect returns the shafts in both sets.
func (s Set) Intersect(other Set) Set { return s & other }

// Difference returns the shafts in this set and not the other.
func (s Set) Difference(other Set) Set { return s &^ other }

// Complement returns the shafts of the loom that the set does not hold, which is what stays
// down when the set is lifted.
func (s Set) Complement(shafts int) (Set, error) {
	if err := checkLoom(shafts); err != nil {
		return 0, err
	}
	out := Set(0)
	for shaft := 1; shaft <= shafts; shaft++ {
		if !s.Has(shaft) {
			out = out.Add(shaft)
		}
	}
	return out, nil
}

// Equal reports whether two sets hold the same shafts.
func (s Set) Equal(other Set) bool { return s == other }

// Compare orders two sets, by how many shafts they hold and then by the shafts themselves.
func Compare(a, b Set) int {
	if a.Count() != b.Count() {
		if a.Count() < b.Count() {
			return -1
		}
		return 1
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// String returns the printed form of the set: the shafts in order, separated by full stops.
func (s Set) String() string {
	if s.IsEmpty() {
		return "-"
	}
	parts := make([]string, 0, s.Count())
	for _, shaft := range s.Shafts() {
		parts = append(parts, strconv.Itoa(shaft))
	}
	return strings.Join(parts, ".")
}

// Compact returns the printed form used when no shaft needs two digits, which is how a tie-up
// is usually written down: 12 for shafts one and two.
func (s Set) Compact() string {
	if s.IsEmpty() {
		return "-"
	}
	if s.Highest() > 9 {
		return s.String()
	}
	text := make([]byte, 0, s.Count())
	for _, shaft := range s.Shafts() {
		text = append(text, byte('0'+shaft))
	}
	return string(text)
}

// Grid returns the set as one character per shaft, x for lifted and dot for down, shaft one
// first, which is how a tie-up is drawn on squared paper.
func (s Set) Grid(shafts int) string {
	text := make([]byte, 0, shafts)
	for shaft := 1; shaft <= shafts; shaft++ {
		if s.Has(shaft) {
			text = append(text, 'x')
			continue
		}
		text = append(text, '.')
	}
	return string(text)
}

// Validate checks the set against a loom of the given number of shafts.
func (s Set) Validate(shafts int) error {
	if err := checkLoom(shafts); err != nil {
		return err
	}
	if s.IsEmpty() {
		return fmt.Errorf("a set of shafts cannot be empty")
	}
	if highest := s.Highest(); highest > shafts {
		return fmt.Errorf("shaft %d is not one of the %d shafts of the loom", highest, shafts)
	}
	return nil
}

// checkLoom checks a shaft count.
func checkLoom(shafts int) error {
	if shafts < 2 || shafts > MaxShafts {
		return fmt.Errorf("a loom of %d shaft(s) is outside the range 2 to %d", shafts, MaxShafts)
	}
	return nil
}

// Parse reads a set of shafts against a loom.
//
// Two forms are read. A run of single digits is the usual way a tie-up is written, so 12 is
// shafts one and two; and full stops separate the shafts when any of them needs two digits,
// so 1.10.12 is shafts one, ten and twelve. A single dash is refused rather than read as the
// empty set, because a treadle that lifts nothing opens no shed.
func Parse(text string, shafts int) (Set, error) {
	if err := checkLoom(shafts); err != nil {
		return 0, err
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0, fmt.Errorf("a set of shafts needs at least one shaft")
	}
	numbers := []string{}
	if strings.Contains(trimmed, ".") {
		numbers = strings.Split(trimmed, ".")
	} else {
		for index := 0; index < len(trimmed); index++ {
			numbers = append(numbers, trimmed[index:index+1])
		}
	}
	out := Set(0)
	for _, part := range numbers {
		field := strings.TrimSpace(part)
		if field == "" {
			return 0, fmt.Errorf("shaft set %q has an empty shaft", trimmed)
		}
		shaft, err := strconv.Atoi(field)
		if err != nil {
			return 0, fmt.Errorf("shaft set %q: %q is not a shaft number", trimmed, field)
		}
		if shaft < 1 {
			return 0, fmt.Errorf("shaft set %q: shaft %d is not a shaft", trimmed, shaft)
		}
		if shaft > shafts {
			return 0, fmt.Errorf("shaft set %q: shaft %d is not one of the %d shafts of the loom",
				trimmed, shaft, shafts)
		}
		if out.Has(shaft) {
			return 0, fmt.Errorf("shaft set %q: shaft %d is named twice", trimmed, shaft)
		}
		out = out.Add(shaft)
	}
	if err := out.Validate(shafts); err != nil {
		return 0, fmt.Errorf("shaft set %q: %w", trimmed, err)
	}
	return out, nil
}

// ParseGrid reads a set written as one character per shaft, where anything other than a dot,
// a space or a zero lifts that shaft.
func ParseGrid(text string, shafts int) (Set, error) {
	if err := checkLoom(shafts); err != nil {
		return 0, err
	}
	trimmed := strings.TrimSpace(text)
	if len(trimmed) != shafts {
		return 0, fmt.Errorf("a tie-up column for %d shaft(s) needs %d character(s), got %d",
			shafts, shafts, len(trimmed))
	}
	out := Set(0)
	for index := 0; index < len(trimmed); index++ {
		switch trimmed[index] {
		case '.', ' ', '0', '-':
		default:
			out = out.Add(index + 1)
		}
	}
	if err := out.Validate(shafts); err != nil {
		return 0, fmt.Errorf("tie-up column %q: %w", trimmed, err)
	}
	return out, nil
}

// Sort puts a list of sets in order.
func Sort(sets []Set) {
	sort.Slice(sets, func(i, j int) bool { return Compare(sets[i], sets[j]) < 0 })
}

// Describe renders the set for a report.
func (s Set) Describe(shafts int) string {
	return fmt.Sprintf("%s, %d of %d shaft(s) lifted, %s", s.Compact(), s.Count(), shafts, s.Grid(shafts))
}
