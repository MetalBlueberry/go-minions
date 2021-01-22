package quests

import (
	"context"
	"testing"

	"github.com/MetalBlueberry/go-minions/pkg/minions"
	"github.com/stretchr/testify/assert"
)

func TestFunction_Work(t *testing.T) {
	done := false
	work := Function(func(c context.Context) {
		done = true
	})

	works := []minions.Worker{
		work,
	}

	lord := minions.NewLord()
	lord.StartQuest(context.Background(), 1, minions.NewQuest(works))

	lord.Wait()

	assert.True(t, done)
}
