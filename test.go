package main

import "fmt"

func main() {
	x := []int{1, 2, 3, 4}
	y := []int{5, 6, 7, 8}

	x = append(x, y...)

	fmt.Println(x)
}
