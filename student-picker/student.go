package main

import (
	"regexp"
	"strings"
	"unicode"
)

// Student represents a student from SchaleDB
type Student struct {
	Id           int
	Name         string // full display name from JSON e.g. "Hina (Swimsuit)"
	FamilyName   string
	PersonalName string
	Base         string // computed: base name without variant e.g. "Hina"
	Variant      string // computed: variant string e.g. "Swimsuit" (empty if none)
	Abbr         string // computed: abbreviation e.g. "SW" (empty if no variant)
}

var variantRegex = regexp.MustCompile(`^(.+?)\s*\(([^()]+)\)\s*$`)

// computeVariant parses Name and fills Base/Variant/Abbr.
// Abbr for single-word variant = first 2 uppercase chars ("Swimsuit" → "SW").
// Abbr for multi-word variant = initials of each word ("New Year" → "NY").
func computeVariant(s *Student) {
	if s.Name == "" {
		s.Name = s.PersonalName
	}
	m := variantRegex.FindStringSubmatch(s.Name)
	if m == nil {
		s.Base = s.Name
		s.Variant = ""
		s.Abbr = ""
		return
	}
	s.Base = strings.TrimSpace(m[1])
	s.Variant = strings.TrimSpace(m[2])
	s.Abbr = makeAbbr(s.Variant)
}

// makeAbbr converts a variant string to its abbreviation.
func makeAbbr(variant string) string {
	words := strings.Fields(variant)
	if len(words) == 0 {
		return ""
	}
	if len(words) == 1 {
		// Single word: first 2 uppercase letters
		runes := []rune(variant)
		abbr := ""
		for _, r := range runes {
			if unicode.IsLetter(r) {
				abbr += string(unicode.ToUpper(r))
				if len([]rune(abbr)) >= 2 {
					break
				}
			}
		}
		return abbr
	}
	// Multi-word: first letter of each word
	abbr := ""
	for _, w := range words {
		runes := []rune(w)
		if len(runes) > 0 {
			abbr += string(unicode.ToUpper(runes[0]))
		}
	}
	return abbr
}

// displayName returns the best name to show in a grid cell of given innerWidth.
// Returns the display string and whether abbreviation was used.
func displayName(s Student, innerW int) (name string, abbrUsed bool) {
	full := s.Name
	if full == "" {
		full = s.PersonalName
	}
	if runeLen(full) <= innerW {
		return full, false
	}
	// Try abbreviated variant
	if s.Variant != "" && s.Abbr != "" {
		abbrName := s.Base + " (" + s.Abbr + ")"
		if runeLen(abbrName) <= innerW {
			return abbrName, true
		}
	}
	// Truncate
	runes := []rune(full)
	if innerW <= 1 {
		return "…", false
	}
	return string(runes[:innerW-1]) + "…", s.Variant != ""
}

func runeLen(s string) int {
	return len([]rune(s))
}