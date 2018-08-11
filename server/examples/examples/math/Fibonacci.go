package main

import "fmt"

func main() {
	a, b := 0, 1
	for b < 1000 {
		fmt.Println(b)
		a, b = b, a+b
	}
}
