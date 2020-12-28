package pool

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestPool_AddingOnRunningServer(t *testing.T) {

	t.Run("A simple information", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(1)

		add(t, p, func(taskID string) error {
			time.Sleep(5 * time.Millisecond)

			return nil
		})

		go p.Server()

		xid := add(t, p, func(taskID string) error {
			time.Sleep(10 * time.Millisecond)
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		wf := <-ch

		if wf.ID != xid {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

	t.Run("multipe adds on same rotine after server", func(t *testing.T) {
		chQuit := make(chan struct{})

		p := New(100)
		go p.Server()

		time.Sleep(10 * time.Microsecond)

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

		<-chQuit

		for i, v := range checkList {
			if v != true {
				t.Errorf("id was not received on info channel: %s", i)
			}
		}

	})

	t.Run("multipe adds with multiples rotine after server", func(t *testing.T) {
		chQuit := make(chan struct{})
		chMap := make(chan string)

		p := New(10)

		checkListGenerated := []string{}
		checkListReceived := []string{}

		nRoutines := 10
		nPerRoutines := 100
		nIter := nRoutines * nPerRoutines
		nMaxIter := nIter * 2

		go func() {
			count := 0
			for xid := range chMap {
				count++
				checkListGenerated = append(checkListGenerated, xid)
				if count >= nMaxIter {
					break
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

		ch := p.GetInfoChannel()

		go func() {
			count := 0

			for w := range ch {
				count++

				checkListReceived = append(checkListReceived, w.ID)

				if count >= nMaxIter {
					break
				}

			}

			chQuit <- struct{}{}

		}()

		go p.Server()

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

		<-chQuit

		if len(checkListGenerated) != len(checkListReceived) {
			t.Errorf("Generate list and Receive list have differente sizes. Generate: %d, Received: %d", len(checkListGenerated), len(checkListReceived))
		}

		sort.Strings(checkListGenerated)
		sort.Strings(checkListReceived)

		if !cmp.Equal(checkListGenerated, checkListReceived) {
			t.Errorf("something went wrong: %s", cmp.Diff(checkListGenerated, checkListReceived))
		}

	})

}
