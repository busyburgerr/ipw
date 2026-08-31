package review

import "testing"

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{5, 5},
		{4.0, 4},
		{4.6666666, 4.67},
		{3.333333, 3.33},
		{4.125, 4.13},
		{0, 0},
	}
	for _, c := range cases {
		if got := round2(c.in); got != c.want {
			t.Errorf("round2(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
