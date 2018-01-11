package cyclicbarrier

import (
	"context"
	"fmt"
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

func TestAwaitOnce2(t *testing.T) {
	b := New(3)

	ch := make(chan struct{})
	for i := 0; i < 3; i++ {
		go func(n int) {
			if n == 0 {
				time.Sleep(5 * time.Second)
			}
			err := b.Await(nil)
			if err != nil {
				panic(err)
			}
			ch <- struct{}{}
		}(i)
	}

	<-ch // first gorouitine done
	<-ch // second goroutine done

	ticker := time.NewTimer(100 * time.Millisecond)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	select {
	case <-ch: // third gorouitine done
		t.Error("third goroutine must not arrive earlier")
		return
	case <-time.After(1 * time.Second):
		fmt.Println("HELLO")
		//checkBarrier(t, b, 3, 2)
	case <-ticker.C:
		fmt.Println("HELLO 2")
	case <-ctx.Done():
		fmt.Println("HELLO 3")
	}
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
