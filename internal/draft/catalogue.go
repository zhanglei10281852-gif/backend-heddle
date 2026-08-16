package draft

import (
	"fmt"
	"strings"
)

// Catalogue returns the drafts this tool knows, keyed by name.
//
// The list is short on purpose. Nothing about the cloth any of them weaves is written down
// here: the tie-ups are the only thing given, and every figure the drafts command prints is
// worked out by weaving the draft out and reading the result.
func Catalogue() (map[string]Draft, error) {
	out := map[string]Draft{}
	for _, item := range []struct {
		key, name string
		shafts    int
		treadles  int
		threading string
		tieup     string
		treadling string
	}{
		// Plain weave on two shafts: the ends alternate and so do the picks.
		{"plain-weave", "Plain Weave", 2, 2, "straight 8", "1 2", "straight 8"},
		// The same cloth threaded over four shafts, which is how it is usually threaded when
		// something else is to be woven on the same warp.
		{"plain-weave-4", "Plain Weave on Four", 4, 2, "straight 8", "13 24", "straight 8"},
		// Two ends and two picks together in each shed, which gives a basket.
		{"basket-2-2", "Basket 2/2", 2, 2, "1 1 2 2 1 1 2 2", "1 2", "1 1 2 2 1 1 2 2"},
		// The four twills of a four shaft straight draw, from weft faced to warp faced.
		{"twill-1-3", "1/3 Twill", 4, 4, "straight 8", "1 2 3 4", "straight 8"},
		{"twill-2-2", "2/2 Twill", 4, 4, "straight 8", "12 23 34 14", "straight 8"},
		{"twill-3-1", "3/1 Twill", 4, 4, "straight 8", "123 234 134 124", "straight 8"},
		// One repeat of the same twill, which is the smallest block that shows the floats
		// running round the edge of the repeat rather than along the piece.
		{"twill-3-1-repeat", "3/1 Twill, One Repeat", 4, 4, "straight 4", "123 234 134 124", "straight 4"},
		// The same tie-up over a point draw, which turns the diagonal back on itself.
		{"herringbone", "Herringbone", 4, 4, "point 12", "12 23 34 14", "straight 8"},
		// A five shaft sateen: one shaft up per pick, stepped by two.
		{"sateen-5", "Sateen 5", 5, 5, "straight 10", "1 3 5 2 4", "straight 10"},
		// A five shaft satin: four shafts up per pick, the same step.
		{"satin-5", "Satin 5", 5, 5, "straight 10", "2345 1245 1234 1345 1235", "straight 10"},
		// An eight shaft satin stepped by three, which is the smallest step an eight shaft
		// satin can use.
		{"satin-8", "Satin 8", 8, 8, "straight 16", "1 4 7 2 5 8 3 6", "straight 16"},
		// A tie-up in which shaft one is tied to every treadle, so the ends on shaft one are
		// raised over every pick and never woven in at all. It is kept here because it is the
		// fault the quality report exists to find.
		{"floating-end", "Floating End", 4, 4, "straight 8", "12 13 14 1", "straight 8"},
	} {
		built := Draft{Name: item.name, Shafts: item.shafts, Treadles: item.treadles}
		threading, err := ParseThreading(item.threading, item.shafts)
		if err != nil {
			return nil, fmt.Errorf("draft %s: %w", item.key, err)
		}
		built.Threading = threading
		tieup, err := ParseTieUp(item.tieup, item.shafts, item.treadles)
		if err != nil {
			return nil, fmt.Errorf("draft %s: %w", item.key, err)
		}
		built.TieUp = tieup
		treadling, err := ParseTreadling(item.treadling, item.treadles, threading)
		if err != nil {
			return nil, fmt.Errorf("draft %s: %w", item.key, err)
		}
		built.Treadling = treadling
		if err := built.Validate(); err != nil {
			return nil, fmt.Errorf("draft %s: %w", item.key, err)
		}
		out[item.key] = built
	}
	return out, nil
}

// Names returns the catalogue keys in order.
func Names() ([]string, error) {
	catalogue, err := Catalogue()
	if err != nil {
		return nil, err
	}
	return sortedNames(catalogue), nil
}

// Lookup finds a draft in the catalogue.
func Lookup(key string) (Draft, error) {
	catalogue, err := Catalogue()
	if err != nil {
		return Draft{}, err
	}
	built, known := catalogue[strings.ToLower(strings.TrimSpace(key))]
	if !known {
		names, err := Names()
		if err != nil {
			return Draft{}, err
		}
		return Draft{}, fmt.Errorf("no draft called %q; the ones this tool knows are %s",
			key, strings.Join(names, ", "))
	}
	return built, nil
}
