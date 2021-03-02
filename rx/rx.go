package rx

import (
	"regexp"
	"strconv"
)

// FindAllGroups returns all successive matches as a slice of maps (one per match) from matches' group names to group values.
// For unnamed groups, strings representing group indexes are used as keys.
// Non-groups are non included in the result.
func FindAllGroups(r *regexp.Regexp, s string) ([]map[string]string, bool) {
	if !r.MatchString(s) {
		return nil, false
	}
	groupNames := r.SubexpNames()[1:]
	if len(groupNames) == 0 {
		return nil, true
	}
	var res []map[string]string
	for _, matchGroups := range r.FindAllStringSubmatch(s, -1) {
		m := make(map[string]string, len(groupNames))
		for i, v := range matchGroups[1:] {
			groupName := getGroupName(groupNames, i)
			m[groupName] = v
		}
		res = append(res, m)
	}
	return res, true
}

// FindGroups returns the groups of the first match as a map from group name to group value.
// For unnamed groups, strings representing group indexes are used as keys.
// Non-groups are non included in the result.
func FindGroups(r *regexp.Regexp, s string) (map[string]string, bool) {
	if !r.MatchString(s) {
		return nil, false
	}
	groupNames := r.SubexpNames()[1:]
	if len(groupNames) == 0 {
		return nil, true
	}
	res := make(map[string]string, len(groupNames))
	for i, v := range r.FindStringSubmatch(s)[1:] {
		groupName := getGroupName(groupNames, i)
		res[groupName] = v
	}
	return res, true
}

// FindAllMatches returns all successive matches as a slice of strings.
func FindAllMatches(r *regexp.Regexp, s string) ([]string, bool) {
	if !r.MatchString(s) {
		return nil, false
	}
	return r.FindAllString(s, -1), true
}

// FindMatch returns the first match as a string.
func FindMatch(r *regexp.Regexp, s string) (string, bool) {
	if !r.MatchString(s) {
		return "", false
	}
	return r.FindString(s), true
}

// ReplaceAllGroupsFunc replaces all successive matches on the per-group basis using the provided replace callback.
func ReplaceAllGroupsFunc(r *regexp.Regexp, s string, replace func(map[string]string) string) string {
	if !r.MatchString(s) || replace == nil {
		return s
	}
	res := ""
	lastIdx := 0
	groupNames := r.SubexpNames()[1:]

	for _, idxs := range r.FindAllSubmatchIndex([]byte(s), -1) {
		groups := map[string]string{}
		for i := 2; i < len(idxs); i += 2 { // skip the overall match
			groupIdx := (i - 2) / 2
			groupName := getGroupName(groupNames, groupIdx)
			if idxs[i] == -1 || idxs[i+1] == -1 { // optional group is missing
				groups[groupName] = ""
			} else {
				groups[groupName] = s[idxs[i]:idxs[i+1]]
			}
		}
		res += s[lastIdx:idxs[0]] + replace(groups)
		lastIdx = idxs[1]
	}

	return res + s[lastIdx:]
}

func getGroupName(groupNames []string, idx int) string {
	res := groupNames[idx]
	if len(res) == 0 { // unnamed group
		res = strconv.Itoa(idx + 1)
	}
	return res
}
