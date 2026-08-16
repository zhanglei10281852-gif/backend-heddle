// Package cloth is the drawdown: which warp ends are raised over which picks.
//
// The drawdown is the whole of what a draft produces, and reading it means reading floats. A
// float is a run of picks over which one end stays raised, or a run of ends over which one
// pick stays on top, and it is what decides whether the cloth holds together: a long float is
// a thread with nothing crossing it, which snags, slips and wears. Floats here are counted
// round the repeat rather than along the piece, because a draft is woven over and over: the
// last pick of the repeat is followed by the first, so a float that runs off one edge carries
// on at the other, and cutting it off at the edge understates exactly the longest floats a
// weaver cares about.
package cloth

import (
	"fmt"
	"sort"
	"strings"

	"Heddle/internal/draft"
)

// Cloth is a woven repeat: one cell per warp end per pick, true where the end is raised.
type Cloth struct {
	Ends  int
	Picks int
	// up holds the cells pick by pick, the ends of one pick together.
	up []bool
}

// Weave works out the cloth a draft produces.
func Weave(d draft.Draft) (Cloth, error) {
	if err := d.Validate(); err != nil {
		return Cloth{}, err
	}
	out := Cloth{
		Ends:  d.Ends(),
		Picks: d.Picks(),
		up:    make([]bool, d.Ends()*d.Picks()),
	}
	for pick := 1; pick <= out.Picks; pick++ {
		lifted, err := d.Lifted(pick)
		if err != nil {
			return Cloth{}, err
		}
		for end := 1; end <= out.Ends; end++ {
			out.up[(pick-1)*out.Ends+end-1] = lifted.Has(d.Threading[end-1])
		}
	}
	return out, nil
}

// New returns a cloth from its cells, given pick by pick.
func New(ends, picks int, cells []bool) (Cloth, error) {
	if ends < 1 || picks < 1 {
		return Cloth{}, fmt.Errorf("a cloth of %d end(s) by %d pick(s) is empty", ends, picks)
	}
	if len(cells) != ends*picks {
		return Cloth{}, fmt.Errorf("a cloth of %d end(s) by %d pick(s) needs %d cell(s), got %d",
			ends, picks, ends*picks, len(cells))
	}
	return Cloth{Ends: ends, Picks: picks, up: append([]bool(nil), cells...)}, nil
}

// Validate checks the cloth over.
func (c Cloth) Validate() error {
	if c.Ends < 1 || c.Picks < 1 {
		return fmt.Errorf("a cloth of %d end(s) by %d pick(s) is empty", c.Ends, c.Picks)
	}
	if len(c.up) != c.Ends*c.Picks {
		return fmt.Errorf("a cloth of %d end(s) by %d pick(s) holds %d cell(s)",
			c.Ends, c.Picks, len(c.up))
	}
	return nil
}

// Cells is how many cells the cloth holds.
func (c Cloth) Cells() int { return c.Ends * c.Picks }

// At reports whether the end is raised over the pick.
func (c Cloth) At(end, pick int) (bool, error) {
	if end < 1 || end > c.Ends {
		return false, fmt.Errorf("end %d is not one of the %d ends", end, c.Ends)
	}
	if pick < 1 || pick > c.Picks {
		return false, fmt.Errorf("pick %d is not one of the %d picks", pick, c.Picks)
	}
	return c.up[(pick-1)*c.Ends+end-1], nil
}

// Row returns one pick across all the ends, as a slice the caller owns.
func (c Cloth) Row(pick int) ([]bool, error) {
	if pick < 1 || pick > c.Picks {
		return nil, fmt.Errorf("pick %d is not one of the %d picks", pick, c.Picks)
	}
	out := make([]bool, c.Ends)
	copy(out, c.up[(pick-1)*c.Ends:pick*c.Ends])
	return out, nil
}

// Column returns one end down all the picks, as a slice the caller owns.
func (c Cloth) Column(end int) ([]bool, error) {
	if end < 1 || end > c.Ends {
		return nil, fmt.Errorf("end %d is not one of the %d ends", end, c.Ends)
	}
	out := make([]bool, c.Picks)
	for pick := 1; pick <= c.Picks; pick++ {
		out[pick-1] = c.up[(pick-1)*c.Ends+end-1]
	}
	return out, nil
}

// WarpUp counts the cells in which the warp is raised.
func (c Cloth) WarpUp() int {
	count := 0
	for _, raised := range c.up {
		if raised {
			count++
		}
	}
	return count
}

// Balance is the share of the cells in which the warp is raised. A half is a balanced cloth,
// more than that is warp faced and less is weft faced.
func (c Cloth) Balance() float64 {
	if c.Cells() == 0 {
		return 0
	}
	return float64(c.WarpUp()) / float64(c.Cells())
}

// Face names which side of the cloth shows.
func (c Cloth) Face() string {
	switch balance := c.Balance(); {
	case balance > 0.5:
		return "warp faced"
	case balance < 0.5:
		return "weft faced"
	default:
		return "balanced"
	}
}

// Equal reports whether two cloths hold the same cells.
func (c Cloth) Equal(other Cloth) bool {
	if c.Ends != other.Ends || c.Picks != other.Picks {
		return false
	}
	for index := range c.up {
		if c.up[index] != other.up[index] {
			return false
		}
	}
	return true
}

// Transpose returns the cloth with warp and weft exchanged, which is the same cloth turned
// over: what was a warp float becomes a weft float.
func (c Cloth) Transpose() Cloth {
	out := Cloth{Ends: c.Picks, Picks: c.Ends, up: make([]bool, len(c.up))}
	for pick := 1; pick <= c.Picks; pick++ {
		for end := 1; end <= c.Ends; end++ {
			out.up[(end-1)*out.Ends+pick-1] = !c.up[(pick-1)*c.Ends+end-1]
		}
	}
	return out
}

// Render draws the cloth, one line per pick with the first pick at the top, an x where the
// warp is raised and a dot where the weft crosses over it.
func (c Cloth) Render() string {
	var text strings.Builder
	for pick := 1; pick <= c.Picks; pick++ {
		for end := 1; end <= c.Ends; end++ {
			if c.up[(pick-1)*c.Ends+end-1] {
				text.WriteByte('x')
				continue
			}
			text.WriteByte('.')
		}
		text.WriteByte('\n')
	}
	return text.String()
}

// Float is a run of cells along one end or one pick in which the same thread stays on top.
type Float struct {
	// Along is warp for a run down one end and weft for a run across one pick.
	Along string
	// Index is the end or the pick the float lies on, counting from one.
	Index int
	// Start is the pick or the end the float starts at, counting from one.
	Start int
	// Length is how many cells the float covers.
	Length int
	// WarpUp says whether it is the warp that is on top over the float.
	WarpUp bool
}

// Describe renders the float for a report.
func (f Float) Describe() string {
	thread := "weft"
	if f.WarpUp {
		thread = "warp"
	}
	return fmt.Sprintf("%s float of %d on %s %d from %d, %s on top",
		f.Along, f.Length, f.Along, f.Index, f.Start, thread)
}

// runs returns the floats along one line of cells, counting round the repeat: the line is
// followed by itself, so a run that reaches the end carries on at the beginning.
func runs(cells []bool, along string, index int) []Float {
	out := []Float{}
	if len(cells) == 0 {
		return out
	}
	// A line that never changes is one float as long as the whole repeat.
	changed := false
	for position := 1; position < len(cells); position++ {
		if cells[position] != cells[position-1] {
			changed = true
			break
		}
	}
	if !changed {
		return append(out, Float{
			Along: along, Index: index, Start: 1, Length: len(cells), WarpUp: cells[0],
		})
	}
	// Start counting at a change, so that no float is cut in half by the edge.
	first := 0
	for position := 0; position < len(cells); position++ {
		previous := (position - 1 + len(cells)) % len(cells)
		if cells[position] != cells[previous] {
			first = position
			break
		}
	}
	start := first
	length := 1
	for step := 1; step <= len(cells); step++ {
		position := (first + step) % len(cells)
		previous := (first + step - 1) % len(cells)
		if step < len(cells) && cells[position] == cells[previous] {
			length++
			continue
		}
		out = append(out, Float{
			Along: along, Index: index, Start: start + 1, Length: length,
			WarpUp: cells[previous],
		})
		start = position
		length = 1
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start < out[j].Start })
	return out
}

// WarpFloats returns the floats down each end, in end order and then in start order.
func (c Cloth) WarpFloats() ([]Float, error) {
	out := []Float{}
	for end := 1; end <= c.Ends; end++ {
		column, err := c.Column(end)
		if err != nil {
			return nil, err
		}
		out = append(out, runs(column, "warp", end)...)
	}
	return out, nil
}

// WeftFloats returns the floats across each pick, in pick order and then in start order.
func (c Cloth) WeftFloats() ([]Float, error) {
	out := []Float{}
	for pick := 1; pick <= c.Picks; pick++ {
		row, err := c.Row(pick)
		if err != nil {
			return nil, err
		}
		out = append(out, runs(row, "weft", pick)...)
	}
	return out, nil
}

// Floats returns the floats along both directions.
func (c Cloth) Floats() ([]Float, error) {
	warp, err := c.WarpFloats()
	if err != nil {
		return nil, err
	}
	weft, err := c.WeftFloats()
	if err != nil {
		return nil, err
	}
	return append(warp, weft...), nil
}

// Longest returns the longest float in a list, and zero for an empty list.
func Longest(floats []Float) int {
	longest := 0
	for _, float := range floats {
		if float.Length > longest {
			longest = float.Length
		}
	}
	return longest
}

// Histogram counts the floats of each length.
func Histogram(floats []Float) map[int]int {
	out := map[int]int{}
	for _, float := range floats {
		out[float.Length]++
	}
	return out
}

// Over returns the floats longer than a limit, longest first and then in a settled order.
func Over(floats []Float, limit int) []Float {
	out := []Float{}
	for _, float := range floats {
		if float.Length > limit {
			out = append(out, float)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Length != out[j].Length {
			return out[i].Length > out[j].Length
		}
		if out[i].Along != out[j].Along {
			return out[i].Along < out[j].Along
		}
		if out[i].Index != out[j].Index {
			return out[i].Index < out[j].Index
		}
		return out[i].Start < out[j].Start
	})
	return out
}

// Interlacements counts the places where the thread on top changes, along the ends and along
// the picks, counted round the repeat. It is the plainest measure of how firm a cloth is:
// plain weave changes at every cell and a long float changes at neither of its ends.
func (c Cloth) Interlacements() (int, error) {
	count := 0
	for end := 1; end <= c.Ends; end++ {
		column, err := c.Column(end)
		if err != nil {
			return 0, err
		}
		count += changes(column)
	}
	for pick := 1; pick <= c.Picks; pick++ {
		row, err := c.Row(pick)
		if err != nil {
			return 0, err
		}
		count += changes(row)
	}
	return count, nil
}

// changes counts the places round a line where the value changes.
func changes(cells []bool) int {
	if len(cells) < 2 {
		return 0
	}
	count := 0
	for position := range cells {
		previous := (position - 1 + len(cells)) % len(cells)
		if cells[position] != cells[previous] {
			count++
		}
	}
	return count
}

// Firmness is the share of the possible interlacements the cloth actually makes, so plain
// weave comes out at one and a cloth of long floats comes out near zero.
func (c Cloth) Firmness() (float64, error) {
	interlacements, err := c.Interlacements()
	if err != nil {
		return 0, err
	}
	possible := 2 * c.Cells()
	if possible == 0 {
		return 0, nil
	}
	return float64(interlacements) / float64(possible), nil
}

// Describe renders the cloth for a report.
func (c Cloth) Describe() string {
	return fmt.Sprintf("%d end(s) by %d pick(s), %d of %d cell(s) warp up, %s",
		c.Ends, c.Picks, c.WarpUp(), c.Cells(), c.Face())
}
