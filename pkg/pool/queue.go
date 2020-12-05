package pool

import (
	"sync"
)

//queue interface
type queue interface {
	put(w worker)
	get(inc *int) worker
	len() int
}

//queue Struct
type queueImpl struct {
	mu     sync.Mutex
	queued []worker
	length int
}

//put worker on queue
func (q *queueImpl) put(w worker) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queued = append(q.queued, w)
	q.length = len(q.queued)
}

//get worker from pool
func (q *queueImpl) get(inc *int) (w worker) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queued) > 0 {
		w = q.queued[0]
		q.queued = q.queued[1:]
		*inc++
		q.length = len(q.queued)
		return w
	}

	return nil
}

func (q *queueImpl) len() int {
	return q.length
}
