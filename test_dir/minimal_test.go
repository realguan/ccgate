package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("Starting test...")
	
	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	
	// Check if we have arguments
	if len(os.Args) > 1 {
		fmt.Printf("Arguments: %v\n", os.Args[1:])
	}
	
	fmt.Println("Test completed.")
}