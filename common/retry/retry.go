package retry

import (
	"time"

	"github.com/sirupsen/logrus"
)

// WithBackOff will retry a function up to `maxTimes` times,
// with an doubling backoff strategy.
func WithBackOff(f func() error, maxTimes int, wait time.Duration, desc string) error {
	err := f()
	if err == nil {
		return nil
	}

	var retErr error
	waitTime := wait
	for i := 1; i < maxTimes; i++ {
		logrus.Infof("Retry %d for '%s'", i, desc)
		time.Sleep(waitTime)
		err := f()
		if err == nil {
			return nil
		}
		retErr = err
		waitTime = 2 * waitTime
	}
	return retErr
}
