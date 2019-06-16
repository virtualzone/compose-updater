package main

import (
	"fmt"
)

func main() {
	fmt.Println("Building docker compose overview...")
	composeFiles := getWatchedComposeFiles()
	var numFiles = len(composeFiles)
	var numContainers = 0
	for _, containers := range composeFiles {
		numContainers += len(containers)
	}
	fmt.Printf("Found %d compose files with %d watched containers.\n", numFiles, numContainers)
	fmt.Println("Trying to update containers...")
	for _, containers := range composeFiles {
		for _, container := range containers {
			fmt.Printf("Pulling image %s ... ", container.image.id)
			var res = pullImage(container.image.id)
			if res {
				fmt.Println("OK")
			} else {
				fmt.Println("Failed")
			}
		}
	}
	fmt.Println("Updating docker compose overview...")
	composeFiles = getWatchedComposeFiles()
	for composeFile, containers := range composeFiles {
		var needsRestart = false
		for _, container := range containers {
			needsRestart = needsRestart || (container.image.hash != getImageHash(container.image.id))
		}
		if needsRestart {
			fmt.Printf("Restarting %s ... ", composeFile)
			downDockerCompose(composeFile)
			upDockerCompose(composeFile)
			fmt.Println("OK")
		} else {
			fmt.Printf("Skipping %s\n", composeFile)
		}
	}
}
