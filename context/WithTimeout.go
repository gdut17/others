/*

func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}
*/

package main

import (
	"fmt"
	"sync"
	"time"

	"context"
)

var (
	wg sync.WaitGroup
)

func work(ctx context.Context) error {
	defer wg.Done()

	//for {
	select {
	//case <-time.After(1 * time.Second):
	//	fmt.Println("sleep 1s")

	// we received the signal of cancelation in this channel
	case <-ctx.Done():
		fmt.Println("read <-chan, Cancel by main")
		return ctx.Err()
	}
	//}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	//defer cancel()

	ctx2, cancel2 := context.WithTimeout(ctx, 4*time.Second)
	//defer cancel2()

	fmt.Println("main begin")

	wg.Add(2)
	go work(ctx)
	go work(ctx2)

	time.Sleep(time.Second)
	cancel()
	cancel2()
	wg.Wait()

	fmt.Println("main end")
}
