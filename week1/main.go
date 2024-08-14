package main

import "fmt"

func DelSliceByIndex[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}

	copy(slice[index:], slice[index+1:])
	newSlice := slice[:len(slice)-1]

	if cap(newSlice) > len(newSlice)*5/4 {
		newSlice = append([]T(nil), newSlice...)
	}

	return newSlice
}

func main() {
	testSlice := []int{1, 2, 3, 4, 5}
	retSlice := DelSliceByIndex(testSlice, 2)
	fmt.Println(retSlice)
}
