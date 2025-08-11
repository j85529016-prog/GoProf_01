package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	inputText := "Hello, OTUS!"
	outputText := reverse.String(inputText)
	fmt.Println(outputText)
}
