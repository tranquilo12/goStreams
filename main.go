package main

import (
	"lightning/cmd"
)

func main() {
	// CPU profiling
	//defer profile.Start(profile.CPUProfile).Stop()

	// Memory profiling
	//defer profile.Start(profile.MemProfile).Stop()

	cmd.Execute()
}
