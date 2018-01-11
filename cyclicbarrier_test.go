package cyclicbarrier

import (
	"sync"
	"testing"
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
