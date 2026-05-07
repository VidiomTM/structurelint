package structure

import "testing"

func TestCountLeadingParents(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"./foo", 0},
		{"../foo", 1},
		{"../../foo", 2},
		{"../../../foo", 3},
		{"foo/bar", 0},
		{"./../foo", 1},
	}
	for _, c := range cases {
		if got := countLeadingParents(c.in); got != c.want {
			t.Errorf("countLeadingParents(%q) = %d, want %d", c.in, got, c.want)
		}
	}
	}
}
