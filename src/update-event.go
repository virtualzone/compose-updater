package main

import "log"

type UpdateEvent struct{}

var EventBus *UpdateEvent = nil

func CreateEventBus() {
	EventBus = &UpdateEvent{}
}

func (e *UpdateEvent) OnPerformUpdatesStart() {
	log.Println("Gathering details about running containers...")
}

func (e *UpdateEvent) OnPerformUpdatesComplete() {
	log.Println("Done.")
}

func (e *UpdateEvent) OnProcessComposeFileStart(composeFile *ComposeFile) {
	log.Printf("Checking for updates of services in %s...\n", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnSkipRestartComposeFileDryMode(composeFile *ComposeFile) {
	log.Printf("Dry-Mode enabled, not restarting services in %s\n", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnSkipRestartComposeFileNoUpdates(composeFile *ComposeFile) {
	log.Printf("No need to restart services in %s\n", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnRestartComposeFile(composeFile *ComposeFile) {
	log.Printf("Restarting services in %s...\n", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnRestartComposeFileComplete(composeFile *ComposeFile) {
	log.Printf("Restarted services in %s\n", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnProcessServiceStart(service *ComposeService) {
	log.Printf("Processing service %s (requires build: %t, watched: %t)...\n", service.Name, service.RequiresBuild(), service.IsWatched())
}

func (e *UpdateEvent) OnServiceNewImageBuilt(service *ComposeService) {
	log.Printf("Built new image for service %s\n", service.Name)
}

func (e *UpdateEvent) OnServiceNewImagePulled(service *ComposeService) {
	log.Printf("Pulled new image %s for service %s\n", service.ImageName, service.Name)
}

func (e *UpdateEvent) OnImagePruneStart() {
	log.Println("Removing unused images...")
}
