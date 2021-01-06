/*
func WithValue(parent Context, key, val interface{}) Context
*/

package main

import (
	"fmt"
	"time"

	"context"
)

var key = 1

func work(ctx context.Context) error {

	fmt.Println("value = ", ctx.Value(key))

	return nil
}

func main() {
	ctx := context.WithValue(context.Background(), key, "v1")
	// defer cancel()

	fmt.Println("main begin")

	go work(ctx)

	time.Sleep(time.Second)

	fmt.Println("main end")
}
