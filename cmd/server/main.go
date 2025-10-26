package main

import (
	"Geomyidae/cmd/server/sock_server"
	"fmt"
	"github.com/doby162/go-higher-order"
)

func main() {

	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	b := higher_order.ReduceSlice(a, func(a, b int) int { return a + b })
	fmt.Println(b)

	err := sock_server.Api()
	if err != nil {
		return
	}
}
