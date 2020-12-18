package pool

import (
	"testing"
	"time"
)

func TestPool_ReSettingPoolSize(t *testing.T) {

	t.Run("Resize to a lesser pool size", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(10)

		add(t, p, func(taskID string) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		})

		go p.Server()

		p.SetMaxPoolSize(5)

		xidAfterServerStarted := add(t, p, func(taskID string) error {
			time.Sleep(10 * time.Millisecond)
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		wf := <-ch

		if wf.ID != xidAfterServerStarted {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

	t.Run("Resize to a bigger pool size", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(10)

		add(t, p, func(taskID string) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		})

		go p.Server()

		p.SetMaxPoolSize(50)

		xidAfterServerStarted := add(t, p, func(taskID string) error {
			time.Sleep(10 * time.Millisecond)
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		wf := <-ch

		if wf.ID != xidAfterServerStarted {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

	t.Run("Resize with Qeued tasks to a bigger pool", func(t *testing.T) {
		chQuit := make(chan struct{})
		chWait := make(chan struct{})

		chMap := make(chan string)

		p := New(1)
		ch := p.GetInfoChannel()

		checkListGenerated := []string{}
		checkListReceived := []string{}

		nRoutines := 10
		nPerRoutines := 20
		nIter := nRoutines * nPerRoutines
		nMaxIter := nIter * 2

		go func() {
			//count start from 1 becase the first one will be waiting
			count := 0
			for xid := range chMap {
				count++

				checkListGenerated = append(checkListGenerated, xid)

				if count == nMaxIter {
					chWait <- struct{}{}
				}

				if count > nMaxIter {
					break
				}
			}

		}()

		go func() {

			count := 0

			for w := range ch {
				count++

				checkListReceived = append(checkListReceived, w.ID)

				if count > nMaxIter {
					break
				}

			}

			chQuit <- struct{}{}

		}()

		time.Sleep(10 * time.Microsecond)

		xid := add(t, p, func(taskID string) error {
			<-chWait
			return nil
		})

		chMap <- xid

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

		go p.Server()

		p.SetMaxPoolSize(10)

		for i := 0; i < nRoutines; i++ {

			go func() {

				for j := 0; j < nPerRoutines; j++ {

					xid := add(t, p, func(taskID string) error {
						time.Sleep(10 * time.Millisecond)
						return nil
					})

					chMap <- xid

				}

			}()

		}

		<-chQuit

		if len(checkListGenerated) != len(checkListReceived) {
			t.Errorf("Generate list and Receive list have differente sizes. Generate: %d, Received: %d", len(checkListGenerated), len(checkListReceived))
		}

		mGenerated := toMap(t, checkListGenerated)
		mReceived := toMap(t, checkListReceived)

		for v := range mGenerated {
			if _, ok := mReceived[v]; !ok {
				t.Errorf("id not found: %s", v)
			}
		}

	})

}
