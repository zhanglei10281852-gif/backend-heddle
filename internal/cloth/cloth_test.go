package cloth

import (
	"strings"
	"testing"

	"Heddle/internal/draft"
)

// mustWeave weaves a catalogue draft or fails the test.
func mustWeave(t *testing.T, key string) Cloth {
	t.Helper()
	built, err := draft.Lookup(key)
	if err != nil {
		t.Fatalf("draft.Lookup(%s): %v", key, err)
	}
	out, err := Weave(built)
	if err != nil {
		t.Fatalf("Weave(%s): %v", key, err)
	}
	return out
}

// fromRows builds a cloth from drawn rows, x for the warp raised.
func fromRows(t *testing.T, rows ...string) Cloth {
	t.Helper()
	if len(rows) == 0 {
		t.Fatalf("a cloth needs at least one pick")
	}
	cells := []bool{}
	for _, row := range rows {
		if len(row) != len(rows[0]) {
			t.Fatalf("the rows are of different widths: %q and %q", rows[0], row)
		}
		for index := 0; index < len(row); index++ {
			cells = append(cells, row[index] == 'x')
		}
	}
	out, err := New(len(rows[0]), len(rows), cells)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return out
}

func TestWeavePlainWeave(t *testing.T) {
	fabric := mustWeave(t, "plain-weave")
	if fabric.Ends != 8 || fabric.Picks != 8 {
		t.Fatalf("the cloth is %d end(s) by %d pick(s)", fabric.Ends, fabric.Picks)
	}
	if fabric.Cells() != 64 {
		t.Fatalf("Cells = %d, want 64", fabric.Cells())
	}
	// Half the cells are warp up, so the cloth is balanced.
	if fabric.WarpUp() != 32 {
		t.Fatalf("WarpUp = %d, want 32", fabric.WarpUp())
	}
	if fabric.Balance() != 0.5 {
		t.Fatalf("Balance = %g, want 0.5", fabric.Balance())
	}
	if got := fabric.Face(); got != "balanced" {
		t.Fatalf("Face = %q, want balanced", got)
	}
	// The first pick raises the odd ends and the second the even ones.
	if got := strings.Split(fabric.Render(), "\n")[0]; got != "x.x.x.x." {
		t.Fatalf("the first pick draws as %q", got)
	}
	if got := strings.Split(fabric.Render(), "\n")[1]; got != ".x.x.x.x" {
		t.Fatalf("the second pick draws as %q", got)
	}
}

func TestWeaveTwill(t *testing.T) {
	fabric := mustWeave(t, "twill-2-2")
	lines := strings.Split(strings.TrimRight(fabric.Render(), "\n"), "\n")
	want := []string{
		"xx..xx..", ".xx..xx.", "..xx..xx", "x..xx..x",
		"xx..xx..", ".xx..xx.", "..xx..xx", "x..xx..x",
	}
	if len(lines) != len(want) {
		t.Fatalf("the cloth draws as %d line(s)", len(lines))
	}
	for index := range want {
		if lines[index] != want[index] {
			t.Fatalf("pick %d draws as %q, want %q", index+1, lines[index], want[index])
		}
	}
	// A 2/2 twill is balanced, and every pick raises half the ends.
	if fabric.Balance() != 0.5 {
		t.Fatalf("Balance = %g", fabric.Balance())
	}
	for pick := 1; pick <= fabric.Picks; pick++ {
		row, err := fabric.Row(pick)
		if err != nil {
			t.Fatalf("Row: %v", err)
		}
		raised := 0
		for _, up := range row {
			if up {
				raised++
			}
		}
		if raised != 4 {
			t.Fatalf("pick %d raises %d end(s), want 4", pick, raised)
		}
	}
}

func TestWeaveRejectsABadDraft(t *testing.T) {
	built, err := draft.Lookup("twill-2-2")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	broken := built.Clone()
	broken.Threading[0] = 9
	if _, err := Weave(broken); err == nil {
		t.Fatalf("a threading off the loom must be reported")
	}
	unnamed := built.Clone()
	unnamed.Name = ""
	if _, err := Weave(unnamed); err == nil {
		t.Fatalf("a draft with no name must be reported")
	}
}

func TestNewAndValidate(t *testing.T) {
	fabric := fromRows(t, "x.", ".x")
	if fabric.Ends != 2 || fabric.Picks != 2 {
		t.Fatalf("the cloth is %d by %d", fabric.Ends, fabric.Picks)
	}
	if err := fabric.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	// The cells belong to the cloth.
	cells := []bool{true, false, false, true}
	built, err := New(2, 2, cells)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	cells[0] = false
	if up, err := built.At(1, 1); err != nil || !up {
		t.Fatalf("writing to the cells changed the cloth: %v %v", up, err)
	}
	for label, item := range map[string]struct {
		ends, picks int
		cells       []bool
	}{
		"no ends":        {0, 2, []bool{}},
		"no picks":       {2, 0, []bool{}},
		"too few cells":  {2, 2, []bool{true}},
		"too many cells": {2, 2, []bool{true, false, true, false, true}},
	} {
		if _, err := New(item.ends, item.picks, item.cells); err == nil {
			t.Fatalf("%s: New = nil error, want a failure", label)
		}
	}
	if err := (Cloth{Ends: 2, Picks: 2}).Validate(); err == nil {
		t.Fatalf("a cloth with no cells must be reported")
	}
	if err := (Cloth{}).Validate(); err == nil {
		t.Fatalf("an empty cloth must be reported")
	}
}

func TestAtRowAndColumn(t *testing.T) {
	fabric := fromRows(t, "xx..", "..xx")
	for _, item := range []struct {
		end, pick int
		want      bool
	}{
		{1, 1, true}, {2, 1, true}, {3, 1, false}, {4, 1, false},
		{1, 2, false}, {3, 2, true},
	} {
		got, err := fabric.At(item.end, item.pick)
		if err != nil {
			t.Fatalf("At(%d, %d): %v", item.end, item.pick, err)
		}
		if got != item.want {
			t.Fatalf("At(%d, %d) = %v, want %v", item.end, item.pick, got, item.want)
		}
	}
	row, err := fabric.Row(1)
	if err != nil {
		t.Fatalf("Row: %v", err)
	}
	if len(row) != 4 || !row[0] || !row[1] || row[2] || row[3] {
		t.Fatalf("Row(1) = %v", row)
	}
	// The row belongs to the caller.
	row[0] = false
	if up, err := fabric.At(1, 1); err != nil || !up {
		t.Fatalf("writing to the row changed the cloth: %v %v", up, err)
	}
	column, err := fabric.Column(3)
	if err != nil {
		t.Fatalf("Column: %v", err)
	}
	if len(column) != 2 || column[0] || !column[1] {
		t.Fatalf("Column(3) = %v", column)
	}
	column[1] = false
	if up, err := fabric.At(3, 2); err != nil || !up {
		t.Fatalf("writing to the column changed the cloth: %v %v", up, err)
	}
	for label, item := range map[string]struct{ end, pick int }{
		"an end that is not there": {5, 1},
		"an end of zero":           {0, 1},
		"a pick that is not there": {1, 3},
		"a pick of zero":           {1, 0},
	} {
		if _, err := fabric.At(item.end, item.pick); err == nil {
			t.Fatalf("%s: At = nil error, want a failure", label)
		}
	}
	if _, err := fabric.Row(3); err == nil {
		t.Fatalf("a pick that is not there must be reported")
	}
	if _, err := fabric.Column(5); err == nil {
		t.Fatalf("an end that is not there must be reported")
	}
}

func TestFaceAndBalance(t *testing.T) {
	// Three cells of four warp up is warp faced.
	warp := fromRows(t, "xx", "x.")
	if warp.Balance() != 0.75 || warp.Face() != "warp faced" {
		t.Fatalf("Balance = %g, Face = %q", warp.Balance(), warp.Face())
	}
	weft := fromRows(t, "..", ".x")
	if weft.Balance() != 0.25 || weft.Face() != "weft faced" {
		t.Fatalf("Balance = %g, Face = %q", weft.Balance(), weft.Face())
	}
	// The satin drafts of the catalogue face opposite ways.
	if got := mustWeave(t, "satin-5").Face(); got != "warp faced" {
		t.Fatalf("the satin is %q", got)
	}
	if got := mustWeave(t, "sateen-5").Face(); got != "weft faced" {
		t.Fatalf("the sateen is %q", got)
	}
}

func TestEqualAndTranspose(t *testing.T) {
	fabric := fromRows(t, "xx..", "..xx")
	if !fabric.Equal(fromRows(t, "xx..", "..xx")) {
		t.Fatalf("the same cloth is not equal to itself")
	}
	if fabric.Equal(fromRows(t, "xx..", "..x.")) {
		t.Fatalf("two different cloths are equal")
	}
	if fabric.Equal(fromRows(t, "xx", "..")) {
		t.Fatalf("cloths of different sizes are equal")
	}
	// Turning the cloth over exchanges warp and weft, so what was raised is now crossed.
	turned := fabric.Transpose()
	if turned.Ends != fabric.Picks || turned.Picks != fabric.Ends {
		t.Fatalf("the turned cloth is %d by %d", turned.Ends, turned.Picks)
	}
	if turned.WarpUp() != fabric.Cells()-fabric.WarpUp() {
		t.Fatalf("the turned cloth has %d cell(s) warp up and the original had %d of %d",
			turned.WarpUp(), fabric.WarpUp(), fabric.Cells())
	}
	// Plain weave turned over is plain weave again.
	plain := mustWeave(t, "plain-weave")
	if got := plain.Transpose().WarpUp(); got != plain.WarpUp() {
		t.Fatalf("plain weave turned over has %d cell(s) warp up and had %d", got, plain.WarpUp())
	}
}

func TestFloatsRunRoundTheRepeat(t *testing.T) {
	// One end raised over picks 4 and 1 is one float of two, not two floats of one, because
	// the repeat is woven again and again.
	fabric := fromRows(t, "x", ".", ".", "x")
	warp, err := fabric.WarpFloats()
	if err != nil {
		t.Fatalf("WarpFloats: %v", err)
	}
	if len(warp) != 2 {
		t.Fatalf("the end carries %d float(s), want 2: %+v", len(warp), warp)
	}
	if Longest(warp) != 2 {
		t.Fatalf("the longest float is %d, want 2", Longest(warp))
	}
	for _, float := range warp {
		if float.Length != 2 {
			t.Fatalf("float %+v", float)
		}
		if float.Along != "warp" || float.Index != 1 {
			t.Fatalf("float %+v", float)
		}
	}
	// Three raised cells running off the end and on again are one float of three.
	longer := fromRows(t, "x", ".", "x", "x")
	floats, err := longer.WarpFloats()
	if err != nil {
		t.Fatalf("WarpFloats: %v", err)
	}
	if Longest(floats) != 3 {
		t.Fatalf("the longest float is %d, want 3: %+v", Longest(floats), floats)
	}
	if len(floats) != 2 {
		t.Fatalf("the end carries %d float(s), want 2: %+v", len(floats), floats)
	}
}

func TestFloatsOfALineThatNeverChanges(t *testing.T) {
	// An end raised over every pick is one float as long as the repeat.
	fabric := fromRows(t, "x.", "x.")
	warp, err := fabric.WarpFloats()
	if err != nil {
		t.Fatalf("WarpFloats: %v", err)
	}
	first := []Float{}
	for _, float := range warp {
		if float.Index == 1 {
			first = append(first, float)
		}
	}
	if len(first) != 1 || first[0].Length != 2 || !first[0].WarpUp {
		t.Fatalf("the raised end carries %+v", first)
	}
	second := []Float{}
	for _, float := range warp {
		if float.Index == 2 {
			second = append(second, float)
		}
	}
	if len(second) != 1 || second[0].Length != 2 || second[0].WarpUp {
		t.Fatalf("the lowered end carries %+v", second)
	}
}

func TestFloatsOfPlainWeave(t *testing.T) {
	fabric := mustWeave(t, "plain-weave")
	warp, err := fabric.WarpFloats()
	if err != nil {
		t.Fatalf("WarpFloats: %v", err)
	}
	weft, err := fabric.WeftFloats()
	if err != nil {
		t.Fatalf("WeftFloats: %v", err)
	}
	// Plain weave changes at every cell, so every float is one long and there are as many of
	// them as there are cells in each direction.
	if Longest(warp) != 1 || Longest(weft) != 1 {
		t.Fatalf("the longest floats are %d and %d", Longest(warp), Longest(weft))
	}
	if len(warp) != fabric.Cells() || len(weft) != fabric.Cells() {
		t.Fatalf("there are %d warp float(s) and %d weft float(s) for %d cell(s)",
			len(warp), len(weft), fabric.Cells())
	}
	all, err := fabric.Floats()
	if err != nil {
		t.Fatalf("Floats: %v", err)
	}
	if len(all) != len(warp)+len(weft) {
		t.Fatalf("Floats gave %d of %d", len(all), len(warp)+len(weft))
	}
	histogram := Histogram(all)
	if len(histogram) != 1 || histogram[1] != len(all) {
		t.Fatalf("Histogram = %v", histogram)
	}
	if got := Over(all, 1); len(got) != 0 {
		t.Fatalf("Over(1) gave %d float(s)", len(got))
	}
	if got := Over(all, 0); len(got) != len(all) {
		t.Fatalf("Over(0) gave %d of %d float(s)", len(got), len(all))
	}
}

func TestFloatsAddUpToTheCells(t *testing.T) {
	// Every cell of the cloth belongs to exactly one warp float and one weft float, whatever
	// the cloth is, which is the check that the floats were counted right.
	for _, key := range []string{"plain-weave", "twill-2-2", "twill-3-1", "twill-3-1-repeat",
		"basket-2-2", "herringbone", "satin-5", "sateen-5", "satin-8", "floating-end"} {
		fabric := mustWeave(t, key)
		warp, err := fabric.WarpFloats()
		if err != nil {
			t.Fatalf("%s: WarpFloats: %v", key, err)
		}
		weft, err := fabric.WeftFloats()
		if err != nil {
			t.Fatalf("%s: WeftFloats: %v", key, err)
		}
		warpCells := 0
		for _, float := range warp {
			warpCells += float.Length
		}
		weftCells := 0
		for _, float := range weft {
			weftCells += float.Length
		}
		if warpCells != fabric.Cells() {
			t.Fatalf("%s: the warp floats cover %d cell(s) of %d", key, warpCells, fabric.Cells())
		}
		if weftCells != fabric.Cells() {
			t.Fatalf("%s: the weft floats cover %d cell(s) of %d", key, weftCells, fabric.Cells())
		}
	}
}

func TestOverIsSortedWorstFirst(t *testing.T) {
	floats := []Float{
		{Along: "weft", Index: 2, Start: 1, Length: 4},
		{Along: "warp", Index: 1, Start: 1, Length: 6},
		{Along: "warp", Index: 3, Start: 2, Length: 4},
	}
	got := Over(floats, 3)
	if len(got) != 3 {
		t.Fatalf("Over gave %d float(s)", len(got))
	}
	if got[0].Length != 6 {
		t.Fatalf("Over put a float of %d first", got[0].Length)
	}
	// Floats of the same length come back warp before weft, then by index.
	if got[1].Along != "warp" || got[2].Along != "weft" {
		t.Fatalf("Over gave %+v", got)
	}
	if Longest(nil) != 0 {
		t.Fatalf("the longest of no floats is not zero")
	}
}

func TestInterlacementsAndFirmness(t *testing.T) {
	// Plain weave changes at every cell in both directions, so it interlaces twice per cell
	// and comes out at a firmness of one.
	plain := mustWeave(t, "plain-weave")
	interlacements, err := plain.Interlacements()
	if err != nil {
		t.Fatalf("Interlacements: %v", err)
	}
	if interlacements != 2*plain.Cells() {
		t.Fatalf("Interlacements = %d, want %d", interlacements, 2*plain.Cells())
	}
	firmness, err := plain.Firmness()
	if err != nil {
		t.Fatalf("Firmness: %v", err)
	}
	if firmness != 1 {
		t.Fatalf("Firmness = %g, want 1", firmness)
	}
	// A twill interlaces less, and a satin less again.
	twill := mustWeave(t, "twill-2-2")
	twillFirmness, err := twill.Firmness()
	if err != nil {
		t.Fatalf("Firmness: %v", err)
	}
	satin := mustWeave(t, "satin-8")
	satinFirmness, err := satin.Firmness()
	if err != nil {
		t.Fatalf("Firmness: %v", err)
	}
	if !(firmness > twillFirmness && twillFirmness > satinFirmness) {
		t.Fatalf("firmness went %g, %g, %g", firmness, twillFirmness, satinFirmness)
	}
	// A cloth that never changes interlaces nowhere.
	solid := fromRows(t, "xx", "xx")
	solidCount, err := solid.Interlacements()
	if err != nil {
		t.Fatalf("Interlacements: %v", err)
	}
	if solidCount != 0 {
		t.Fatalf("Interlacements = %d, want 0", solidCount)
	}
	solidFirmness, err := solid.Firmness()
	if err != nil {
		t.Fatalf("Firmness: %v", err)
	}
	if solidFirmness != 0 {
		t.Fatalf("Firmness = %g, want 0", solidFirmness)
	}
}

func TestDescribe(t *testing.T) {
	got := mustWeave(t, "twill-2-2").Describe()
	for _, fragment := range []string{"8 end(s) by 8 pick(s)", "32 of 64 cell(s) warp up", "balanced"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("Describe = %q, which is missing %q", got, fragment)
		}
	}
	float := Float{Along: "warp", Index: 3, Start: 2, Length: 4, WarpUp: true}
	if got := float.Describe(); !strings.Contains(got, "warp float of 4") ||
		!strings.Contains(got, "warp on top") {
		t.Fatalf("Describe = %q", got)
	}
}
