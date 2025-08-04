package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse" //nolint:depguard
)

func main() {
	inputText := "Hello, OTUS!"
	outputText := reverse.String(inputText)
	fmt.Println(outputText)
}
