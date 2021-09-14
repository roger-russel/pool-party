package pool

import (
	"testing"
)

func TestPool_AddingOnBeforeRunningServer(t *testing.T) {
	t.Run("A simple information", func(t *testing.T) {
		ch := make(chan struct{ ID string })

		p := New(1)

		xid := add(t, p, func(taskID string) error {
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		go p.Server()

		wf := <-ch

		if wf.ID != xid {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}
	})

	t.Run("multiple adds on same rotine before server", func(t *testing.T) {
		chQuit := make(chan struct{})
		p := New(100)

		checkList := make(map[string]bool)

		nIter := 1000

		for i := 0; i < nIter; i++ {
			xid := add(t, p, func(taskID string) error {
				return nil
			})

			checkList[xid] = false
		}

		ch := p.GetInfoChannel()

		go func() {
			count := 0
			for w := range ch {
				count++

				used, ok := checkList[w.ID]
				checkList[w.ID] = true

				if !ok {
					t.Errorf("received unexpected id: %s", w.ID)
				}

				if used {
					t.Errorf("received already used id: %s, it looks like the task run twice", w.ID)
				}

				if count >= nIter {
					break
				}
			}

			chQuit <- struct{}{}
		}()

		go p.Server()

		<-chQuit
	})

	t.Run("multiple adds with multiples rotine before server", func(t *testing.T) {
		chQuit := make(chan struct{})
		chMap := make(chan string)
		chNext := make(chan struct{})

		p := New(10)

		checkList := make(map[string]bool)

		nRoutines := 10
		nPerRoutines := 100
		nMaxIter := nRoutines * nPerRoutines

		go func() {
			count := 0
			for xid := range chMap {
				count++
				checkList[xid] = false
				if count >= nMaxIter {
					chNext <- struct{}{}
				}
			}
		}()

		for i := 0; i < nRoutines; i++ {
			go func() {
				for j := 0; j < nPerRoutines; j++ {
					xid := add(t, p, func(taskID string) error {
						return nil
					})

					chMap <- xid
				}
			}()
		}

		<-chNext

		ch := p.GetInfoChannel()

		go func() {
			count := 0
			for w := range ch {
				count++

				used, ok := checkList[w.ID]
				checkList[w.ID] = true

				if !ok {
					t.Errorf("received unexpected id: %s", w.ID)
				}

				if used {
					t.Errorf("received already used id: %s, it looks like the task run twice", w.ID)
				}

				if count >= nMaxIter {
					break
				}
			}

			chQuit <- struct{}{}
		}()

		go p.Server()

		<-chQuit
	})
}
