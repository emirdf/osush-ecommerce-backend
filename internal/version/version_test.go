package version

import "testing"

func TestIsLower(t *testing.T) {
	cases := []struct {
		current, target string
		want            bool
	}{
		{"1.0.0+4", "1.0.1", true},
		{"1.0.0+4", "1.0.0", false},  // build ignored
		{"1.0.0+4", "1.0.0+5", true}, // build compared
		{"1.2.0", "1.10.0", true},    // numbers, not text
		{"2.0.0+1", "1.9.9", false},
		{"1.0", "1.0.1", true},
	}
	for _, c := range cases {
		if got := IsLower(c.current, c.target); got != c.want {
			t.Errorf("IsLower(%q, %q) = %v, want %v", c.current, c.target, got, c.want)
		}
	}
}

func TestUpdateType(t *testing.T) {
	cases := map[string]string{
		"1.0.9+4": UpdateForce,
		"1.1.0+1": UpdateOptional,
		"1.2.5":   UpdateOptional,
		"1.3.0+9": UpdateNone,
		"2.0.0":   UpdateNone,
	}
	for current, want := range cases {
		if got := UpdateType(current, "1.1.0", "1.3.0"); got != want {
			t.Errorf("UpdateType(%q) = %q, want %q", current, got, want)
		}
	}
}

func TestIsValid(t *testing.T) {
	for _, v := range []string{"1", "1.2", "1.2.3", "1.2.3+7"} {
		if !IsValid(v) {
			t.Errorf("%q should be valid", v)
		}
	}
	for _, v := range []string{"", "v1.0", "1.2.3.4", "1.2.3+", "abc", "1..2"} {
		if IsValid(v) {
			t.Errorf("%q should be invalid", v)
		}
	}
}
