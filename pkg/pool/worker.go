package pool

import (
	"time"
)

type worker interface {
	Start(chDone chan WorkerInfo)
	ID() string
}

//workerImpl Struct
type workerImpl struct {
	run       func() error
	id        string
	queuedAt  time.Time
	startedAt time.Time
}

func (w *workerImpl) Start(chDone chan WorkerInfo) {

	w.startedAt = time.Now()
	err := w.run()

	chDone <- &WorkerInfoImpl{
		id:            w.id,
		err:           err,
		queuedAt:      w.queuedAt,
		waitingTime:   w.startedAt.Sub(w.queuedAt),
		executionTime: time.Since(w.startedAt),
	}
}

func (w *workerImpl) ID() string {
	return w.id
}
