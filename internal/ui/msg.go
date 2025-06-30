package ui

import "fmt"

func Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}

func Println(a ...any) {
	fmt.Println(a...)
}
