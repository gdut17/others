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

	for {
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("sleep 1s")

		// we received the signal of cancelation in this channel
		case <-ctx.Done():
			fmt.Println("Cancel by main")
			return ctx.Err()
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	fmt.Println("main begin")

	wg.Add(1)
	go work(ctx)
	wg.Wait()

	fmt.Println("main end")
}
