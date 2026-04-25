package scheduler

import (
	"fmt"
	"time"
)

const MinCronInterval = time.Minute

func ValidateCronSchedule(cronExpr string) error {
	schedule, err := StandardParser.Parse(cronExpr)
	if err != nil {
		return err
	}

	now := time.Now()
	var lastRun time.Time
	for i := 0; i < 5; i++ {
		nextRun := schedule.Next(now)
		if i > 0 && !lastRun.IsZero() {
			interval := nextRun.Sub(lastRun)
			if interval < MinCronInterval {
				return &CronIntervalError{
					Expression: cronExpr,
					Interval:   interval,
					MinAllowed: MinCronInterval,
				}
			}
		}
		lastRun = nextRun
		now = nextRun
	}

	return nil
}

type CronIntervalError struct {
	Expression string
	Interval   time.Duration
	MinAllowed time.Duration
}

func (e *CronIntervalError) Error() string {
	return fmt.Sprintf("cron schedule runs too frequently: %s would run every %s, minimum allowed is %s",
		e.Expression, e.Interval, e.MinAllowed)
}
