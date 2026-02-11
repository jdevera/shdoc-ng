package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	parser := NewParser()
	for scanner.Scan() {
		parser.ProcessLine(scanner.Text())
	}
	fmt.Print(parser.Render())
}
