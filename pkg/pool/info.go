package pool

import (
	"time"
)

//WorkerInfo interface are some info about the running workers
type WorkerInfo interface {
	ID() string
	QueuedAt() time.Time
	WaitingTime() time.Duration
	ExecutionTime() time.Duration
	Error() error
}

//WorkerInfoImpl is the implementation about the worker
type WorkerInfoImpl struct {
	id            string
	err           error
	queuedAt      time.Time
	waitingTime   time.Duration
	executionTime time.Duration
}

//ID returns the worker ID
func (wi *WorkerInfoImpl) ID() string {
	return wi.id
}

//QueuedAt return the queued wait time
func (wi *WorkerInfoImpl) QueuedAt() time.Time {
	return wi.queuedAt
}

//WaitingTime return the waiting on queue time
func (wi *WorkerInfoImpl) WaitingTime() time.Duration {
	return wi.waitingTime
}

//ExecutionTime is the time that the process took to run
func (wi *WorkerInfoImpl) ExecutionTime() time.Duration {
	return wi.executionTime
}

//Error if some error happened on worker execution
func (wi *WorkerInfoImpl) Error() error {
	return wi.err
}
