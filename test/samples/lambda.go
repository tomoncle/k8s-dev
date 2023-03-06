package main

import "fmt"

func LambdaCompare(f func() bool, a, b interface{}) interface{} {
	if f() {
		return a
	} else {
		return b
	}
}

var opt = func(b bool) *bool { return &b }(false)
var ops = func(i int32) *int32 { return &i }(1)

func main() {
	x, y := 0, 30

	fmt.Println(LambdaCompare(func() bool {
		return x > 0
	}, x, y))

	fmt.Println(*opt, *ops)
}
