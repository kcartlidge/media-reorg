package main

import (
	"fmt"
	"os"
)

// check prints an error message and exits the program if the error is not nil
func check(err error) {
	if err != nil {

		// show the error and stop
		fmt.Println()
		fmt.Println()
		fmt.Println("ERROR:")
		fmt.Fprintln(os.Stderr, err)
		fmt.Println()
		fmt.Println()
		os.Exit(1)
	}
}
