package minions

import (
	"context"
	"sync"
)

// Finisher provides a simple way to notify about individual works finished.
type Finisher struct {
	done chan struct{}
	once sync.Once
}

func (f *Finisher) init() {
	f.once.Do(func() {
		f.done = make(chan struct{})
	})
}

// Finish must be called only once when the job is done
func (f *Finisher) Finish() {
	f.init()
	close(f.done)
}

// Wait can be called to block until the job is finished. It can be aborted with the context.
// This method only returns an error if the context is cancelled
func (f *Finisher) Wait(ctx context.Context) error {
	f.init()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.done:
		return nil
	}
}
