/*
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)


*/

package main

import (
	"fmt"
	"time"

	"context"
)

func work(ctx context.Context) error {

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
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	fmt.Println("main begin")

	go work(ctx)

	var arg string
	fmt.Scanf("%s", &arg)
	cancel()

	fmt.Println("main end")
}
