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
		alg      algorithm
		attempts int
		floor    time.Duration
		ceil     time.Duration 
		fn       func() error
		want     error
	}{
		{
			"normale usage",
			ExponentialBackoff,
			3,
			slotMillis*2 + slotMillis*4,
			slotMillis*2 + slotMillis*4 + jitterRange*2,
			func() error {
				return fail
			},
			fail,
		},
		{
			"NegativeAttempts",
			ExponentialBackoff,
			-1,
			0,
			10,
			func() error { return nil },
			errors.New("retry: Do only accepts non negative attempts."),
		},
		{
			"ZeroAttempts",
			ExponentialBackoff,
			0,
			0,
			10,
			func() error { return nil },
			ZeroAttempts,
		},
		{
			"OneAttempt",
			ExponentialBackoff,
			1,
			0,
			10,
			func() error { return nil },
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
