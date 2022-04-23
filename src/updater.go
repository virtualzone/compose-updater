package main

import (
	"log"
	"os/exec"
)

func PerformComposeUpdates() {
	log.Println("Gathering details about running containers...")
	composeFiles := createComposeFileContainerMapping()
	for _, composeFile := range composeFiles {
		compositionRestart := false
		log.Printf("Checking for updates of services in %s...\n", composeFile.YamlFilePath)
		for _, service := range composeFile.Services {
			if service.Instance == nil {
				continue
			}
			requiresBuild := len(service.BuildInfo) > 0
			log.Printf("Processing service %s (requires build: %t)...\n", service.Name, requiresBuild)
			if !requiresBuild {
				service.Pull()
			} else if GlobalSettings.Build {
				service.Build()
			}
			newImage := CreateDockerImageInstance(service.Instance.Image.ID)
			if service.Instance.Image.Hash != newImage.Hash {
				if requiresBuild {
					log.Printf("Built new image for service %s\n", service.Name)
				} else {
					log.Printf("Pulled new image %s for service %s\n", service.ImageName, service.Name)
				}
				/*
					Not working with Docker Compose V2 yet
					if !GlobalSettings.CompleteStop {
						log.Printf("Restarting service %s in %s...\n", service.Name, composeFile.YamlFilePath)
						service.Restart()
					} else {
						compositionRestart = true
					}
				*/
				compositionRestart = true
			}
		}
		if compositionRestart {
			if GlobalSettings.Dry {
				log.Printf("Dry-Mode enabled, not restarting services in %s\n", composeFile.YamlFilePath)
			} else {
				log.Printf("Restarting services in %s...\n", composeFile.YamlFilePath)
				composeFile.Down()
				composeFile.Up()
				log.Printf("Restarted services in %s\n", composeFile.YamlFilePath)
			}
		} else {
			log.Printf("No need to restart services in %s\n", composeFile.YamlFilePath)
		}
	}
}

func CleanUp() bool {
	err := exec.Command("docker", "image", "prune", "-a", "-f").Run()
	if err != nil {
		return false
	}
	return true
}

func createComposeFileContainerMapping() []*ComposeFile {
	containers := GetWatchedRunningContainers()
	cache := make(map[string]*ComposeFile)
	for _, container := range containers {
		var composeFile *ComposeFile
		if curComposeFile, ok := cache[container.ComposeFile]; ok {
			composeFile = curComposeFile
		} else {
			var err error
			composeFile, err = ParseComposeYaml(container.ComposeFile)
			if err != nil {
				log.Fatalf("Could not parse compose YAML: %s\n", err)
			}
			cache[container.ComposeFile] = composeFile
		}
		composeFile.AttachRuntimeInfo()
		for _, service := range composeFile.Services {
			if service.ContainerID == container.ID {
				service.Instance = container
			}
		}
	}
	res := make([]*ComposeFile, 0)
	for _, composeFile := range cache {
		res = append(res, composeFile)
	}
	return res
}
