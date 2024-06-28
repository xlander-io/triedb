package main

import "fmt"

func main() {
	x := []byte{111, 2, 3}
	z := (x[0] & 0xF0) >> 4
	fmt.Println(fmt.Sprintf("%x", z))
}
