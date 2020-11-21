package pool

import (
	"sync"
	"testing"
	"time"
)

func TestPool_SetMaxPoolSize(t *testing.T) {
	type fields struct {
		size      int
		workers   map[int]workerInterface
		running   map[int]bool
		wg        sync.WaitGroup
		chDone    chan int
		tasks     int
		startedAt time.Time
	}
	type args struct {
		s int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				size:      tt.fields.size,
				workers:   tt.fields.workers,
				running:   tt.fields.running,
				wg:        tt.fields.wg,
				chDone:    tt.fields.chDone,
				tasks:     tt.fields.tasks,
				startedAt: tt.fields.startedAt,
			}
			p.SetMaxPoolSize(tt.args.s)
		})
	}
}
