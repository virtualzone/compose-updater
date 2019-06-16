package main

import (
	"log"
	"os/exec"
	"strings"
)

// DockerImage is a docker image
type DockerImage struct {
	id   string
	hash string
}

// DockerContainer is a docker container
type DockerContainer struct {
	id          string
	name        string
	composeFile string
	image       DockerImage
}

func getWatchedComposeFiles() map[string][]DockerContainer {
	files := make(map[string][]DockerContainer)
	containers := getWatchedRunningContainers()
	for _, container := range containers {
		files[container.composeFile] = append(files[container.composeFile], container)
	}
	return files
}

func getWatchedRunningContainers() []DockerContainer {
	containers := []DockerContainer{}
	containerIDs := getWatchedRunningContainerIDs()
	for _, containerID := range containerIDs {
		containers = append(containers, getRunningContainerDetails(containerID))
	}
	return containers
}

func getWatchedRunningContainerIDs() []string {
	containerIDs := []string{}
	out, err := exec.Command("docker", "ps", "-a", "-q", "--filter", "label=docker-compose-watcher.watch=1").Output()
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(out), "\n")
	for _, containerID := range lines {
		if strings.TrimSpace(containerID) != "" {
			containerIDs = append(containerIDs, containerID)
		}
	}
	return containerIDs
}

func getRunningContainerDetails(id string) DockerContainer {
	return DockerContainer{
		id:          id,
		name:        getRunningContainerImageName(id),
		composeFile: getRunningContainerComposeFile(id),
		image: DockerImage{
			id:   getRunningContainerImageID(id),
			hash: getRunningContainerImageHash(id),
		},
	}
}

func getImageDetails(id string) DockerImage {
	return DockerImage{
		id:   id,
		hash: getImageHash(id),
	}
}

func getImageHash(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "image", "--format", "{{.Id}}", id).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerImageHash(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{.Image}}", id).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerImageID(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{.Config.Image}}", id).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerImageName(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{.Name}}", id).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerComposeFile(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{index .Config.Labels \"docker-compose-watcher.file\"}}", id).Output()
	if err != nil {
		log.Fatal(err)
	}
	fileName := strings.TrimSpace(string(out))
	if fileName == "" {
		out, err = exec.Command("docker", "inspect", "--type", "container", "--format", "{{index .Config.Labels \"docker-compose-watcher.dir\"}}", id).Output()
		if err != nil {
			log.Fatal(err)
		}
		fileName = strings.TrimSpace(string(out))
		if strings.Index(fileName, "/") != len(fileName)-1 {
			fileName += "/"
		}
		fileName += "docker-compose.yml"
	}
	return fileName
}

func pullImage(name string) bool {
	err := exec.Command("docker", "pull", name).Run()
	if err != nil {
		return false
	}
	return true
}

func downDockerCompose(composeFile string) bool {
	err := exec.Command("docker-compose", "-f", composeFile, "--remove-orphans", "down").Run()
	if err != nil {
		return false
	}
	return true
}

func upDockerCompose(composeFile string) bool {
	err := exec.Command("docker-compose", "-f", composeFile, "up", "-d").Run()
	if err != nil {
		return false
	}
	return true
}
