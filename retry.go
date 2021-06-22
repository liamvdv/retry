package retry

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"time"
)

// parameters
const (
	slotMillis  = 100
	jitterRange = 100
)

type algorithm int

const (
	ExponentialBackoff algorithm = iota
)

var algToString = []string{
	ExponentialBackoff: "ExponentialBackoff",
}

func (alg algorithm) String() string {
	if int(alg) >= len(algToString) {
		return "<retry: algorithm not known>"
	}
	return algToString[alg]
}

var Stop = errors.New("return immediately")
var ZeroAttempts = errors.New("Do called with 0 attempts, must fail")

// Do returns the last error returned by fn after all attempts fail. If fn
// returns nil or retry.Stop, Do will exit and return nil. Do returns nil for 0
// attempts
func Do(attempts int, alg algorithm, fn func() error) error {
	if attempts < 0 {
		return errors.New("retry: Do only accepts non negative attempts.")
	}
	if attempts == 0 {
		return ZeroAttempts
	}

	now := time.Now().UnixNano()
	src := rand.NewSource(now)
	r := rand.New(src)

	wait := getWaitFunc(alg)

	var err error
	for tries := 1; tries <= attempts; tries++ {
		err = fn()
		switch err {
		case nil:
			return nil
		case Stop:
			return nil
		}
		if tries == attempts {
			return err
		}
		<-time.After(wait(tries, r.Intn(jitterRange)))
	}
	return errors.New("retry: Invalid. Open issue on github.com/liamvdv/retry")
}

/******************************* Wait functions *******************************/

// Wait should return a time.Duration. jitter must be added to the duration the
// function calculated to prevent synchronous peaks.
type Wait func(tries int, jitter int) time.Duration

var registry = []Wait{
	ExponentialBackoff: exponentialBackoff,
}

func AddAlgorithm(name string, fn Wait) algorithm {
	registry = append(registry, fn)
	algToString = append(algToString, name)

	if len(registry) != len(algToString) {
		log.Panicf("retry: function registry and algorithm index do not match up. Open issue on github.com/liamvdv/retry")
	}

	return algorithm(len(registry) - 1)
}

func getWaitFunc(alg algorithm) Wait {
	if int(alg) >= len(registry) {
		log.Panicf("retry: algorithm %d is not know.", alg)
	}

	return registry[alg]
}

func exponentialBackoff(tries int, jitterMillis int) time.Duration {
	n := int(math.Pow(2, float64(tries)))
	return time.Duration(n*slotMillis+jitterMillis) * time.Millisecond
}
