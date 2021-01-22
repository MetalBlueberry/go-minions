package minions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type SleepyWorker struct {
	NapTime time.Duration
	Rested  bool
}

func (worker *SleepyWorker) Work(ctx context.Context) {
	time.Sleep(worker.NapTime)
	worker.Rested = true
}

func TestMaster_Start(t *testing.T) {
	type args struct {
		ctx     context.Context
		minions int
		input   []Worker
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1 work",
			args: args{
				ctx:     context.Background(),
				minions: 1,
				input: []Worker{
					&SleepyWorker{
						NapTime: time.Millisecond,
					},
				},
			},
		},
		{
			name: "5 works 2 minions",
			args: args{
				ctx:     context.Background(),
				minions: 2,
				input: []Worker{
					&SleepyWorker{
						NapTime: time.Millisecond,
					},
					&SleepyWorker{
						NapTime: time.Millisecond,
					},
					&SleepyWorker{
						NapTime: time.Millisecond,
					},
					&SleepyWorker{
						NapTime: time.Millisecond,
					},
					&SleepyWorker{
						NapTime: time.Millisecond,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lord := NewLord()
			lord.StartQuest(tt.args.ctx, tt.args.minions, NewQuest(tt.args.input))
			lord.Wait()
			for i := range tt.args.input {
				assert.True(t, tt.args.input[i].(*SleepyWorker).Rested)
			}
		})
	}
}
