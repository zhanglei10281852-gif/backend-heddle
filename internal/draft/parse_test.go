package draft

import (
	"strings"
	"testing"
)

func TestParseThreading(t *testing.T) {
	got, err := ParseThreading("1 2 3 4 1 2", 4)
	if err != nil {
		t.Fatalf("ParseThreading: %v", err)
	}
	same(t, got, []int{1, 2, 3, 4, 1, 2}, "ParseThreading")
	// Commas separate the ends as well as spaces.
	commas, err := ParseThreading("1,2,3,4", 4)
	if err != nil {
		t.Fatalf("ParseThreading: %v", err)
	}
	same(t, commas, []int{1, 2, 3, 4}, "ParseThreading with commas")
	// The word straight gives a straight draw, and point a point draw.
	straight, err := ParseThreading("straight 6", 4)
	if err != nil {
		t.Fatalf("ParseThreading: %v", err)
	}
	same(t, straight, []int{1, 2, 3, 4, 1, 2}, "a straight draw")
	point, err := ParseThreading("POINT 6", 4)
	if err != nil {
		t.Fatalf("ParseThreading: %v", err)
	}
	same(t, point, []int{1, 2, 3, 4, 3, 2}, "a point draw")
	for label, item := range map[string]struct {
		text string
		loom int
	}{
		"nothing":                {"", 4},
		"only space":             {"   ", 4},
		"a letter":               {"1 2 a", 4},
		"a shaft off the loom":   {"1 2 5", 4},
		"a shaft of zero":        {"1 0", 4},
		"straight with no count": {"straight", 4},
		"straight with two":      {"straight 4 4", 4},
		"straight with a word":   {"straight many", 4},
		"point with no count":    {"point", 4},
		"a loom of one":          {"1", 1},
	} {
		if _, err := ParseThreading(item.text, item.loom); err == nil {
			t.Fatalf("%s: ParseThreading(%q, %d) = nil error, want a failure",
				label, item.text, item.loom)
		}
	}
}

func TestParseTreadling(t *testing.T) {
	threading := Threading{1, 2, 3, 4, 1, 2}
	got, err := ParseTreadling("1 2 3 4", 4, threading)
	if err != nil {
		t.Fatalf("ParseTreadling: %v", err)
	}
	same(t, got, []int{1, 2, 3, 4}, "ParseTreadling")
	straight, err := ParseTreadling("straight 6", 4, threading)
	if err != nil {
		t.Fatalf("ParseTreadling: %v", err)
	}
	same(t, straight, []int{1, 2, 3, 4, 1, 2}, "a straight treadling")
	point, err := ParseTreadling("point 6", 4, threading)
	if err != nil {
		t.Fatalf("ParseTreadling: %v", err)
	}
	same(t, point, []int{1, 2, 3, 4, 3, 2}, "a point treadling")
	// Tromp reads the treadling off the threading.
	tromp, err := ParseTreadling("tromp", 4, threading)
	if err != nil {
		t.Fatalf("ParseTreadling: %v", err)
	}
	same(t, tromp, []int{1, 2, 3, 4, 1, 2}, "a tromped treadling")
	for label, item := range map[string]struct {
		text      string
		treadles  int
		threading Threading
	}{
		"nothing":                     {"", 4, threading},
		"a letter":                    {"1 b", 4, threading},
		"a treadle that is not there": {"1 5", 4, threading},
		"tromp with no threading":     {"tromp", 4, nil},
		"tromp with a value":          {"tromp 4", 4, threading},
		"tromp onto too few treadles": {"tromp", 2, threading},
		"straight with no count":      {"straight", 4, threading},
		"straight onto one treadle":   {"straight 4", 1, threading},
		"straight with no picks":      {"straight 0", 4, threading},
		"point with no count":         {"point", 4, threading},
	} {
		if _, err := ParseTreadling(item.text, item.treadles, item.threading); err == nil {
			t.Fatalf("%s: ParseTreadling(%q) = nil error, want a failure", label, item.text)
		}
	}
}

func TestParseTieUp(t *testing.T) {
	got, err := ParseTieUp("12 23 34 14", 4, 4)
	if err != nil {
		t.Fatalf("ParseTieUp: %v", err)
	}
	if got.Treadles() != 4 {
		t.Fatalf("ParseTieUp gave %d treadle(s)", got.Treadles())
	}
	if text := got.String(); text != "12 23 34 14" {
		t.Fatalf("the tie-up reads back as %q", text)
	}
	// A tie-up on a wide loom needs full stops between the shafts.
	wide, err := ParseTieUp("1.10 2.11", 12, 2)
	if err != nil {
		t.Fatalf("ParseTieUp: %v", err)
	}
	if text := wide.String(); text != "1.10 2.11" {
		t.Fatalf("the wide tie-up reads back as %q", text)
	}
	for label, item := range map[string]struct {
		text     string
		loom     int
		treadles int
	}{
		"nothing":                  {"", 4, 4},
		"a shaft off the loom":     {"12 25", 4, 2},
		"a treadle that lifts all": {"1234 12", 4, 2},
		"too few treadles":         {"12 23", 4, 4},
		"too many treadles":        {"12 23 34 14", 4, 2},
		"a letter":                 {"1x", 4, 1},
	} {
		if _, err := ParseTieUp(item.text, item.loom, item.treadles); err == nil {
			t.Fatalf("%s: ParseTieUp(%q) = nil error, want a failure", label, item.text)
		}
	}
}

func TestParseDraft(t *testing.T) {
	built, err := ParseDraft(strings.Join([]string{
		"# a draft",
		"",
		"name 2/2 Twill",
		"shafts 4",
		"treadles 4",
		"threading straight 8",
		"tieup 12 23 34 14",
		"treadling tromp",
	}, "\n"))
	if err != nil {
		t.Fatalf("ParseDraft: %v", err)
	}
	if built.Name != "2/2 Twill" || built.Shafts != 4 || built.Treadles != 4 {
		t.Fatalf("ParseDraft = %+v", built)
	}
	same(t, built.Threading, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the threading")
	same(t, built.Treadling, []int{1, 2, 3, 4, 1, 2, 3, 4}, "the treadling")
	if text := built.TieUp.String(); text != "12 23 34 14" {
		t.Fatalf("the tie-up read as %q", text)
	}
	if err := built.Validate(); err != nil {
		t.Fatalf("the draft in the file is not a draft: %v", err)
	}
	// The keys are read in either case, and a draft with no name still comes back usable.
	upper, err := ParseDraft("SHAFTS 2\nTREADLES 2\nTHREADING straight 4\nTIEUP 1 2\nTREADLING straight 4\n")
	if err != nil {
		t.Fatalf("ParseDraft: %v", err)
	}
	if strings.TrimSpace(upper.Name) == "" {
		t.Fatalf("a draft with no name came back with no name")
	}
	if upper.Shafts != 2 {
		t.Fatalf("ParseDraft = %+v", upper)
	}
}

func TestParseDraftRejectsBadFiles(t *testing.T) {
	sound := []string{
		"shafts 4", "treadles 4", "threading straight 8",
		"tieup 12 23 34 14", "treadling straight 8",
	}
	for label, text := range map[string]string{
		"nothing":             "",
		"only a comment":      "# a draft\n",
		"a key with no value": "shafts\n",
		"an empty value":      "shafts \n",
		"an unknown key":      strings.Join(append(sound, "sett 24"), "\n"),
		"a key given twice":   strings.Join(append(sound, "shafts 8"), "\n"),
		"no shafts": strings.Join([]string{
			"treadles 4", "threading 1 2", "tieup 12", "treadling 1"}, "\n"),
		"no treadles": strings.Join([]string{
			"shafts 4", "threading straight 8", "tieup 12", "treadling 1"}, "\n"),
		"no threading": strings.Join([]string{
			"shafts 4", "treadles 4", "tieup 12 23 34 14", "treadling straight 8"}, "\n"),
		"no tie-up": strings.Join([]string{
			"shafts 4", "treadles 4", "threading straight 8", "treadling straight 8"}, "\n"),
		"no treadling": strings.Join([]string{
			"shafts 4", "treadles 4", "threading straight 8", "tieup 12 23 34 14"}, "\n"),
		"a threading before the shafts": strings.Join([]string{
			"threading straight 8", "shafts 4", "treadles 4",
			"tieup 12 23 34 14", "treadling straight 8"}, "\n"),
		"a tie-up before the treadles": strings.Join([]string{
			"shafts 4", "tieup 12 23 34 14", "treadles 4",
			"threading straight 8", "treadling straight 8"}, "\n"),
		"a treadling before the treadles": strings.Join([]string{
			"shafts 4", "treadling straight 8", "treadles 4",
			"threading straight 8", "tieup 12 23 34 14"}, "\n"),
		"a shaft count that is not a number": "shafts four\ntreadles 4\n",
		"a shaft count out of range":         "shafts 40\ntreadles 4\n",
		"a treadle count out of range":       "shafts 4\ntreadles 40\n",
		"a tie-up that opens no shed": strings.Join([]string{
			"shafts 4", "treadles 4", "threading straight 8",
			"tieup 1234 23 34 14", "treadling straight 8"}, "\n"),
		"a threading off the loom": strings.Join([]string{
			"shafts 4", "treadles 4", "threading 1 2 3 9",
			"tieup 12 23 34 14", "treadling straight 8"}, "\n"),
	} {
		if _, err := ParseDraft(text); err == nil {
			t.Fatalf("%s: ParseDraft = nil error, want a failure", label)
		}
	}
}

func TestParseDraftReadsTheCatalogueBack(t *testing.T) {
	// Every draft in the catalogue can be written out in the file format and read back to the
	// same draft, which is what makes the format worth having.
	names, err := Names()
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	for _, key := range names {
		built, err := Lookup(key)
		if err != nil {
			t.Fatalf("Lookup(%s): %v", key, err)
		}
		text := strings.Join([]string{
			"name " + built.Name,
			"shafts " + itoa(built.Shafts),
			"treadles " + itoa(built.Treadles),
			"threading " + built.Threading.String(),
			"tieup " + built.TieUp.String(),
			"treadling " + built.Treadling.String(),
		}, "\n")
		back, err := ParseDraft(text)
		if err != nil {
			t.Fatalf("%s read back: %v", key, err)
		}
		if back.Shafts != built.Shafts || back.Treadles != built.Treadles {
			t.Fatalf("%s read back as %+v", key, back)
		}
		same(t, back.Threading, ints(built.Threading), key+" threading")
		same(t, back.Treadling, ints(built.Treadling), key+" treadling")
		if back.TieUp.String() != built.TieUp.String() {
			t.Fatalf("%s tie-up read back as %q", key, back.TieUp.String())
		}
	}
}

// itoa renders a small number.
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// ints copies a sequence into a plain slice.
func ints[T ~int](values []T) []int {
	out := make([]int, len(values))
	for index, value := range values {
		out[index] = int(value)
	}
	return out
}
