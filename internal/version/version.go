// Package version compares mobile app versions ("1.2.3" or "1.2.3+7") and
// decides the update the app must show.
package version

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	Android = "android"
	IOS     = "ios"

	UpdateNone     = "none"     // the app is up to date
	UpdateOptional = "optional" // a newer version exists; the old one still works
	UpdateForce    = "force"    // older than min_version: the app must update
)

var versionRe = regexp.MustCompile(`^\d+(\.\d+){0,2}(\+\d+)?$`)

func IsValidPlatform(p string) bool { return p == Android || p == IOS }

// IsValid reports whether v looks like "1", "1.2", "1.2.3" or "1.2.3+7".
func IsValid(v string) bool { return versionRe.MatchString(v) }

func parse(v string) []int {
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == '.' || r == '+' })
	out := make([]int, len(parts))
	for i, p := range parts {
		out[i], _ = strconv.Atoi(p)
	}
	return out
}

// IsLower reports whether current is older than target. Only the parts that
// target has are compared: a target of "1.0.0" ignores the build number of
// current, a target of "1.0.0+5" compares it too.
func IsLower(current, target string) bool {
	c, t := parse(current), parse(target)
	for i, tv := range t {
		cv := 0
		if i < len(c) {
			cv = c[i]
		}
		if cv != tv {
			return cv < tv
		}
	}
	return false
}

// UpdateType decides what the app with version current must do.
func UpdateType(current, minVersion, latestVersion string) string {
	switch {
	case IsLower(current, minVersion):
		return UpdateForce
	case IsLower(current, latestVersion):
		return UpdateOptional
	default:
		return UpdateNone
	}
}
