/*
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)

WithCancel返回值第一个为子节点，第二个为cancel函数
外部函数调用cancal(), 关闭这个chan, 子协程的只读chan就会退出。

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
		//case <-time.After(1 * time.Second):
		//	fmt.Println("sleep 1s")

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

	//var arg string
	//fmt.Scanf("%s", &arg)
	cancel()

	fmt.Println("main end")

	time.Sleep(time.Second * 3)
}
