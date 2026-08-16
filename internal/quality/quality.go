// Package quality reads a cloth for the faults a weaver would find on the loom.
//
// Three of them matter enough to be worth naming. A float longer than the cloth can carry is
// a thread with nothing crossing it, and it snags and wears. An end or a pick that never
// interlaces at all is worse: it is not woven in, it lies on the surface of the cloth, and it
// will pull straight out. And a cloth that interlaces very little across its whole area is
// sleazy: it holds together, but it slips and grows out of square.
//
// A report keeps hold of the cloth it was made from, because every one of those findings is a
// place in the cloth and a reader who is told there is a fault will want to be shown it.
package quality

import (
	"fmt"
	"sort"
	"strings"

	"Heddle/internal/cloth"
)

// DefaultLimit is the longest float this package lets pass without comment.
const DefaultLimit = 3

// SleazyBelow is the firmness under which a cloth is called sleazy.
const SleazyBelow = 0.35

// Report is what this package says about a cloth.
type Report struct {
	// woven is the cloth the report was made from, and what every finding below points into.
	woven *cloth.Cloth
	Ends  int
	Picks int
	// Limit is the longest float that was allowed to pass.
	Limit int
	// WarpUp, Balance and Face describe which side of the cloth shows.
	WarpUp  int
	Balance float64
	Face    string
	// LongestWarp and LongestWeft are the longest floats in each direction.
	LongestWarp int
	LongestWeft int
	// Over holds the floats longer than the limit, longest first.
	Over []cloth.Float
	// UnboundEnds and UnboundPicks are the ends and picks that never interlace, which is the
	// one fault that makes a cloth fall apart rather than merely wear badly.
	UnboundEnds  []int
	UnboundPicks []int
	// Interlacements and Firmness measure how much the cloth crosses itself.
	Interlacements int
	Firmness       float64
	Sleazy         bool
}

// Cloth returns the cloth the report was made from.
func (r Report) Cloth() cloth.Cloth { return *r.woven }

// Sound reports whether the cloth has none of the faults this package looks for.
func (r Report) Sound() bool {
	return len(r.Over) == 0 && len(r.UnboundEnds) == 0 && len(r.UnboundPicks) == 0 && !r.Sleazy
}

// Problems lists the faults in words, worst first.
func (r Report) Problems() []string {
	out := []string{}
	if len(r.UnboundEnds) > 0 {
		out = append(out, fmt.Sprintf("%d end(s) never interlace: %s",
			len(r.UnboundEnds), numbers(r.UnboundEnds)))
	}
	if len(r.UnboundPicks) > 0 {
		out = append(out, fmt.Sprintf("%d pick(s) never interlace: %s",
			len(r.UnboundPicks), numbers(r.UnboundPicks)))
	}
	if len(r.Over) > 0 {
		out = append(out, fmt.Sprintf("%d float(s) are longer than %d, the longest %d",
			len(r.Over), r.Limit, r.Over[0].Length))
	}
	if r.Sleazy {
		out = append(out, fmt.Sprintf("the cloth interlaces at %.0f%% of its cells, which is sleazy",
			100*r.Firmness))
	}
	return out
}

// Explain renders the cloth beside the findings, so that a fault can be looked at rather than
// taken on trust.
func (r Report) Explain() string {
	var text strings.Builder
	fmt.Fprintf(&text, "%d end(s) by %d pick(s), %s, firmness %.3f\n",
		r.woven.Ends, r.woven.Picks, r.woven.Face(), r.Firmness)
	text.WriteString(r.woven.Render())
	problems := r.Problems()
	if len(problems) == 0 {
		text.WriteString("no fault found\n")
		return text.String()
	}
	for _, problem := range problems {
		fmt.Fprintf(&text, "%s\n", problem)
	}
	return text.String()
}

// Worst returns the longest float in the cloth together with the cells it covers, which is the
// single thing a weaver would look at first.
func (r Report) Worst() (cloth.Float, string, error) {
	floats, err := r.woven.Floats()
	if err != nil {
		return cloth.Float{}, "", err
	}
	if len(floats) == 0 {
		return cloth.Float{}, "", fmt.Errorf("the cloth holds no float")
	}
	worst := floats[0]
	for _, float := range floats {
		if float.Length > worst.Length {
			worst = float
		}
	}
	line := []bool{}
	if worst.Along == "warp" {
		line, err = r.woven.Column(worst.Index)
	} else {
		line, err = r.woven.Row(worst.Index)
	}
	if err != nil {
		return cloth.Float{}, "", err
	}
	return worst, render(line), nil
}

// render draws one line of cells.
func render(cells []bool) string {
	text := make([]byte, 0, len(cells))
	for _, raised := range cells {
		if raised {
			text = append(text, 'x')
			continue
		}
		text = append(text, '.')
	}
	return string(text)
}

// numbers renders a list of indexes.
func numbers(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%d", value))
	}
	return strings.Join(parts, " ")
}

// allSame reports whether every cell of a line holds the same thread on top, which is what an
// end or a pick that never interlaces looks like.
func allSame(cells []bool) bool {
	if len(cells) < 2 {
		return true
	}
	for index := 1; index < len(cells); index++ {
		if cells[index] != cells[0] {
			return false
		}
	}
	return true
}

// Assess reads a cloth for faults, allowing floats up to the given limit.
func Assess(c cloth.Cloth, limit int) (Report, error) {
	if err := c.Validate(); err != nil {
		return Report{}, err
	}
	if limit < 1 {
		return Report{}, fmt.Errorf("a float limit of %d allows no float at all", limit)
	}
	held := c
	out := Report{
		woven:        &held,
		Ends:         c.Ends,
		Picks:        c.Picks,
		Limit:        limit,
		WarpUp:       c.WarpUp(),
		Balance:      c.Balance(),
		Face:         c.Face(),
		Over:         []cloth.Float{},
		UnboundEnds:  []int{},
		UnboundPicks: []int{},
	}
	warp, err := c.WarpFloats()
	if err != nil {
		return Report{}, err
	}
	weft, err := c.WeftFloats()
	if err != nil {
		return Report{}, err
	}
	out.LongestWarp = cloth.Longest(warp)
	out.LongestWeft = cloth.Longest(weft)
	out.Over = cloth.Over(append(append([]cloth.Float{}, warp...), weft...), limit)

	for end := 1; end <= c.Ends; end++ {
		column, err := c.Column(end)
		if err != nil {
			return Report{}, err
		}
		if allSame(column) {
			out.UnboundEnds = append(out.UnboundEnds, end)
		}
	}
	for pick := 1; pick <= c.Picks; pick++ {
		row, err := c.Row(pick)
		if err != nil {
			return Report{}, err
		}
		if allSame(row) {
			out.UnboundPicks = append(out.UnboundPicks, pick)
		}
	}
	sort.Ints(out.UnboundEnds)
	sort.Ints(out.UnboundPicks)

	interlacements, err := c.Interlacements()
	if err != nil {
		return Report{}, err
	}
	out.Interlacements = interlacements
	firmness, err := c.Firmness()
	if err != nil {
		return Report{}, err
	}
	out.Firmness = firmness
	out.Sleazy = firmness < SleazyBelow
	return out, nil
}

// Histogram counts the floats of each length in the cloth the report was made from.
func (r Report) Histogram() (map[int]int, error) {
	floats, err := r.woven.Floats()
	if err != nil {
		return nil, err
	}
	return cloth.Histogram(floats), nil
}

// Describe renders the report for a summary line.
func (r Report) Describe() string {
	verdict := fmt.Sprintf("%d fault(s)", len(r.Problems()))
	if r.Sound() {
		verdict = "no fault"
	}
	return fmt.Sprintf("%d end(s) by %d pick(s), %s, longest float %d warp and %d weft, firmness %.3f, %s",
		r.Ends, r.Picks, r.Face, r.LongestWarp, r.LongestWeft, r.Firmness, verdict)
}
