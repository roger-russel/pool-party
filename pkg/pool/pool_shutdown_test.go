package pool

import (
	"errors"
	"testing"
)

func TestPool_AddingAfterShutdown(t *testing.T) {

	t.Run("Adding on Shutdown", func(t *testing.T) {
		//chWait := make(chan struct{})

		p := New(1)

		p.Shutdown()

		xid, err := p.Add(func(taskID string) error {
			//	<-chWait
			return nil
		})

		if xid != "" {
			t.Errorf("On shutdown add should not generate any xid but received: %s", xid)
		}

		if !errors.Is(err, errAddONShuttingDown) {
			t.Errorf("Error is different from expected got: %s, expected: %s", err, errAddONShuttingDown)
		}

	})

}

/*
func TestPool_AddingBeforeShutdown(t *testing.T) {

	t.Run("Adding receiving queued on shutdown", func(t *testing.T) {

		chWaitStart := make(chan struct{})
		chLock := make(chan struct{})
		chQuit := make(chan struct{})

		p := New(1)

		pid := add(t, p, func(taskID string) error {
			chWaitStart <- struct{}{}
			return nil
		})

		vid := add(t, p, func(taskID string) error {
			<-chLock
			chQuit <- struct{}{}
			return nil
		})

		xid := add(t, p, func(taskID string) error {
			return nil
		})

		ch := p.GetInfoChannel()

		go func() {

			for w := range ch {

				fmt.Println(w.ID, pid, xid, vid)

				if w.ID == pid || w.ID == vid {
					continue
				}

				if !errors.Is(w.Err, errQueuedTasksWillBeExecutedNoMore) {
					t.Errorf("unexpected error got: %s, expected: %s", w.Err, errQueuedTasksWillBeExecutedNoMore)
				}

				if w.ID != xid {
					t.Errorf("unexpected id\n%s\n%s\n%s", w.ID, xid, pid)
				}

				break

			}

			chLock <- struct{}{}

		}()

		go p.Server()

		<-chWaitStart

		p.Shutdown()

		<-chQuit

	})

}
*/
