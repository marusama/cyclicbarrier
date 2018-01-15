package cyclicbarrier

import (
	"context"
	"sync"
	"testing"
	"time"
)

func checkBarrier(t *testing.T, b CyclicBarrier, parties, count int) {
	// cb := b.(*cyclicBarrier)
	// if cb.parties != parties {
	// 	t.Error("barrier must have parties = ", parties, ", but has ", cb.parties)
	// }
	// if cb.count != count {
	// 	t.Error("barrier must have count = ", count, ", but has ", cb.count)
	// }
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

func TestAwaitOnceCtxDone(t *testing.T) {
	n := 100        // goroutines count
	b := New(n + 1) // more than goroutines count so all goroutines will wait
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			err := b.Await(ctx)
			if err != context.DeadlineExceeded {
				panic("must be context.DeadlineExceeded error")
			}
			wg.Done()
		}()
	}

	wg.Wait()
	checkBarrier(t, b, n+1, 0)
}

//go:norace
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
		go func(num int) {
			for j := 0; j < m; j++ {
				err := b.Await(nil) //, num)
				if err != nil {
					panic(err)
				}
				//println("pass", num)
			}
			wg.Done()
			//println("exit", num)
		}(i)
	}

	wg.Wait()
	checkBarrier(t, b, n, 0)
}

func TestAwaitManyCtxDone(t *testing.T) {
	n := 100 // goroutines count
	b := New(n)
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			for {
				err := b.Await(ctx)
				if err != nil {
					if err != context.DeadlineExceeded {
						panic("must be context.DeadlineExceeded error")
					}
					break
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
