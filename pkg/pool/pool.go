package pool

import (
	"sync"
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

	mu sync.Mutex
}

//Server start pool as service
func (p *Impl) Server() {

	p.startedAt = time.Now()

	p.mu.Lock()
	p.started = true
	p.mu.Unlock()

	go p.done()

	//wait for done channel be on
	time.Sleep(10 * time.Millisecond)

	until := p.callNextWorker()
	for until {
		until = p.callNextWorker()
	}

	<-p.chQuit

	time.Sleep(10 * time.Microsecond)

	close(p.chQuit)
	close(p.chDone)
	close(p.chShutdown)
	//close(p.chInfo)

}

/*Add new workers to pool
 *The given function will be queued and called further
 */
func (p *Impl) Add(f func(taskID string) error) (taskID string, err error) {

	p.mu.Lock()

	if p.shuttingDown {
		p.mu.Unlock()
		return "", errAddONShuttingDown
	}

	p.mu.Unlock()

	taskID = xid.New().String()

	p.queue.put(&workerImpl{
		id:       taskID,
		queuedAt: time.Now(),
		run: func() error {
			return f(taskID)
		},
		chDone: p.chDone,
	})

	p.callNextWorker()

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

	p.mu.Lock()

	current := p.size
	p.size = size

	p.mu.Unlock()

	if current >= size {
		return
	}

	until := p.callNextWorker()
	for until {
		until = p.callNextWorker()
	}

}

//Call Next Worker to be executed
func (p *Impl) callNextWorker() bool {

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started || p.shuttingDown || p.size <= p.runningWorkers || p.queue.len() <= 0 {
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

	p.mu.Lock()
	p.shuttingDown = true
	p.mu.Unlock()
	//	p.chShutdown <- struct{}{}

	w := p.queue.get(nil)

	for w != nil {
		p.chInfo <- w.Info(false, errQueuedTasksWillBeExecutedNoMore)
		w = p.queue.get(nil)
	}

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

	for w := range p.chDone {

		p.mu.Lock()
		p.runningWorkers--
		p.mu.Unlock()

		if p.shuttingDown && p.runningWorkers == 0 {
			p.chQuit <- struct{}{}
			break
		}

		p.callNextWorker()

		p.chInfo <- w

	}
}
