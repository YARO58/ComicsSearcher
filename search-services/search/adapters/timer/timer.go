package timer

import (
	"context"
	"time"

	"yadro.com/course/search/core"
)

type Timer struct {
	seconds int
	updater core.UpdateIndex
}

func New(seconds int, updater core.UpdateIndex) *Timer {
	return &Timer{
		seconds: seconds,
		updater: updater,
	}
}

func (t *Timer) Start(ctx context.Context) {
	t.updater.UpdateIndex(ctx)
	for {
		select {
		case <-time.After(time.Second * time.Duration(t.seconds)):
			t.updater.UpdateIndex(ctx)
		case <-ctx.Done():
			return
		}
	}
}
