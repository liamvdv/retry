package retry

import (
	"errors"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	fail := errors.New("fail")

	cases := []struct {
		name     string
		attempts int
		alg      algorithm
		fn       func() error

		floor    time.Duration
		ceil     time.Duration 
		want     error
	}{
		{
			"NormalUsage",
			3,
			ExponentialBackoff,
			func() error {
				return fail
			},

			slotMillis*2 + slotMillis*4,
			slotMillis*2 + slotMillis*4 + jitterRange*2,
			fail,
		},
		{
			"NegativeAttempts",
			-1,
			ExponentialBackoff,
			func() error { return nil },

			0,
			10,
			errors.New("retry: Do only accepts non negative attempts."),
		},
		{
			"ZeroAttempts",
			0,
			ExponentialBackoff,
			func() error { return nil },

			0,
			10,
			ZeroAttempts,
		},
		{
			"OneAttempt",
			1,
			ExponentialBackoff,
			func() error { return nil },

			0,
			10,
			nil,
		},
	}
	for _, c := range cases {
		var got error
		start := time.Now()
		got = Do(c.attempts, c.alg, c.fn)
		elapsed := time.Since(start)
		if !errors.Is(got, c.want) {
			if got != nil && c.want != nil {
				if got.Error() != c.want.Error() {
					t.Errorf("%s: Wanted: %q Got: %q Elapsed: %v\n", c.name, c.want, got, elapsed)
				}
			} else {
				t.Errorf("%s: Wanted: %q Got: %q Elapsed: %v\n", c.name, c.want, got, elapsed)
			} 
		}
		if !(elapsed >= c.floor*time.Millisecond && c.ceil*time.Millisecond >= elapsed) {
			t.Errorf("%s: Wanted floor: %v ceil: %v Got: %v\n", c.name, c.floor, c.ceil, elapsed)
		}
	}
}
