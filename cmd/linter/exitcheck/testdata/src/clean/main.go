package main

import "os"

// shutdown calls os.Exit but is not the main function, so it must be ignored.
func shutdown() {
	os.Exit(1)
}

func main() {
	// os.Exit inside a nested closure is out of scope and must be ignored.
	f := func() {
		os.Exit(1)
	}
	_ = f
	_ = shutdown
}
