package main

import (
	"fmt"
	"strconv"
)

func main() {
	composeFiles := getWatchedComposeFiles()
	fmt.Println("Found " + strconv.Itoa(len(composeFiles)) + " compose files with watched containers")
	for composeFile, containers := range composeFiles {
		fmt.Println("Compose File: " + composeFile)
		for _, container := range containers {
			fmt.Println("    Container: " + container.id + " (" + container.name + ")")
			fmt.Println("        Running Image: " + container.image.id + " (" + container.image.hash + ")")
			fmt.Println("        Latest Image:  " + container.image.id + " (" + getImageHash(container.image.id) + ")")
		}
	}
}
