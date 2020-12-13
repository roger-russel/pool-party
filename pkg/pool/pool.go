package pool

import (
	"time"

	"github.com/rs/xid"
)

//New Pool Server
func New(size int) Pool {

	return &Impl{
		size:       size,
		chDone:     make(chan *WorkerInfo),
		chQuit:     make(chan struct{}),
		chShutdown: make(chan struct{}),
		chInfo:     make(chan *WorkerInfo),
		queue:      &queueImpl{},
	}

}

//Pool to create a pool
type Pool interface {

	// Server will be a pool server wich will wait until a shootdown as send
	Server()

	// Add worker to pool
	Add(f func(taskID string) error) (id string, err error)

	// SetMaxPoolSize pool max size to a different number
	SetMaxPoolSize(s int)

	// Shutdown start shutdown pool server
	Shutdown()

	GetInfoChannel() chan *WorkerInfo
}

//Impl is the Pool interface implementation
type Impl struct {

	//Pool size
	size int

	// started is true when the pool is already calling the workers ( Run() or Server() )
	started bool

	// Time of the run was called
	startedAt time.Time

	//shuttingDown is true when this pool is asked to shutdown
	shuttingDown bool

	//Quit is the last channel to be called, it will lead to end of server.
	chQuit chan struct{}

	//Done channel called when an workerd finish
	chDone chan *WorkerInfo

	//Shutdown channel called when want to shutdown this pool it will send a shutdown fignal for ,
	chShutdown chan struct{}

	//This channel receive info when the worker is finished
	chInfo chan *WorkerInfo

	//queue
	queue queue

	//number of worker running
	runningWorkers int
}

//Server start pool as service
func (p *Impl) Server() {

	p.started = true
	p.startedAt = time.Now()

	go p.done()

	//wait for done channel be on
	time.Sleep(10 * time.Millisecond)

	for p.callNextWorker() {
		// call until false
	}

	<-p.chQuit

	close(p.chQuit)
	close(p.chDone)
	close(p.chShutdown)
	close(p.chInfo)

}

/*Add new workers to pool
 *The given function will be queued and called further
 */
func (p *Impl) Add(f func(taskID string) error) (taskID string, err error) {

	if p.shuttingDown {
		return "", errAddONShuttingDown
	}

	taskID = xid.New().String()

	p.queue.put(&workerImpl{
		id:       taskID,
		queuedAt: time.Now(),
		run: func() error {
			return f(taskID)
		},
		chDone: p.chDone,
	})

	if p.started {
		p.callNextWorker()
	}

	return taskID, nil
}

/*SetMaxPoolSize pool max size to a different number
 * It should change the pull size and start more workers
 * if it the new pool size is greater then the previous one
 * and if there is more workers to start.
 * if the new pull size is lower, it should do nothing,
 * because when a current task finish it will check if
 * it should call a new worker or not.
 */
func (p *Impl) SetMaxPoolSize(size int) {

	current := p.size
	p.size = size

	if current >= size {
		return
	}

	for p.callNextWorker() {
		// call until false
	}

}

//Call Next Worker to be executed
func (p *Impl) callNextWorker() bool {

	if p.shuttingDown || p.size <= p.runningWorkers || p.queue.len() <= 0 {
		return false
	}

	w := p.queue.get(&p.runningWorkers)

	if w != nil {
		go w.Start()
	}

	return p.queue.len() > 0

}

/*Shutdown force a shutdown on application, it will wait only for
 *the running tasks finish, for a Graceful shutdown take a look into
 *which will try to run all tasks on pool which are been queued ShutdownGraceful
 */
func (p *Impl) Shutdown() {

	p.shuttingDown = true
	p.chShutdown <- struct{}{}

	/*
		for _, w := range p.queue.len() {
			p.chInfo <- w.Info(false, errQueuedTasksWillBeExecutedNoMore)
		}
	*/
}

/*GetInfoChannel returns the info channel used on pool
 *The info channel will receive information about finished
 *tasks.
 *on shutdown it will receive all tasks which are not
 *runned yet
 */
func (p *Impl) GetInfoChannel() chan *WorkerInfo {
	return p.chInfo
}

func (p *Impl) done() {

	for range p.chDone {

		p.runningWorkers--

		if p.shuttingDown && p.runningWorkers == 0 {
			p.chQuit <- struct{}{}
			break
		}

		p.callNextWorker()

	}
}
