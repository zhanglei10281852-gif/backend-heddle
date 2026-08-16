package draft

import (
	"fmt"
	"strconv"
	"strings"

	"Heddle/internal/shafts"
)

// ParseThreading reads a threading.
//
// Three forms are read: a list of shaft numbers separated by spaces or commas, the word
// straight followed by a number of ends, and the word point followed by a number of ends.
func ParseThreading(text string, loom int) (Threading, error) {
	fields := splitFields(text)
	if len(fields) == 0 {
		return nil, fmt.Errorf("a threading needs at least one end")
	}
	switch strings.ToLower(fields[0]) {
	case "straight", "point":
		if len(fields) != 2 {
			return nil, fmt.Errorf("a %s threading needs a number of ends", strings.ToLower(fields[0]))
		}
		ends, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("%q is not a number of ends", fields[1])
		}
		if strings.EqualFold(fields[0], "straight") {
			return Straight(loom, ends)
		}
		return Point(loom, ends)
	}
	out := make(Threading, 0, len(fields))
	for _, field := range fields {
		shaft, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("threading: %q is not a shaft number", field)
		}
		out = append(out, shaft)
	}
	if err := out.Validate(loom); err != nil {
		return nil, err
	}
	return out, nil
}

// ParseTreadling reads a treadling. The forms are the same as a threading, with the addition
// of the word tromp, which treadles the draft the way it is threaded.
func ParseTreadling(text string, treadles int, threading Threading) (Treadling, error) {
	fields := splitFields(text)
	if len(fields) == 0 {
		return nil, fmt.Errorf("a treadling needs at least one pick")
	}
	switch strings.ToLower(fields[0]) {
	case "tromp":
		if len(fields) != 1 {
			return nil, fmt.Errorf("tromp takes no further value, got %d", len(fields)-1)
		}
		if len(threading) == 0 {
			return nil, fmt.Errorf("tromp needs the threading, so it has to come after it")
		}
		out := make(Treadling, len(threading))
		for index, shaft := range threading {
			out[index] = shaft
		}
		if err := out.Validate(treadles); err != nil {
			return nil, fmt.Errorf("tromp: %w", err)
		}
		return out, nil
	case "straight", "point":
		if len(fields) != 2 {
			return nil, fmt.Errorf("a %s treadling needs a number of picks", strings.ToLower(fields[0]))
		}
		picks, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("%q is not a number of picks", fields[1])
		}
		if picks < 1 || picks > MaxPicks {
			return nil, fmt.Errorf("%d pick(s) is outside the range 1 to %d", picks, MaxPicks)
		}
		if treadles < 2 {
			return nil, fmt.Errorf("a %s treadling needs at least two treadles, and there is %d",
				strings.ToLower(fields[0]), treadles)
		}
		out := make(Treadling, picks)
		if strings.EqualFold(fields[0], "straight") {
			for index := range out {
				out[index] = index%treadles + 1
			}
		} else {
			span := 2 * (treadles - 1)
			for index := range out {
				step := index % span
				if step < treadles {
					out[index] = step + 1
					continue
				}
				out[index] = span - step + 1
			}
		}
		if err := out.Validate(treadles); err != nil {
			return nil, err
		}
		return out, nil
	}
	out := make(Treadling, 0, len(fields))
	for _, field := range fields {
		treadle, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("treadling: %q is not a treadle number", field)
		}
		out = append(out, treadle)
	}
	if err := out.Validate(treadles); err != nil {
		return nil, err
	}
	return out, nil
}

// ParseTieUp reads a tie-up: one set of shafts per treadle, separated by spaces or commas.
func ParseTieUp(text string, loom, treadles int) (TieUp, error) {
	fields := splitFields(text)
	if len(fields) == 0 {
		return nil, fmt.Errorf("a tie-up needs at least one treadle")
	}
	out := make(TieUp, 0, len(fields))
	for index, field := range fields {
		set, err := shafts.Parse(field, loom)
		if err != nil {
			return nil, fmt.Errorf("treadle %d: %w", index+1, err)
		}
		out = append(out, set)
	}
	if err := out.Validate(loom, treadles); err != nil {
		return nil, err
	}
	return out, nil
}

// splitFields splits a value on spaces and commas.
func splitFields(text string) []string {
	out := []string{}
	for _, part := range strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ','
	}) {
		field := strings.TrimSpace(part)
		if field != "" {
			out = append(out, field)
		}
	}
	return out
}

// ParseDraft reads a whole draft from the small file format this tool takes.
//
// Each line is a key and its value; a line starting with a hash is a comment and a blank line
// is skipped. The keys are name, shafts, treadles, threading, tieup and treadling. The shaft
// and treadle counts have to come before the parts they size, and the treadling has to come
// after the threading if it is written as tromp.
//
//	name 2/2 Twill
//	shafts 4
//	treadles 4
//	threading straight 8
//	tieup 12 23 34 14
//	treadling straight 8
func ParseDraft(text string) (Draft, error) {
	out := Draft{}
	seen := map[string]bool{}
	for number, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		space := strings.IndexAny(trimmed, " \t")
		if space < 0 {
			return Draft{}, fmt.Errorf("draft line %d: %q has a key and no value", number+1, trimmed)
		}
		key := strings.ToLower(strings.TrimSpace(trimmed[:space]))
		value := strings.TrimSpace(trimmed[space+1:])
		if value == "" {
			return Draft{}, fmt.Errorf("draft line %d: %q has a key and no value", number+1, trimmed)
		}
		if seen[key] {
			return Draft{}, fmt.Errorf("draft line %d: %s is given twice", number+1, key)
		}
		seen[key] = true
		switch key {
		case "name":
			out.Name = value
		case "shafts":
			count, err := strconv.Atoi(value)
			if err != nil {
				return Draft{}, fmt.Errorf("draft line %d: %q is not a number of shafts", number+1, value)
			}
			if count < 2 || count > shafts.MaxShafts {
				return Draft{}, fmt.Errorf("draft line %d: a loom of %d shaft(s) is outside the range 2 to %d",
					number+1, count, shafts.MaxShafts)
			}
			out.Shafts = count
		case "treadles":
			count, err := strconv.Atoi(value)
			if err != nil {
				return Draft{}, fmt.Errorf("draft line %d: %q is not a number of treadles", number+1, value)
			}
			if count < 1 || count > MaxTreadles {
				return Draft{}, fmt.Errorf("draft line %d: %d treadle(s) is outside the range 1 to %d",
					number+1, count, MaxTreadles)
			}
			out.Treadles = count
		case "threading":
			if out.Shafts == 0 {
				return Draft{}, fmt.Errorf("draft line %d: the shaft count has to come before the threading",
					number+1)
			}
			threading, err := ParseThreading(value, out.Shafts)
			if err != nil {
				return Draft{}, fmt.Errorf("draft line %d: %w", number+1, err)
			}
			out.Threading = threading
		case "tieup":
			if out.Shafts == 0 || out.Treadles == 0 {
				return Draft{}, fmt.Errorf("draft line %d: the shaft and treadle counts have to come before the tie-up",
					number+1)
			}
			tieup, err := ParseTieUp(value, out.Shafts, out.Treadles)
			if err != nil {
				return Draft{}, fmt.Errorf("draft line %d: %w", number+1, err)
			}
			out.TieUp = tieup
		case "treadling":
			if out.Treadles == 0 {
				return Draft{}, fmt.Errorf("draft line %d: the treadle count has to come before the treadling",
					number+1)
			}
			treadling, err := ParseTreadling(value, out.Treadles, out.Threading)
			if err != nil {
				return Draft{}, fmt.Errorf("draft line %d: %w", number+1, err)
			}
			out.Treadling = treadling
		default:
			return Draft{}, fmt.Errorf("draft line %d: %q is not one of name, shafts, treadles, threading, tieup, treadling",
				number+1, key)
		}
	}
	if out.Shafts == 0 {
		return Draft{}, fmt.Errorf("a draft needs a shaft count")
	}
	if out.Treadles == 0 {
		return Draft{}, fmt.Errorf("a draft needs a treadle count")
	}
	if len(out.Threading) == 0 {
		return Draft{}, fmt.Errorf("a draft needs a threading")
	}
	if len(out.TieUp) == 0 {
		return Draft{}, fmt.Errorf("a draft needs a tie-up")
	}
	if len(out.Treadling) == 0 {
		return Draft{}, fmt.Errorf("a draft needs a treadling")
	}
	if strings.TrimSpace(out.Name) == "" {
		out.Name = "the draft in the file"
	}
	if err := out.Validate(); err != nil {
		return Draft{}, err
	}
	return out, nil
}
