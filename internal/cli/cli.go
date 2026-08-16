// Package cli is the Heddle command surface.
//
// Exit codes are part of the contract: 0 means the command ran and produced an answer, 1 means
// it could not run, and 2 means it ran and there was nothing to report. The last one earns its
// place here because a number of shafts with no satin step at all, and a cloth with no float
// over the limit, are both answers a script has to be able to act on.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"Heddle/internal/cloth"
	"Heddle/internal/draft"
	"Heddle/internal/pattern"
	"Heddle/internal/quality"
	"Heddle/internal/report"
)

// Version identifies the build.
const Version = "heddle 1.0.0"

// Exit codes.
const (
	ExitOK    = 0
	ExitError = 1
	ExitEmpty = 2
)

// MaxDrawnEnds and MaxDrawnPicks are how large a cloth this tool draws out cell by cell.
const (
	MaxDrawnEnds  = 96
	MaxDrawnPicks = 96
)

// command is one subcommand.
type command struct {
	name    string
	summary string
	run     func(args []string, stdout, stderr io.Writer) (int, error)
}

// commands returns the subcommand table.
func commands() map[string]command {
	out := map[string]command{}
	for _, item := range []command{
		{"draft", "read a draft and report its four parts", runDraft},
		{"cloth", "weave a draft out and draw the cloth", runCloth},
		{"floats", "the floats of a cloth along the warp and the weft", runFloats},
		{"pattern", "name the structure a cloth has", runPattern},
		{"quality", "read a cloth for the faults a weaver would find", runQuality},
		{"satin", "the steps a satin can use, and the satin they give", runSatin},
		{"tromp", "treadle a draft the way it is threaded and compare", runTromp},
		{"drafts", "the drafts this tool knows, with their checks", runDrafts},
		{"report", "one figure per package for a draft", runReport},
	} {
		out[item.name] = item
	}
	return out
}

// Run dispatches one invocation.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return ExitError
	}
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, Version)
		fmt.Fprintln(stdout, "study tool only: a teaching implementation of weaving draft arithmetic")
		return ExitOK
	case "help", "--help", "-h":
		usage(stdout)
		return ExitOK
	}
	chosen, known := commands()[args[0]]
	if !known {
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		usage(stderr)
		return ExitError
	}
	code, err := chosen.run(args[1:], stdout, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ExitError
	}
	return code
}

// usage prints the command surface.
func usage(writer io.Writer) {
	fmt.Fprintln(writer, Version)
	fmt.Fprintln(writer, "study tool only: a teaching implementation of weaving draft arithmetic")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "usage: heddle <command> [flags]")
	fmt.Fprintln(writer, "")
	table := report.Table{Gap: 2}
	names := make([]string, 0, len(commands()))
	for name := range commands() {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		table.AddRow("  "+name, commands()[name].summary)
	}
	fmt.Fprint(writer, table.Render())
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "exit codes: 0 produced an answer, 1 could not run, 2 ran and found nothing")
	if names, err := draft.Names(); err == nil {
		fmt.Fprintln(writer, "drafts: "+strings.Join(names, ", "))
	}
	fmt.Fprintf(writer, "a cloth is drawn with x where the warp is raised and a dot where the weft crosses it\n")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "the drafts and the float limits are this tool's own")
}

// parse reads the flags of one subcommand. A request for help is a successful run rather than a
// failure, and a bad flag reports itself once instead of once per layer.
func parse(set *flag.FlagSet, args []string, stdout, stderr io.Writer) (int, bool) {
	set.SetOutput(io.Discard)
	err := set.Parse(args)
	switch {
	case err == nil:
		return ExitOK, true
	case errors.Is(err, flag.ErrHelp):
		fmt.Fprintf(stdout, "%s - %s\n\n", set.Name(), commands()[set.Name()].summary)
		set.SetOutput(stdout)
		set.PrintDefaults()
		return ExitOK, false
	default:
		fmt.Fprintln(stderr, err)
		set.SetOutput(stderr)
		set.PrintDefaults()
		return ExitError, false
	}
}

// draftFlags declares the flags every command that needs a draft shares.
type draftFlags struct {
	name   *string
	file   *string
	mirror *bool
	tromp  *bool
}

// declareDraft adds the draft flags to a set.
func declareDraft(set *flag.FlagSet) *draftFlags {
	return &draftFlags{
		name:   set.String("draft", "twill-2-2", "draft from the catalogue"),
		file:   set.String("file", "", "read the draft from this file"),
		mirror: set.Bool("mirror", false, "thread the draft from the other selvedge"),
		tromp:  set.Bool("tromp", false, "treadle the draft the way it is threaded"),
	}
}

// resolve returns the draft a command should work on.
func (d *draftFlags) resolve() (draft.Draft, error) {
	var built draft.Draft
	var err error
	if strings.TrimSpace(*d.file) != "" {
		text, readErr := os.ReadFile(*d.file)
		if readErr != nil {
			return draft.Draft{}, readErr
		}
		built, err = draft.ParseDraft(string(text))
		if err != nil {
			return draft.Draft{}, fmt.Errorf("%s: %w", *d.file, err)
		}
	} else {
		built, err = draft.Lookup(*d.name)
		if err != nil {
			return draft.Draft{}, err
		}
	}
	if *d.tromp {
		built, err = built.TrompAsWrit()
		if err != nil {
			return draft.Draft{}, err
		}
	}
	if *d.mirror {
		built, err = built.Mirror()
		if err != nil {
			return draft.Draft{}, err
		}
	}
	return built, nil
}

// woven returns the draft a command should work on together with its cloth.
func woven(flags *draftFlags) (draft.Draft, cloth.Cloth, error) {
	built, err := flags.resolve()
	if err != nil {
		return draft.Draft{}, cloth.Cloth{}, err
	}
	fabric, err := cloth.Weave(built)
	if err != nil {
		return draft.Draft{}, cloth.Cloth{}, err
	}
	return built, fabric, nil
}

// runDraft reads a draft and reports its four parts.
func runDraft(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("draft", flag.ContinueOnError)
	flags := declareDraft(set)
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	built, err := flags.resolve()
	if err != nil {
		return ExitError, err
	}
	fmt.Fprintf(stdout, "%s\n\n", built.Describe())
	table := report.Table{
		Header:     []string{"part", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignLeft},
	}
	table.AddRow("shafts", report.Int(built.Shafts))
	table.AddRow("treadles", report.Int(built.Treadles))
	table.AddRow("ends", report.Int(built.Ends()))
	table.AddRow("picks", report.Int(built.Picks()))
	table.AddRow("threading", built.Threading.String())
	table.AddRow("threading repeat", report.Int(built.Threading.Repeat()))
	table.AddRow("tie-up", built.TieUp.String())
	table.AddRow("treadling", built.Treadling.String())
	table.AddRow("treadling repeat", report.Int(built.Treadling.Repeat()))
	fmt.Fprint(stdout, table.Render())

	tieup := report.Table{
		Header:     []string{"treadle", "shafts", "lifted", "grid"},
		Alignments: []report.Alignment{report.AlignRight, report.AlignRight, report.AlignRight, report.AlignLeft},
	}
	for index, lifted := range built.TieUp {
		tieup.AddRow(report.Int(index+1), lifted.Compact(),
			report.Int(lifted.Count()), lifted.Grid(built.Shafts))
	}
	fmt.Fprint(stdout, "\n"+tieup.Render())

	usage := report.Table{
		Header:     []string{"shaft", "ends", "lifted by", "first ends"},
		Alignments: []report.Alignment{report.AlignRight, report.AlignRight, report.AlignRight, report.AlignLeft},
	}
	transposed := built.TieUp.Transpose(built.Shafts)
	counts := built.Threading.Usage(built.Shafts)
	for shaft := 1; shaft <= built.Shafts; shaft++ {
		ends := built.Threading.EndsOnShaft(shaft)
		if len(ends) > 8 {
			ends = ends[:8]
		}
		usage.AddRow(report.Int(shaft), report.Int(counts[shaft-1]),
			transposed[shaft-1].Compact(), report.Ints(ends))
	}
	fmt.Fprint(stdout, "\n"+usage.Render())

	mirrored := built.Threading.Reverse()
	summary := report.Table{
		Header:     []string{"measure", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignLeft},
	}
	summary.AddRow("threading from the other selvedge", mirrored.String())
	summary.AddRow("reads the same from either selvedge",
		report.Bool(pattern.ThreadingSymmetric(built.Threading)))
	if empty := built.Threading.EmptyShafts(built.Shafts); len(empty) > 0 {
		summary.AddRow("shafts with no end on them", report.Ints(empty))
	} else {
		summary.AddRow("shafts with no end on them", "none")
	}
	summary.AddRow("treadles used", report.Ints(usedTreadles(built)))
	fmt.Fprint(stdout, "\n"+summary.Render())
	return ExitOK, nil
}

// usedTreadles returns the treadles the treadling actually treads.
func usedTreadles(d draft.Draft) []int {
	out := []int{}
	for treadle, count := range d.Treadling.Usage(d.Treadles) {
		if count > 0 {
			out = append(out, treadle+1)
		}
	}
	return out
}

// runCloth weaves a draft out and draws the cloth.
func runCloth(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("cloth", flag.ContinueOnError)
	flags := declareDraft(set)
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	built, fabric, err := woven(flags)
	if err != nil {
		return ExitError, err
	}
	fmt.Fprintf(stdout, "%s\n\n", built.Describe())
	if fabric.Ends <= MaxDrawnEnds && fabric.Picks <= MaxDrawnPicks {
		fmt.Fprint(stdout, drawn(fabric, built))
		fmt.Fprintln(stdout, "")
	} else {
		fmt.Fprintf(stdout, "the cloth is %d end(s) by %d pick(s), which is more than the %d by %d this tool draws\n\n",
			fabric.Ends, fabric.Picks, MaxDrawnEnds, MaxDrawnPicks)
	}
	structure, err := pattern.Classify(fabric)
	if err != nil {
		return ExitError, err
	}
	repeatEnds, repeatPicks, err := pattern.Repeat(fabric)
	if err != nil {
		return ExitError, err
	}
	interlacements, err := fabric.Interlacements()
	if err != nil {
		return ExitError, err
	}
	firmness, err := fabric.Firmness()
	if err != nil {
		return ExitError, err
	}
	table := report.Table{
		Header:     []string{"measure", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignRight},
	}
	table.AddRow("ends", report.Int(fabric.Ends))
	table.AddRow("picks", report.Int(fabric.Picks))
	table.AddRow("cells", report.Int(fabric.Cells()))
	table.AddRow("cells with the warp raised", report.Int(fabric.WarpUp()))
	table.AddRow("balance", report.Float(fabric.Balance()))
	table.AddRow("face", fabric.Face())
	table.AddRow("repeat in ends", report.Int(repeatEnds))
	table.AddRow("repeat in picks", report.Int(repeatPicks))
	table.AddRow("interlacements", report.Int(interlacements))
	table.AddRow("firmness", report.Float(firmness))
	table.AddRow("structure", structure)
	fmt.Fprint(stdout, table.Render())
	return ExitOK, nil
}

// drawn renders the cloth with the tie-up and the treadling beside it, which is how a draft is
// drawn on paper.
func drawn(fabric cloth.Cloth, built draft.Draft) string {
	var text strings.Builder
	for pick := 1; pick <= fabric.Picks; pick++ {
		row, err := fabric.Row(pick)
		if err != nil {
			return err.Error()
		}
		for _, raised := range row {
			if raised {
				text.WriteByte('x')
				continue
			}
			text.WriteByte('.')
		}
		lifted, err := built.Lifted(pick)
		if err != nil {
			return err.Error()
		}
		fmt.Fprintf(&text, "  %2d %s\n", built.Treadling[pick-1], lifted.Compact())
	}
	return text.String()
}

// runFloats reports the floats of a cloth.
func runFloats(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("floats", flag.ContinueOnError)
	flags := declareDraft(set)
	over := set.Int("over", quality.DefaultLimit, "report the floats longer than this")
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	built, fabric, err := woven(flags)
	if err != nil {
		return ExitError, err
	}
	if *over < 1 {
		return ExitError, fmt.Errorf("a float limit of %d allows no float at all", *over)
	}
	warp, err := fabric.WarpFloats()
	if err != nil {
		return ExitError, err
	}
	weft, err := fabric.WeftFloats()
	if err != nil {
		return ExitError, err
	}
	all := append(append([]cloth.Float{}, warp...), weft...)
	fmt.Fprintf(stdout, "%s\n\n", built.Describe())
	table := report.Table{
		Header:     []string{"measure", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignRight},
	}
	table.AddRow("floats along the warp", report.Int(len(warp)))
	table.AddRow("floats along the weft", report.Int(len(weft)))
	table.AddRow("longest warp float", report.Int(cloth.Longest(warp)))
	table.AddRow("longest weft float", report.Int(cloth.Longest(weft)))
	table.AddRow("limit", report.Int(*over))
	fmt.Fprint(stdout, table.Render())

	histogram := cloth.Histogram(all)
	lengths := make([]int, 0, len(histogram))
	for length := range histogram {
		lengths = append(lengths, length)
	}
	sort.Ints(lengths)
	largest := 0
	for _, count := range histogram {
		if count > largest {
			largest = count
		}
	}
	block := report.Table{
		Header:     []string{"length", "floats", ""},
		Alignments: []report.Alignment{report.AlignRight, report.AlignRight, report.AlignLeft},
	}
	for _, length := range lengths {
		block.AddRow(report.Int(length), report.Int(histogram[length]),
			report.Bar(histogram[length], largest, 28))
	}
	fmt.Fprint(stdout, "\n"+block.Render())

	long := cloth.Over(all, *over)
	if len(long) == 0 {
		fmt.Fprintf(stdout, "\nno float is longer than %d\n", *over)
		return ExitEmpty, nil
	}
	worst := report.Table{
		Header: []string{"along", "index", "from", "length", "on top"},
		Alignments: []report.Alignment{
			report.AlignLeft, report.AlignRight, report.AlignRight,
			report.AlignRight, report.AlignLeft,
		},
	}
	for index, float := range long {
		if index >= 30 {
			break
		}
		thread := "weft"
		if float.WarpUp {
			thread = "warp"
		}
		worst.AddRow(float.Along, report.Int(float.Index), report.Int(float.Start),
			report.Int(float.Length), thread)
	}
	fmt.Fprint(stdout, "\n"+worst.Render())
	if len(long) > 30 {
		fmt.Fprintf(stdout, "\n%d more float(s) are longer than %d\n", len(long)-30, *over)
	}
	return ExitOK, nil
}

// runPattern names the structure a cloth has.
func runPattern(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("pattern", flag.ContinueOnError)
	flags := declareDraft(set)
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	built, fabric, err := woven(flags)
	if err != nil {
		return ExitError, err
	}
	analysis, err := pattern.Analyse(fabric)
	if err != nil {
		return ExitError, err
	}
	fmt.Fprintf(stdout, "%s\n\n", built.Describe())
	table := report.Table{
		Header:     []string{"measure", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignRight},
	}
	table.AddRow("ends", report.Int(analysis.Ends))
	table.AddRow("picks", report.Int(analysis.Picks))
	table.AddRow("repeat in ends", report.Int(analysis.RepeatEnds))
	table.AddRow("repeat in picks", report.Int(analysis.RepeatPicks))
	table.AddRow("plain weave", report.Bool(analysis.PlainWeave))
	table.AddRow("twill", report.Bool(analysis.IsTwill))
	if analysis.IsTwill {
		table.AddRow("twill step", report.Int(analysis.Twill.Step))
		table.AddRow("twill slope", report.Int(analysis.Twill.Shift))
		table.AddRow("twill direction", analysis.Twill.Direction)
		table.AddRow("warp over", report.Int(analysis.Twill.Up))
		table.AddRow("warp under", report.Int(analysis.Twill.Down))
	}
	table.AddRow("satin", report.Bool(analysis.IsSatin))
	if analysis.IsSatin {
		table.AddRow("satin step", report.Int(analysis.Satin.Counter))
		table.AddRow("satin shafts", report.Int(analysis.Satin.Shafts))
	}
	table.AddRow("balance", report.Float(analysis.Balance))
	table.AddRow("face", analysis.Face)
	table.AddRow("structure", analysis.Structure)
	fmt.Fprint(stdout, table.Render())
	fmt.Fprintf(stdout, "\n%s\n", analysis.Describe())
	return ExitOK, nil
}

// runQuality reads a cloth for the faults a weaver would find.
func runQuality(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("quality", flag.ContinueOnError)
	flags := declareDraft(set)
	limit := set.Int("limit", quality.DefaultLimit, "longest float to let pass")
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	built, fabric, err := woven(flags)
	if err != nil {
		return ExitError, err
	}
	assessed, err := quality.Assess(fabric, *limit)
	if err != nil {
		return ExitError, err
	}
	fmt.Fprintf(stdout, "%s\n\n", built.Describe())
	table := report.Table{
		Header:     []string{"measure", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignRight},
	}
	table.AddRow("ends", report.Int(assessed.Ends))
	table.AddRow("picks", report.Int(assessed.Picks))
	table.AddRow("float limit", report.Int(assessed.Limit))
	table.AddRow("cells with the warp raised", report.Int(assessed.WarpUp))
	table.AddRow("balance", report.Float(assessed.Balance))
	table.AddRow("face", assessed.Face)
	table.AddRow("longest warp float", report.Int(assessed.LongestWarp))
	table.AddRow("longest weft float", report.Int(assessed.LongestWeft))
	table.AddRow("floats over the limit", report.Int(len(assessed.Over)))
	table.AddRow("ends that never interlace", report.Int(len(assessed.UnboundEnds)))
	table.AddRow("picks that never interlace", report.Int(len(assessed.UnboundPicks)))
	table.AddRow("interlacements", report.Int(assessed.Interlacements))
	table.AddRow("firmness", report.Float(assessed.Firmness))
	table.AddRow("sleazy", report.Bool(assessed.Sleazy))
	table.AddRow("sound", report.Bool(assessed.Sound()))
	fmt.Fprint(stdout, table.Render())

	histogram, err := assessed.Histogram()
	if err != nil {
		return ExitError, err
	}
	lengths := make([]int, 0, len(histogram))
	for length := range histogram {
		lengths = append(lengths, length)
	}
	sort.Ints(lengths)
	block := report.Table{
		Header:     []string{"float length", "floats"},
		Alignments: []report.Alignment{report.AlignRight, report.AlignRight},
	}
	for _, length := range lengths {
		block.AddRow(report.Int(length), report.Int(histogram[length]))
	}
	fmt.Fprint(stdout, "\n"+block.Render())

	fmt.Fprintf(stdout, "\n%s", assessed.Explain())
	if worst, line, err := assessed.Worst(); err == nil {
		fmt.Fprintf(stdout, "worst: %s\n%s\n", worst.Describe(), line)
	}
	if assessed.Sound() {
		return ExitOK, nil
	}
	return ExitOK, nil
}

// runSatin reports the steps a satin can use, and builds one.
func runSatin(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("satin", flag.ContinueOnError)
	loom := set.Int("shafts", 5, "number of shafts")
	counter := set.Int("counter", 0, "step to build a satin with, or 0 for none")
	warpFaced := set.Bool("warp", false, "build a warp faced satin instead of a weft faced one")
	repeats := set.Int("repeats", 2, "repeats of the satin to weave")
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	counters, err := pattern.SatinCounters(*loom)
	if err != nil {
		return ExitError, err
	}
	fmt.Fprintf(stdout, "%d shafts\n\n", *loom)
	table := report.Table{
		Header:     []string{"step", "shares a factor with the shafts", "usable"},
		Alignments: []report.Alignment{report.AlignRight, report.AlignRight, report.AlignRight},
	}
	for step := 1; step < *loom; step++ {
		reason := "no"
		if factor := shared(step, *loom); factor > 1 {
			reason = report.Int(factor)
		}
		table.AddRow(report.Int(step), reason,
			report.Bool(pattern.IsValidSatinCounter(*loom, step) == nil))
	}
	fmt.Fprint(stdout, table.Render())
	fmt.Fprintf(stdout, "\nsteps a %d shaft satin can use: %s\n", *loom, pattern.CounterList(counters))

	if len(counters) == 0 {
		fmt.Fprintf(stdout, "\nno step works on %d shafts, so there is no %d shaft satin\n", *loom, *loom)
		return ExitEmpty, nil
	}
	if *counter == 0 {
		return ExitOK, nil
	}
	built, err := pattern.SatinDraft(*loom, *counter, *warpFaced, *repeats)
	if err != nil {
		return ExitError, err
	}
	fabric, err := cloth.Weave(built)
	if err != nil {
		return ExitError, err
	}
	fmt.Fprintf(stdout, "\n%s\n\n", built.Describe())
	if fabric.Ends <= MaxDrawnEnds && fabric.Picks <= MaxDrawnPicks {
		fmt.Fprint(stdout, fabric.Render())
	}
	assessed, err := quality.Assess(fabric, quality.DefaultLimit)
	if err != nil {
		return ExitError, err
	}
	structure, err := pattern.Classify(fabric)
	if err != nil {
		return ExitError, err
	}
	summary := report.Table{
		Header:     []string{"measure", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignRight},
	}
	summary.AddRow("tie-up", built.TieUp.String())
	empty := "none"
	if missing := built.Threading.EmptyShafts(built.Shafts); len(missing) > 0 {
		empty = report.Ints(missing)
	}
	summary.AddRow("shafts with no end on them", empty)
	summary.AddRow("repeat in ends", report.Int(mustRepeatEnds(fabric)))
	summary.AddRow("longest warp float", report.Int(assessed.LongestWarp))
	summary.AddRow("longest weft float", report.Int(assessed.LongestWeft))
	summary.AddRow("structure", structure)
	fmt.Fprint(stdout, "\n"+summary.Render())
	return ExitOK, nil
}

// shared returns the common factor of two numbers, or one when they share none.
func shared(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

// mustRepeatEnds returns the repeat of a cloth in ends, or zero if it cannot be worked out.
func mustRepeatEnds(fabric cloth.Cloth) int {
	ends, _, err := pattern.Repeat(fabric)
	if err != nil {
		return 0
	}
	return ends
}

// runTromp treadles a draft the way it is threaded and compares the two cloths.
func runTromp(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("tromp", flag.ContinueOnError)
	flags := declareDraft(set)
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	built, err := flags.resolve()
	if err != nil {
		return ExitError, err
	}
	tromped, err := built.TrompAsWrit()
	if err != nil {
		return ExitError, err
	}
	before, err := cloth.Weave(built)
	if err != nil {
		return ExitError, err
	}
	after, err := cloth.Weave(tromped)
	if err != nil {
		return ExitError, err
	}
	fmt.Fprintf(stdout, "%s\n\n", built.Describe())
	table := report.Table{
		Header:     []string{"measure", "as written", "tromped as writ"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignRight, report.AlignRight},
	}
	table.AddRow("treadling", built.Treadling.String(), tromped.Treadling.String())
	table.AddRow("picks", report.Int(before.Picks), report.Int(after.Picks))
	table.AddRow("cells with the warp raised", report.Int(before.WarpUp()), report.Int(after.WarpUp()))
	table.AddRow("balance", report.Float(before.Balance()), report.Float(after.Balance()))
	firstStructure, err := pattern.Classify(before)
	if err != nil {
		return ExitError, err
	}
	secondStructure, err := pattern.Classify(after)
	if err != nil {
		return ExitError, err
	}
	table.AddRow("structure", firstStructure, secondStructure)
	firstFirm, err := before.Firmness()
	if err != nil {
		return ExitError, err
	}
	secondFirm, err := after.Firmness()
	if err != nil {
		return ExitError, err
	}
	table.AddRow("firmness", report.Float(firstFirm), report.Float(secondFirm))
	table.AddRow("the same cloth", report.Bool(before.Equal(after)), "")
	fmt.Fprint(stdout, table.Render())
	if after.Ends <= MaxDrawnEnds && after.Picks <= MaxDrawnPicks {
		fmt.Fprint(stdout, "\n"+after.Render())
	}
	return ExitOK, nil
}

// runDrafts prints the catalogue with its checks.
func runDrafts(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("drafts", flag.ContinueOnError)
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	names, err := draft.Names()
	if err != nil {
		return ExitError, err
	}
	table := report.Table{
		Header: []string{"draft", "name", "shafts", "treadles", "ends", "picks",
			"tie-up", "structure", "warp float", "weft float", "firmness", "sound"},
		Alignments: []report.Alignment{
			report.AlignLeft, report.AlignLeft, report.AlignRight, report.AlignRight,
			report.AlignRight, report.AlignRight, report.AlignLeft, report.AlignLeft,
			report.AlignRight, report.AlignRight, report.AlignRight, report.AlignRight,
		},
	}
	for _, key := range names {
		built, err := draft.Lookup(key)
		if err != nil {
			return ExitError, err
		}
		fabric, err := cloth.Weave(built)
		if err != nil {
			return ExitError, err
		}
		assessed, err := quality.Assess(fabric, quality.DefaultLimit)
		if err != nil {
			return ExitError, err
		}
		structure, err := pattern.Classify(fabric)
		if err != nil {
			return ExitError, err
		}
		table.AddRow(key, built.Name, report.Int(built.Shafts), report.Int(built.Treadles),
			report.Int(built.Ends()), report.Int(built.Picks()), built.TieUp.String(),
			structure, report.Int(assessed.LongestWarp), report.Int(assessed.LongestWeft),
			report.Float(assessed.Firmness), report.Bool(assessed.Sound()))
	}
	fmt.Fprint(stdout, table.Render())
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "every draft above was woven out by this tool and read for structure and for faults")
	return ExitOK, nil
}

// runReport prints one figure per package for a draft.
func runReport(args []string, stdout, stderr io.Writer) (int, error) {
	set := flag.NewFlagSet("report", flag.ContinueOnError)
	flags := declareDraft(set)
	if code, ok := parse(set, args, stdout, stderr); !ok {
		return code, nil
	}
	built, fabric, err := woven(flags)
	if err != nil {
		return ExitError, err
	}
	table := report.Table{
		Header:     []string{"package", "check", "value"},
		Alignments: []report.Alignment{report.AlignLeft, report.AlignLeft, report.AlignRight},
	}
	lifted, err := built.Lifted(1)
	if err != nil {
		return ExitError, err
	}
	table.AddRow("shafts", "shafts lifted by the first pick", lifted.Compact())
	table.AddRow("shafts", "how many that is", report.Int(lifted.Count()))

	table.AddRow("draft", "ends", report.Int(built.Ends()))
	table.AddRow("draft", "threading repeat", report.Int(built.Threading.Repeat()))

	table.AddRow("cloth", "cells with the warp raised", report.Int(fabric.WarpUp()))
	firmness, err := fabric.Firmness()
	if err != nil {
		return ExitError, err
	}
	table.AddRow("cloth", "firmness", report.Float(firmness))

	structure, err := pattern.Classify(fabric)
	if err != nil {
		return ExitError, err
	}
	repeatEnds, repeatPicks, err := pattern.Repeat(fabric)
	if err != nil {
		return ExitError, err
	}
	table.AddRow("pattern", "structure", structure)
	table.AddRow("pattern", "repeat", fmt.Sprintf("%d by %d", repeatEnds, repeatPicks))

	assessed, err := quality.Assess(fabric, quality.DefaultLimit)
	if err != nil {
		return ExitError, err
	}
	table.AddRow("quality", "longest float", report.Int(max(assessed.LongestWarp, assessed.LongestWeft)))
	table.AddRow("quality", "the cloth is sound", report.Bool(assessed.Sound()))
	fmt.Fprint(stdout, table.Render())
	return ExitOK, nil
}

// max returns the larger of two numbers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
