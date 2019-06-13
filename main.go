package main

import "fmt"

func main() {
	fmt.Println("Hello, world.")
	containerIDs := getWatchedRunningContainers()
	for _, container := range containerIDs {
		fmt.Println("Container: " + container.id + " (" + container.name + ")")
		fmt.Println("    ComposeFile: " + container.composeFile)
		fmt.Println("    Image: " + container.image.id + " (" + container.image.hash + ")")
	}
}
