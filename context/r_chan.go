package main

import (
	"fmt"
	"time"
)

// 参数为一个只读chan
func work(r <-chan struct{}) {
	fmt.Println("child begin")
	<-r
	// 后面的不会执行
	fmt.Println("child end")
}
func main() {
	var r chan struct{}
	r = make(chan struct{})

	fmt.Println("main begin")

	go work(r)

	time.Sleep(time.Second * 2)
	close(r)

	fmt.Println("main end")

}
