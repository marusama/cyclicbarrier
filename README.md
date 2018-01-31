cyclicbarrier
=============
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go#goroutines)
[![Build Status](https://travis-ci.org/marusama/cyclicbarrier.svg?branch=master)](https://travis-ci.org/marusama/cyclicbarrier)
[![Go Report Card](https://goreportcard.com/badge/github.com/marusama/cyclicbarrier)](https://goreportcard.com/report/github.com/marusama/cyclicbarrier)
[![Coverage Status](https://coveralls.io/repos/github/marusama/cyclicbarrier/badge.svg?branch=master)](https://coveralls.io/github/marusama/cyclicbarrier?branch=master)
[![GoDoc](https://godoc.org/github.com/marusama/cyclicbarrier?status.svg)](https://godoc.org/github.com/marusama/cyclicbarrier)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)](LICENSE)

CyclicBarrier is a synchronizer that allows a set of goroutines to wait for each other to reach a common execution point, also called a barrier.

Inspired by Java CyclicBarrier https://docs.oracle.com/javase/9/docs/api/java/util/concurrent/CyclicBarrier.html and C# Barrier https://msdn.microsoft.com/en-us/library/system.threading.barrier(v=vs.110).aspx

### Usage
Initiate
```go
import "github.com/marusama/cyclicbarrier"
...
b1 := cyclicbarrier.New(10) // new cyclic barrier with parties = 10
...
b2 := cyclicbarrier.NewWithAction(10, func() error { return nil }) // new cyclic barrier with parties = 10 and with defined barrier action
```
Await
```go
b.Await(ctx)    // await other parties
```
Reset
```go
b.Reset()       // reset the barrier
```

### Simple example
```go
// create a barrier for 10 parties with an action that increments counter
// this action will be called each time when all goroutines reach the barrier
cnt := 0
b := cyclicbarrier.NewWithAction(10, func() error {
    cnt++
    return nil
})

wg := sync.WaitGroup{}
for i := 0; i < 10; i++ {           // create 10 goroutines (the same count as barrier parties)
    wg.Add(1)
    go func() {
        for j := 0; j < 5; j++ {
            
            // do some hard work 5 times
            time.Sleep(100 * time.Millisecond)                     
            
            err := b.Await(context.TODO()) // ..and wait for other parties on the barrier.
                                           // Last arrived goroutine will do the barrier action
                                           // and then pass all other goroutines to the next round
            if err != nil {
                panic(err)
            }
        }
        wg.Done()
    }()
}

wg.Wait()
fmt.Println(cnt)                    // cnt = 5, it means that the barrier was passed 5 times
```

For more documentation see https://godoc.org/github.com/marusama/cyclicbarrier
