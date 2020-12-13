package pool

import (
	"testing"
	"time"
)

func Test_queueImpl_put(t *testing.T) {
	t.Run("Put on Queue", func(t *testing.T) {
		q := queueImpl{}

		q.put(&workerImpl{
			id:       "1",
			queuedAt: time.Now(),
			run: func() error {
				return func(id string) error {
					return nil
				}("1")
			},
		})

		if len(q.queued) != 1 {
			t.Errorf("unexpeced size of queue: %d", len(q.queued))
		}

		if q.len() != 1 {
			t.Errorf("queue dont looks like it increased lenth of queued workers, got: %d", q.len())
		}

	})
}
