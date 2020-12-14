package pool

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestPool_NewPool(t *testing.T) {

	t.Run("Initializing a pool of size 1 with New Command", func(t *testing.T) {
		p := New(1)

		ch := make(chan struct{})

		xid, err := p.Add(func(taskID string) error {
			ch <- struct{}{}
			return nil
		})

		if xid == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err)
		}

		go p.Server()

		<-ch

	})

}

func TestPool_GettingDoneInfo(t *testing.T) {

	t.Run("A simple information", func(t *testing.T) {
		p := New(1)

		xid, err := p.Add(func(taskID string) error {
			return nil
		})

		if xid == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err)
		}

		ch := p.GetInfoChannel()

		go p.Server()

		wf := <-ch

		if wf.ID != xid {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

}

func TestPool_AddingOnBeforeRunningServer(t *testing.T) {

	t.Run("A simple information", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(1)

		xid, err1 := p.Add(func(taskID string) error {
			ch <- struct{ ID string }{ID: taskID}

			return nil
		})

		if xid == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err1 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		go p.Server()

		wf := <-ch

		if wf.ID != xid {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

	t.Run("multipe adds on same rotine before server", func(t *testing.T) {
		chQuit := make(chan struct{})
		p := New(100)

		checkList := make(map[string]bool)

		nIter := 1000

		for i := 0; i < nIter; i++ {
			xid, err := p.Add(func(taskID string) error {
				return nil
			})

			if xid == "" {
				t.Error("p.Add() returned an empty ID")
			}

			if err != nil {
				t.Errorf("p.Add() returned an inexpected error: %s", err)
			}

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

		/*
			if err2 != nil {
				t.Errorf("p.Add() returned an inexpected error: %s", err1)
			}

			if wf.ID != xid2 {
				t.Errorf("unexpected ID on wf, got: %s", wf.ID)
			}
		*/
	})

	t.Run("multipe adds with multiples rotine before server", func(t *testing.T) {
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
					xid, err := p.Add(func(taskID string) error {
						return nil
					})

					if xid == "" {
						t.Error("p.Add() returned an empty ID")
					}

					if err != nil {
						t.Errorf("p.Add() returned an inexpected error: %s", err)
					}

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

func TestPool_AddingOnRunningServer(t *testing.T) {

	t.Run("A simple information", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(1)

		xid1, err1 := p.Add(func(taskID string) error {
			time.Sleep(5 * time.Millisecond)

			return nil
		})

		if xid1 == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err1 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		go p.Server()

		xid2, err2 := p.Add(func(taskID string) error {
			time.Sleep(10 * time.Millisecond)
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		if err2 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		wf := <-ch

		if wf.ID != xid2 {
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
			xid, err := p.Add(func(taskID string) error {
				return nil
			})

			if xid == "" {
				t.Error("p.Add() returned an empty ID")
			}

			if err != nil {
				t.Errorf("p.Add() returned an inexpected error: %s", err)
			}

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

		/*
			if err2 != nil {
				t.Errorf("p.Add() returned an inexpected error: %s", err1)
			}

			if wf.ID != xid2 {
				t.Errorf("unexpected ID on wf, got: %s", wf.ID)
			}
		*/
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
					xid, err := p.Add(func(taskID string) error {
						return nil
					})

					if xid == "" {
						t.Error("p.Add() returned an empty ID")
					}

					if err != nil {
						t.Errorf("p.Add() returned an inexpected error: %s", err)
					}

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
					xid, err := p.Add(func(taskID string) error {
						return nil
					})

					if xid == "" {
						t.Error("p.Add() returned an empty ID")
					}

					if err != nil {
						t.Errorf("p.Add() returned an inexpected error: %s", err)
					}

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

func TestPool_ReSettingPoolSize(t *testing.T) {

	t.Run("Resize to a lesser pool size", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(100)

		xid1, err1 := p.Add(func(taskID string) error {
			time.Sleep(5 * time.Millisecond)

			return nil
		})

		if xid1 == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err1 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		go p.Server()

		p.SetMaxPoolSize(50)

		xid2, err2 := p.Add(func(taskID string) error {
			time.Sleep(10 * time.Millisecond)
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		if err2 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		wf := <-ch

		if wf.ID != xid2 {
			t.Errorf("unexpected ID on wf, got: %s", wf.ID)
		}

	})

	t.Run("Resize to a bigger pool size", func(t *testing.T) {

		ch := make(chan struct{ ID string })

		p := New(10)

		xid1, err1 := p.Add(func(taskID string) error {
			time.Sleep(5 * time.Millisecond)

			return nil
		})

		if xid1 == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err1 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		go p.Server()

		p.SetMaxPoolSize(50)

		xid2, err2 := p.Add(func(taskID string) error {
			time.Sleep(10 * time.Millisecond)
			ch <- struct{ ID string }{ID: taskID}
			return nil
		})

		if err2 != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err1)
		}

		wf := <-ch

		if wf.ID != xid2 {
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
		nPerRoutines := 100
		nIter := nRoutines * nPerRoutines
		nMaxIter := nIter * 2

		go func() {
			//count start from 1 becase the first one will be waiting
			count := 1
			for xid := range chMap {
				count++

				checkListGenerated = append(checkListGenerated, xid)
				if count >= nMaxIter {
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

				if count >= nMaxIter {
					break
				}

			}

			chQuit <- struct{}{}

		}()

		time.Sleep(10 * time.Microsecond)

		xid, err := p.Add(func(taskID string) error {
			<-chWait
			return nil
		})

		if xid == "" {
			t.Error("p.Add() returned an empty ID")
		}

		if err != nil {
			t.Errorf("p.Add() returned an inexpected error: %s", err)
		}

		chMap <- xid

		for i := 0; i < nRoutines; i++ {

			go func() {

				for j := 0; j < nPerRoutines; j++ {
					xid, err := p.Add(func(taskID string) error {
						time.Sleep(100 * time.Millisecond)
						return nil
					})

					if xid == "" {
						t.Error("p.Add() returned an empty ID")
					}

					if err != nil {
						t.Errorf("p.Add() returned an inexpected error: %s", err)
					}

					chMap <- xid

				}

			}()

		}

		go p.Server()

		p.SetMaxPoolSize(10)

		for i := 0; i < nRoutines; i++ {

			go func() {

				for j := 0; j < nPerRoutines; j++ {
					xid, err := p.Add(func(taskID string) error {
						time.Sleep(10 * time.Millisecond)
						return nil
					})

					if xid == "" {
						t.Error("p.Add() returned an empty ID")
					}

					if err != nil {
						t.Errorf("p.Add() returned an inexpected error: %s", err)
					}

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
