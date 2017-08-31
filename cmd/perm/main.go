package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("bosh job")
		time.Sleep(time.Second * 10)
	}
}
