package wallet

import "testing"

func TestCommission(t *testing.T) {
	cases := []struct {
		bps    int64
		amount int64
		want   int64
	}{
		{1000, 10_000_00, 1_000_00}, // 10% of 10 000.00
		{500, 3_333_33, 16_666},     // 5%, truncated
		{0, 5_000_00, 0},
		{10000, 1_234_56, 1_234_56}, // 100%
	}
	for _, c := range cases {
		s := NewService(nil, c.bps)
		if got := s.Commission(c.amount); got != c.want {
			t.Errorf("Commission(bps=%d, amount=%d) = %d, want %d", c.bps, c.amount, got, c.want)
		}
	}
}
