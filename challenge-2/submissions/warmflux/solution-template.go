package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Read input from standard input
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := scanner.Text()

		// Call the ReverseString function
		output := ReverseString(input)

		// Print the result
		fmt.Println(output)
	}
}

// ReverseString returns the reversed string of s.
func ReverseString(s string) string {
	runeSlce := []rune(s)
	var newRune []rune
	for i := len(runeSlce) - 1; i >= 0; i-- {
		newRune = append(newRune, runeSlce[i])
	}
	return string(newRune)
}
