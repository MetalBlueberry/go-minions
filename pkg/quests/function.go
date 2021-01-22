package quests

import "context"

// Function executes the given function
type Function func(context.Context)

// Work implements minions.Worker
func (f Function) Work(ctx context.Context) {
	f(ctx)
}
