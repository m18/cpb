package rx

import (
	"regexp"
	"testing"
)

const (
	regularx = `(?P<greeting>\w+),\s*(?P<name>\w+)!`
	nestedrx = `(?P<greeting>\w+,\s*(?P<name>\w+)!)`
	extendrx = `(?P<greeting>\w+),.*\s+(?P<name>\w+)!`
	nonamerx = `(?P<greeting>\w+),\s*(\w+)(!)`
	simplerx = `\w+,\s*\w+!`

	regstr = "hello, world!"
	altstr = "hi, cosmos!"
	extstr = regstr + " " + altstr
	madstr = "hello, mad, mad, mad, mad world!"
)

func TestFindAllGroups(t *testing.T) {
	tests := []struct {
		desc     string
		pattern  string
		s        string
		expected []map[string]string
		ok       bool
	}{
		{
			desc:     "no match",
			pattern:  `\d+`,
			s:        regstr,
			expected: nil,
			ok:       false,
		},
		{
			desc:     "no groups",
			pattern:  `.*`,
			s:        regstr,
			expected: nil,
			ok:       true,
		},
		{
			desc:    "ignore non-groups",
			pattern: extendrx,
			s:       madstr,
			expected: []map[string]string{
				{"greeting": "hello", "name": "world"},
			},
			ok: true,
		},
		{
			desc:    "regular groups, single match",
			pattern: regularx,
			s:       regstr,
			expected: []map[string]string{
				{"greeting": "hello", "name": "world"},
			},
			ok: true,
		},
		{
			desc:    "regular groups, multiple matches",
			pattern: regularx,
			s:       extstr,
			expected: []map[string]string{
				{"greeting": "hello", "name": "world"},
				{"greeting": "hi", "name": "cosmos"},
			},
			ok: true,
		},
		{
			desc:    "nested groups, single match",
			pattern: nestedrx,
			s:       regstr,
			expected: []map[string]string{
				{"greeting": regstr, "name": "world"},
			},
			ok: true,
		},
		{
			desc:    "nested groups, multiple matches",
			pattern: nestedrx,
			s:       extstr,
			expected: []map[string]string{
				{"greeting": regstr, "name": "world"},
				{"greeting": altstr, "name": "cosmos"},
			},
			ok: true,
		},
		{
			desc:    "unnamed groups",
			pattern: nonamerx,
			s:       extstr,
			expected: []map[string]string{
				{"greeting": "hello", "2": "world", "3": "!"},
				{"greeting": "hi", "2": "cosmos", "3": "!"},
			},
			ok: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			r := regexp.MustCompile(test.pattern)
			res, ok := FindAllGroups(r, test.s)
			if ok != test.ok {
				t.Errorf("expected %t but did not get it", ok)
			}
			for i, m := range res {
				if !mapsAreEqual(test.expected[i], m) {
					t.Errorf("expected %v but got %v", test.expected[i], m)
				}
			}
		})
	}
}

func TestFindGroups(t *testing.T) {
	tests := []struct {
		desc     string
		pattern  string
		s        string
		expected map[string]string
		ok       bool
	}{
		{
			desc:     "no match",
			pattern:  `\d+`,
			s:        regstr,
			expected: nil,
			ok:       false,
		},
		{
			desc:     "no groups",
			pattern:  `.*`,
			s:        regstr,
			expected: nil,
			ok:       true,
		},
		{
			desc:    "ignore non-groups",
			pattern: extendrx,
			s:       madstr,
			expected: map[string]string{
				"greeting": "hello", "name": "world",
			},
			ok: true,
		},
		{
			desc:    "regular groups, single matching substring",
			pattern: regularx,
			s:       regstr,
			expected: map[string]string{
				"greeting": "hello", "name": "world",
			},
			ok: true,
		},
		{
			desc:    "regular groups, multiple matching substrings",
			pattern: regularx,
			s:       extstr,
			expected: map[string]string{
				"greeting": "hello", "name": "world",
			},
			ok: true,
		},
		{
			desc:    "nested groups, single matching substring",
			pattern: nestedrx,
			s:       regstr,
			expected: map[string]string{
				"greeting": regstr, "name": "world",
			},
			ok: true,
		},
		{
			desc:    "nested groups, multiple matching substrings",
			pattern: nestedrx,
			s:       extstr,
			expected: map[string]string{
				"greeting": regstr, "name": "world",
			},
			ok: true,
		},
		{
			desc:    "unnamed groups",
			pattern: nonamerx,
			s:       extstr,
			expected: map[string]string{
				"greeting": "hello", "2": "world", "3": "!",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			r := regexp.MustCompile(test.pattern)
			res, ok := FindGroups(r, test.s)
			if ok != test.ok {
				t.Errorf("expected %t but did not get it", ok)
			}
			if !mapsAreEqual(test.expected, res) {
				t.Errorf("expected %v but got %v", test.expected, res)
			}
		})
	}
}

func TestFindAllMatches(t *testing.T) {
	tests := []struct {
		desc     string
		pattern  string
		s        string
		expected []string
		ok       bool
	}{
		{
			desc:     "no match",
			pattern:  `\d+`,
			s:        regstr,
			expected: nil,
			ok:       false,
		},
		{
			desc:     "single match",
			pattern:  simplerx,
			s:        regstr,
			expected: []string{regstr},
			ok:       true,
		},
		{
			desc:     "single match with groups",
			pattern:  regularx,
			s:        regstr,
			expected: []string{regstr},
			ok:       true,
		},
		{
			desc:    "multiple matches",
			pattern: simplerx,
			s:       extstr,
			expected: []string{
				regstr,
				altstr,
			},
			ok: true,
		},
		{
			desc:    "multiple matches with groups",
			pattern: nestedrx,
			s:       extstr,
			expected: []string{
				regstr,
				altstr,
			},
			ok: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			r := regexp.MustCompile(test.pattern)
			res, ok := FindAllMatches(r, test.s)
			if ok != test.ok {
				t.Errorf("expected %t but did not get it", ok)
			}
			if !slicesAreEqual(test.expected, res) {
				t.Errorf("expected %v but got %v", test.expected, res)
			}
		})
	}
}

func TestFindMatch(t *testing.T) {
	tests := []struct {
		desc     string
		pattern  string
		s        string
		expected string
		ok       bool
	}{
		{
			desc:     "no match",
			pattern:  `\d+`,
			s:        regstr,
			expected: "",
			ok:       false,
		},
		{
			desc:     "single matching substring",
			pattern:  simplerx,
			s:        regstr,
			expected: regstr,
			ok:       true,
		},
		{
			desc:     "single matching substring with groups",
			pattern:  regularx,
			s:        regstr,
			expected: regstr,
			ok:       true,
		},
		{
			desc:     "multiple matching substring",
			pattern:  simplerx,
			s:        extstr,
			expected: regstr,
			ok:       true,
		},
		{
			desc:     "multiple matching substring with groups",
			pattern:  nestedrx,
			s:        extstr,
			expected: regstr,
			ok:       true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			r := regexp.MustCompile(test.pattern)
			res, ok := FindMatch(r, test.s)
			if ok != test.ok {
				t.Errorf("expected %t but did not get it", ok)
			}
			if test.expected != res {
				t.Errorf("expected %v but got %v", test.expected, res)
			}
		})
	}
}

func TestReplaceAllGroupsFunc(t *testing.T) {
	replace := func(m map[string]string) string {
		return m["name"] + ", " + m["greeting"] + "."
	}
	replaceNoName := func(m map[string]string) string {
		return m["2"] + ", " + m["greeting"] + "."
	}

	tests := []struct {
		desc     string
		pattern  string
		s        string
		replace  func(map[string]string) string
		expected string
	}{
		{
			desc:     "no match",
			pattern:  `\d+`,
			s:        regstr,
			replace:  nil,
			expected: regstr,
		},
		{
			desc:     "nil replace",
			pattern:  `.*`,
			s:        regstr,
			replace:  nil,
			expected: regstr,
		},
		{
			desc:     "no groups",
			pattern:  simplerx,
			s:        regstr,
			replace:  replace,
			expected: ", .",
		},
		{
			desc:     "regular groups, single match",
			pattern:  regularx,
			s:        regstr,
			replace:  replace,
			expected: "world, hello.",
		},
		{
			desc:     "regular groups, multiple matches",
			pattern:  regularx,
			s:        extstr,
			replace:  replace,
			expected: "world, hello. cosmos, hi.",
		},
		{
			desc:     "nested groups, single match",
			pattern:  nestedrx,
			s:        regstr,
			replace:  replace,
			expected: "world, hello, world!.",
		},
		{
			desc:     "nested groups, multiple matches",
			pattern:  nestedrx,
			s:        extstr,
			replace:  replace,
			expected: "world, hello, world!. cosmos, hi, cosmos!.",
		},
		{
			desc:     "unnamed groups, sigle match",
			pattern:  nonamerx,
			s:        regstr,
			replace:  replaceNoName,
			expected: "world, hello.",
		},
		{
			desc:     "unnamed groups, multiple matches",
			pattern:  nonamerx,
			s:        extstr,
			replace:  replaceNoName,
			expected: "world, hello. cosmos, hi.",
		},
		{
			desc:     "non-matches",
			pattern:  regularx,
			s:        "pre " + regstr + " post",
			replace:  replace,
			expected: "pre world, hello. post",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			r := regexp.MustCompile(test.pattern)
			res := ReplaceAllGroupsFunc(r, test.s, test.replace)
			if test.expected != res {
				t.Errorf("expected %v but got %v", test.expected, res)
			}
		})
	}
}

func mapsAreEqual(x, y map[string]string) bool {
	if len(x) != len(y) {
		return false
	}
	for k, v := range x {
		if y[k] != v {
			return false
		}
	}
	return true
}

func slicesAreEqual(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	for i, s := range x {
		if y[i] != s {
			return false
		}
	}
	return true
}
