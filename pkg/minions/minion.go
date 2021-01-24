package minions

import (
	"context"
	"sync"
)

// Worker is any struct that has work to be done
type Worker interface {
	Work(context.Context)
}

// Lord coordinates working minions.
type Lord struct {
	wg *sync.WaitGroup
}

// NewLord creates a new Lord
func NewLord() *Lord {
	return &Lord{
		wg: &sync.WaitGroup{},
	}
}

// StartQuest send minions in a quest. the context will be forwarded to each minion.
// The quest is considered finished when the quest chan is closed
func (master *Lord) StartQuest(ctx context.Context, minions int, quest <-chan Worker) {
	master.wg.Add(minions)
	for i := 0; i < minions; i++ {
		go master.worker(ctx, quest)
	}
}

func (master *Lord) worker(ctx context.Context, input <-chan Worker) {
	defer master.wg.Done()
	done := ctx.Done()
	for {
		// Check first if context is cancelled
		select {
		case <-done:
			return
		default:
		}

		select {
		case worker, open := <-input:
			if !open {
				return
			}
			worker.Work(ctx)
		case <-done:
			return
		}
	}
}

// Wait blocks until all minions return from their quests
func (master Lord) Wait() {
	master.wg.Wait()
}

// NewQuest prepares a slice of works as a buffered channel for safe parallel consumption
func NewQuest(work []Worker) <-chan Worker {
	buf := make(chan Worker, len(work))
	for i := range work {
		buf <- work[i]
	}
	close(buf)
	return buf
}
