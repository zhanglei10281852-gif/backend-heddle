package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run executes one invocation and returns the exit code with both streams.
func run(args ...string) (int, string, string) {
	var stdout, stderr strings.Builder
	code := Run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// wants fails the test unless the text holds every fragment.
func wants(t *testing.T, text string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			t.Fatalf("output is missing %q:\n%s", fragment, text)
		}
	}
}

func TestRunWithoutArgumentsPrintsUsage(t *testing.T) {
	code, stdout, stderr := run()
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want nothing", stdout)
	}
	wants(t, stderr, "usage: heddle <command>")
}

func TestRunHelp(t *testing.T) {
	for _, argument := range []string{"help", "--help", "-h"} {
		code, stdout, _ := run(argument)
		if code != ExitOK {
			t.Fatalf("%s: code = %d, want %d", argument, code, ExitOK)
		}
		wants(t, stdout,
			"usage: heddle <command>",
			"study tool only",
			"exit codes: 0 produced an answer, 1 could not run, 2 ran and found nothing",
			"draft", "cloth", "floats", "pattern", "quality", "satin", "tromp",
			"drafts", "report",
			"drafts: basket-2-2",
			"x where the warp is raised",
			"this tool's own")
	}
}

func TestRunVersion(t *testing.T) {
	for _, argument := range []string{"version", "--version", "-v"} {
		code, stdout, _ := run(argument)
		if code != ExitOK {
			t.Fatalf("%s: code = %d, want %d", argument, code, ExitOK)
		}
		wants(t, stdout, Version)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	code, _, stderr := run("sley")
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	wants(t, stderr, `unknown command "sley"`, "usage: heddle <command>")
}

func TestSubcommandHelpFlag(t *testing.T) {
	code, stdout, _ := run("cloth", "-h")
	if code != ExitOK {
		t.Fatalf("code = %d, want %d", code, ExitOK)
	}
	wants(t, stdout, "cloth - weave a draft out", "-draft", "-file", "-mirror", "-tromp")
}

func TestSubcommandRejectsAnUnknownFlag(t *testing.T) {
	code, _, stderr := run("cloth", "-nope")
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	wants(t, stderr, "flag provided but not defined")
}

func TestDraft(t *testing.T) {
	code, stdout, stderr := run("draft", "-draft", "twill-2-2")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "2/2 Twill", "4 shaft(s)", "4 treadle(s)",
		"shafts", "4", "treadles", "ends", "8", "picks",
		"threading", "1 2 3 4 1 2 3 4", "threading repeat", "4",
		"tie-up", "12 23 34 14", "treadling", "treadling repeat",
		"treadle", "lifted", "grid", "xx..", ".xx.", "..xx", "x..x",
		"shaft", "lifted by", "first ends", "1 5",
		"threading from the other selvedge", "4 3 2 1 4 3 2 1",
		"reads the same from either selvedge", "no",
		"shafts with no end on them", "none",
		"treadles used", "1 2 3 4")
}

func TestDraftOfAPointDraw(t *testing.T) {
	code, stdout, stderr := run("draft", "-draft", "herringbone")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "Herringbone", "12 end(s)", "8 pick(s)",
		"1 2 3 4 3 2 1 2 3 4 3 2", "threading repeat", "6",
		"threading from the other selvedge", "2 3 4 3 2 1 2 3 4 3 2 1")
}

func TestDraftMirrored(t *testing.T) {
	code, stdout, stderr := run("draft", "-draft", "twill-2-2", "-mirror")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "2/2 Twill mirrored", "threading", "4 3 2 1 4 3 2 1",
		"threading from the other selvedge", "1 2 3 4 1 2 3 4")
}

func TestDraftRejectsBadInput(t *testing.T) {
	for _, args := range [][]string{
		{"draft", "-draft", "crackle"},
		{"draft", "-draft", "plain-weave-4", "-tromp"},
	} {
		code, _, stderr := run(args...)
		if code != ExitError {
			t.Fatalf("%v: code = %d, want %d", args, code, ExitError)
		}
		if stderr == "" {
			t.Fatalf("%v: stderr is empty, want an explanation", args)
		}
	}
}

func TestCloth(t *testing.T) {
	code, stdout, stderr := run("cloth", "-draft", "twill-2-2")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "2/2 Twill",
		"xx..xx..", ".xx..xx.", "..xx..xx", "x..xx..x",
		"ends", "8", "picks", "cells", "64",
		"cells with the warp raised", "32", "balance", "0.5", "face", "balanced",
		"repeat in ends", "4", "repeat in picks",
		"interlacements", "64", "firmness",
		"structure", "2/2 twill stepping 1 to the right")
}

func TestClothOfPlainWeave(t *testing.T) {
	code, stdout, stderr := run("cloth", "-draft", "plain-weave")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "x.x.x.x.", ".x.x.x.x", "structure", "plain weave",
		"firmness", "1", "interlacements", "128")
}

func TestClothMirroredTurnsTheDiagonalOver(t *testing.T) {
	code, stdout, stderr := run("cloth", "-draft", "twill-2-2", "-mirror")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	// Threading the same twill from the other selvedge makes the diagonal climb the other way.
	wants(t, stdout, "2/2 Twill mirrored", "..xx..xx",
		"structure", "2/2 twill stepping 1 to the left")
}

func TestClothTromped(t *testing.T) {
	code, stdout, stderr := run("cloth", "-draft", "twill-2-2", "-tromp")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "tromped as writ", "structure", "2/2 twill")
}

func TestFloats(t *testing.T) {
	code, stdout, stderr := run("floats", "-draft", "twill-2-2", "-over", "1")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "2/2 Twill",
		"floats along the warp", "floats along the weft",
		"longest warp float", "2", "longest weft float", "limit", "1",
		"length", "floats", "along", "index", "from", "length", "on top")
}

func TestFloatsFindsNothingOverTheLimit(t *testing.T) {
	code, stdout, stderr := run("floats", "-draft", "plain-weave")
	if code != ExitEmpty {
		t.Fatalf("code = %d, want %d, stderr = %q", code, ExitEmpty, stderr)
	}
	wants(t, stdout, "longest warp float", "1", "no float is longer than 3")
}

func TestFloatsRunRoundTheRepeat(t *testing.T) {
	// One repeat of a 3/1 twill floats over three, and the float runs off the last pick and on
	// again at the first, so counting it along the piece instead of round the repeat would find
	// only two.
	code, stdout, stderr := run("floats", "-draft", "twill-3-1-repeat", "-over", "2")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "3/1 Twill, One Repeat",
		"floats along the warp", "8", "floats along the weft", "8",
		"longest warp float", "3", "longest weft float", "3")
}

func TestFloatsRejectsABadLimit(t *testing.T) {
	code, _, stderr := run("floats", "-draft", "plain-weave", "-over", "0")
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	wants(t, stderr, "allows no float at all")
}

func TestPattern(t *testing.T) {
	code, stdout, stderr := run("pattern", "-draft", "plain-weave")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "Plain Weave", "repeat in ends", "2", "repeat in picks",
		"plain weave", "yes", "twill", "twill step", "1", "twill direction", "right",
		"warp over", "warp under", "satin", "no", "balance", "0.5",
		"face", "balanced", "structure", "plain weave",
		"repeating over 2 by 2")
}

func TestPatternOfASatin(t *testing.T) {
	code, stdout, stderr := run("pattern", "-draft", "satin-5")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "Satin 5", "satin", "yes", "satin step", "2", "satin shafts", "5",
		"warp faced", "5 shaft warp faced satin stepping 2")
}

func TestPatternOfABasket(t *testing.T) {
	code, stdout, stderr := run("pattern", "-draft", "basket-2-2")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "Basket 2/2", "plain weave", "no", "twill", "no", "satin", "no",
		"no plain weave, twill or satin")
}

func TestSatinOnFiveShafts(t *testing.T) {
	code, stdout, stderr := run("satin", "-shafts", "5", "-counter", "2")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "5 shafts", "step", "shares a factor with the shafts", "usable",
		"steps a 5 shaft satin can use: 2 3",
		"5 shaft weft faced satin stepping 2", "1 3 5 2 4",
		"x....x....", "shafts with no end on them", "none",
		"repeat in ends", "5", "longest warp float", "4")
}

func TestSatinOnSevenShafts(t *testing.T) {
	code, stdout, stderr := run("satin", "-shafts", "7")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	// Seven shafts are prime, so every step but the two at the ends can be used.
	wants(t, stdout, "7 shafts", "steps a 7 shaft satin can use: 2 3 4 5")
}

func TestSatinRejectsBadInput(t *testing.T) {
	for _, args := range [][]string{
		{"satin", "-shafts", "4"},
		{"satin", "-shafts", "40"},
		{"satin", "-shafts", "5", "-counter", "1"},
		{"satin", "-shafts", "5", "-counter", "4"},
		{"satin", "-shafts", "5", "-counter", "2", "-repeats", "0"},
	} {
		code, _, stderr := run(args...)
		if code != ExitError {
			t.Fatalf("%v: code = %d, want %d", args, code, ExitError)
		}
		if stderr == "" {
			t.Fatalf("%v: stderr is empty, want an explanation", args)
		}
	}
}

func TestTromp(t *testing.T) {
	code, stdout, stderr := run("tromp", "-draft", "twill-2-2")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "2/2 Twill", "measure", "as written", "tromped as writ",
		"treadling", "1 2 3 4 1 2 3 4", "picks", "8",
		"cells with the warp raised", "32", "balance", "0.5",
		"structure", "2/2 twill", "firmness", "the same cloth", "yes")
}

func TestTrompRejectsADraftItCannotTromp(t *testing.T) {
	code, _, stderr := run("tromp", "-draft", "plain-weave-4")
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	wants(t, stderr, "cannot be treadled as it is threaded")
}

func TestDrafts(t *testing.T) {
	code, stdout, stderr := run("drafts")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "draft", "name", "shafts", "treadles", "ends", "picks",
		"tie-up", "structure", "warp float", "weft float", "firmness", "sound",
		"plain-weave", "Plain Weave", "plain weave",
		"twill-2-2", "2/2 Twill", "12 23 34 14", "2/2 twill stepping 1 to the right",
		"satin-8", "1 4 7 2 5 8 3 6", "8 shaft weft faced satin stepping 3",
		"floating-end", "12 13 14 1",
		"woven out by this tool and read for structure and for faults")
}

func TestReportCoversEveryPackage(t *testing.T) {
	code, stdout, stderr := run("report", "-draft", "twill-2-2")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "package", "check", "value",
		"shafts", "shafts lifted by the first pick", "12", "how many that is", "2",
		"draft", "ends", "8", "threading repeat", "4",
		"cloth", "cells with the warp raised", "32", "firmness", "0.5",
		"pattern", "structure", "2/2 twill stepping 1 to the right", "repeat", "4 by 4",
		"quality", "longest float", "2", "the cloth is sound", "yes")
}

func TestReportOfADraftWithAFault(t *testing.T) {
	code, stdout, stderr := run("report", "-draft", "floating-end")
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "longest float", "8", "the cloth is sound", "no")
}

func TestReadingADraftFromAFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "draft.txt")
	if err := os.WriteFile(path, []byte(strings.Join([]string{
		"# a draft",
		"name Written Down",
		"shafts 4",
		"treadles 4",
		"threading straight 8",
		"tieup 12 23 34 14",
		"treadling tromp",
	}, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	code, stdout, stderr := run("cloth", "-file", path)
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "Written Down", "xx..xx..", "structure", "2/2 twill stepping 1 to the right")

	code, stdout, stderr = run("draft", "-file", path)
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	wants(t, stdout, "Written Down", "1 2 3 4 1 2 3 4", "12 23 34 14")
}

func TestReadingADraftFromAFileRejectsBadInput(t *testing.T) {
	directory := t.TempDir()
	missing := filepath.Join(directory, "absent.txt")
	code, _, stderr := run("cloth", "-file", missing)
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	wants(t, stderr, "open")

	broken := filepath.Join(directory, "broken.txt")
	if err := os.WriteFile(broken, []byte(strings.Join([]string{
		"shafts 4",
		"treadles 4",
		"threading straight 8",
		"tieup 1234 23 34 14",
		"treadling straight 8",
	}, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	code, _, stderr = run("cloth", "-file", broken)
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	wants(t, stderr, "broken.txt", "opens no shed")

	offLoom := filepath.Join(directory, "offloom.txt")
	if err := os.WriteFile(offLoom, []byte(strings.Join([]string{
		"shafts 4",
		"treadles 4",
		"threading 1 2 3 9",
		"tieup 12 23 34 14",
		"treadling straight 8",
	}, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	code, _, stderr = run("cloth", "-file", offLoom)
	if code != ExitError {
		t.Fatalf("code = %d, want %d", code, ExitError)
	}
	wants(t, stderr, "offloom.txt", "not one of the 4 shafts")
}
