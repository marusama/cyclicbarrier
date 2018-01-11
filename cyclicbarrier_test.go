package cyclicbarrier

import (
	"sync"
	"testing"
	"time"
)

func checkBarrier(t *testing.T, b CyclicBarrier, parties, count int) {
	cb := b.(*cyclicBarrier)
	if cb.parties != parties {
		t.Error("barrier must have parties = ", parties, ", but has ", cb.parties)
	}
	if cb.count != count {
		t.Error("barrier must have count = ", count, ", but has ", cb.count)
	}
}

func TestAwaitOnce(t *testing.T) {
	n := 100 // goroutines count
	b := New(n)

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			err := b.Await(nil)
			if err != nil {
				panic(err)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	checkBarrier(t, b, n, 0)
}

func TestAwaitStepByStep(t *testing.T) {
	b := New(3)

	ch := [3]chan struct{}{}
	ch[0] = make(chan struct{})
	ch[1] = make(chan struct{})
	ch[2] = make(chan struct{})

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			time.Sleep(time.Duration(100*(n+1)) * time.Millisecond)
			ch[n] <- struct{}{}
			err := b.Await(nil)
			if err != nil {
				panic(err)
			}
			wg.Done()
		}(i)
	}

	checkBarrier(t, b, 3, 0)

	<-ch[0]
	checkBarrier(t, b, 3, 1)

	<-ch[1]
	checkBarrier(t, b, 3, 2)

	<-ch[2]
	wg.Wait()
	checkBarrier(t, b, 3, 0)
}

func TestAwaitMany(t *testing.T) {
	n := 100  // goroutines count
	m := 1000 // inner cycle count
	b := New(n)

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < m; j++ {
				err := b.Await(nil)
				if err != nil {
					panic(err)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	checkBarrier(t, b, n, 0)
}

func TestAwaitAction(t *testing.T) {
	n := 100  // goroutines count
	m := 1000 // inner cycle count

	cnt := 0
	b := NewWithAction(n, func() {
		cnt++
	})

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < m; j++ {
				err := b.Await(nil)
				if err != nil {
					panic(err)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	checkBarrier(t, b, n, 0)
	if cnt != m {
		t.Error("cnt must be equal to = ", m, ", but it's ", cnt)
	}
}
