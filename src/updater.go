package main

import (
	"log"
	"os/exec"
)

type Updater struct{}

func (u *Updater) PerformComposeUpdates() {
	EventBus.OnPerformUpdatesStart()
	composeFiles := u.createComposeFileContainerMapping()
	for _, composeFile := range composeFiles {
		u.processComposeFile(composeFile)
	}
	if GlobalSettings.Cleanup {
		u.CleanUp()
	}
	EventBus.OnPerformUpdatesComplete()
}

func (u *Updater) CleanUp() bool {
	EventBus.OnImagePruneStart()
	err := exec.Command("docker", "image", "prune", "-a", "-f").Run()
	if err != nil {
		return false
	}
	return true
}

func (u *Updater) processComposeFile(composeFile *ComposeFile) {
	EventBus.OnProcessComposeFileStart(composeFile)
	compositionRestart := false
	for _, service := range composeFile.Services {
		compositionRestart = u.processService(service) || compositionRestart
	}
	if compositionRestart {
		if GlobalSettings.Dry {
			EventBus.OnSkipRestartComposeFileDryMode(composeFile)
		} else {
			EventBus.OnRestartComposeFile(composeFile)
			composeFile.Down()
			composeFile.Up()
			EventBus.OnRestartComposeFileComplete(composeFile)
		}
	} else {
		EventBus.OnSkipRestartComposeFileNoUpdates(composeFile)
	}
}

func (u *Updater) processService(service *ComposeService) bool {
	EventBus.OnProcessServiceStart(service)
	if !service.IsWatched() {
		return false
	}
	if !service.RequiresBuild() {
		service.Pull()
	} else if GlobalSettings.Build {
		service.Build()
	}
	if service.Instance.Image.ExistsNewerImageHash() {
		if service.RequiresBuild() {
			EventBus.OnServiceNewImageBuilt(service)
		} else {
			EventBus.OnServiceNewImagePulled(service)
		}
		return true
	}
	return false
}

func (u *Updater) createComposeFileContainerMapping() []*ComposeFile {
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
